package targetview

import (
	"errors"
	"fmt"
	"net"

	"github.com/backendsystems/nibble/internal/ports"

	"github.com/charmbracelet/huh"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	errInvalidCIDR    = errors.New("invalid CIDR (e.g. 10.0.0.0/24)")
	errSubnetTooLarge = errors.New("subnet too large (min /16)")
)

type Result struct {
	Cmd        tea.Cmd
	Quit       bool
	StartScan  bool   // Signal to start scan with saved config
	TargetAddr string // Target address in CIDR notation
	TotalHosts int    // Number of hosts to scan
	Ports      []int  // Resolved port list to scan
}

// Init returns the initialization command for the form
func (m *Model) Init() tea.Cmd {
	if m.Form == nil {
		m.initializeForm()
	}
	return m.Form.Init()
}

// Update handles tea.Msg and delegates to the form
// The model is updated in place to preserve form bindings
func (m *Model) Update(msg tea.Msg) (Result, tea.Cmd) {
	result := Result{}

	// Handle special keys
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "q", "esc":
			result.Quit = true
			return result, nil
		case "up", "down", "k", "j", "w", "s":
			// Cycle through interface IPs only when IP field is focused
			if m.Form != nil {
				focused := m.Form.GetFocusedField()
				if focused != nil && focused.GetKey() == "ip" {
					// Determine direction: up/k/w = forward, down/j/s = backward
					forward := keyMsg.String() == "up" || keyMsg.String() == "k" || keyMsg.String() == "w"
					m.CycleInterfaceIP(forward)
					// Don't recreate the form - the form bindings are already pointing to m.IPInput
					// which we just updated via CycleInterfaceIP
					return result, nil
				}
			}
		default:
			// Block invalid characters for IP field if it's focused
			if m.Form != nil {
				focused := m.Form.GetFocusedField()
				if focused != nil && focused.GetKey() == "ip" {
					// Only allow printable characters that are digits or dots
					if len(keyMsg.String()) == 1 {
						ch := keyMsg.String()[0]
						if !((ch >= '0' && ch <= '9') || ch == '.') {
							// Invalid character for IP field, ignore it
							return result, nil
						}
					}
				}
			}
		}
	}

	if m.Form == nil {
		return result, nil
	}

	// Delegate update to the form
	formModel, cmd := m.Form.Update(msg)
	if f, ok := formModel.(*huh.Form); ok {
		m.Form = f
	}
	result.Cmd = cmd

	// Check if form is completed
	if m.Form.State == huh.StateCompleted {
		// Extract values directly from form fields instead of model fields
		// The .Value() bindings don't update model fields properly
		ipInput := m.Form.GetString("ip")
		cidrInput := m.Form.GetString("cidr")
		portPack := m.Form.GetString("port_mode")
		customPorts := m.Form.GetString("custom_ports")

		// Extract values from form and create scan config
		targetAddr, totalHosts, resolvedPorts, err := buildScanConfig(ipInput, cidrInput, portPack, customPorts)
		if err != nil {
			m.ErrorMsg = err.Error()
			return result, nil
		}

		m.ErrorMsg = ""
		result.StartScan = true
		result.TargetAddr = targetAddr
		result.TotalHosts = totalHosts
		result.Ports = resolvedPorts
	}

	// Check if form is aborted
	if m.Form.State == huh.StateAborted {
		result.Quit = true
	}

	return result, cmd
}

// buildScanConfig extracts form values and builds a complete scan configuration
// Returns: targetAddr (CIDR notation), totalHosts, resolvedPorts, error
func buildScanConfig(ipInput, cidrInput, portPack, customPorts string) (string, int, []int, error) {
	// Validate IP
	if ipInput == "" {
		return "", 0, nil, errors.New("IP address required")
	}

	// Validate CIDR
	cidrStr := cidrInput
	if cidrStr == "" {
		cidrStr = "32" // Default to single host
	}
	cidrVal := 0
	_, err := fmt.Sscanf(cidrStr, "%d", &cidrVal)
	if err != nil || cidrVal < 16 || cidrVal > 32 {
		return "", 0, nil, errors.New("CIDR must be 16-32")
	}

	// Build CIDR notation and validate
	cidrNotation := ipInput + "/" + cidrStr
	_, ipnet, err := net.ParseCIDR(cidrNotation)
	if err != nil {
		return "", 0, nil, errors.New("invalid IP address")
	}

	// Calculate total hosts
	totalHosts := hostCount(ipnet)

	// Resolve ports based on mode
	var resolvedPorts []int
	switch portPack {
	case "all":
		resolvedPorts, err = ports.ParseList("1-65535")
	case "custom":
		if customPorts == "" {
			resolvedPorts = []int{} // Empty list = host-only scan
		} else {
			resolvedPorts, err = ports.ParseList(customPorts)
		}
	default: // "default" mode
		resolvedPorts = ports.DefaultPorts()
	}

	if err != nil {
		return "", 0, nil, err
	}

	return cidrNotation, totalHosts, resolvedPorts, nil
}

// hostCount calculates the number of hosts in an IP network
func hostCount(ipnet *net.IPNet) int {
	ones, bits := ipnet.Mask.Size()
	hostBits := bits - ones

	// /32 has one host
	if hostBits <= 0 {
		return 1
	}

	totalHosts := 1 << uint(hostBits)

	// /31 keeps both addresses as usable hosts (RFC 3021)
	if hostBits == 1 {
		return totalHosts // 2 hosts
	}

	// Larger subnets skip network and broadcast addresses
	return totalHosts - 2
}
