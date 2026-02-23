package targetview

import (
	"errors"
	"fmt"
	"net"

	"github.com/backendsystems/nibble/internal/ports"
	"github.com/backendsystems/nibble/internal/scanner/shared"
	mainview "github.com/backendsystems/nibble/internal/tui/views/main"

	"github.com/charmbracelet/huh"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	errInvalidCIDR    = errors.New("invalid CIDR (e.g. 10.0.0.0/24)")
	errSubnetTooLarge = errors.New("subnet too large (min /16)")
)

type Result struct {
	Model     Model
	Cmd       tea.Cmd
	Quit      bool
	StartScan bool
	Selection mainview.ScanSelection
	SavePorts bool // Signal to save port config
}

// Init returns the initialization command for the form
func (m Model) Init() tea.Cmd {
	if m.Form == nil {
		return nil
	}
	return m.Form.Init()
}

// Update handles tea.Msg and delegates to the form
func (m Model) Update(msg tea.Msg) (Result, tea.Cmd) {
	result := Result{Model: m}

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
					result.Model = m
					// Recreate the form with the new IP
					formModel := NewModel(m.NetworkScan, m.IPInput, m.CIDRInput, m.PortPack, m.CustomPorts, m.Interfaces)
					m.Form = formModel.Form
					result.Model = m
					return result, m.Form.Init()
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
	result.Model = m
	result.Cmd = cmd

	// Check if form is completed
	if m.Form.State == huh.StateCompleted {
		selection, err := validateAndBuild(m)
		if err != nil {
			m.ErrorMsg = err.Error()
			result.Model = m
			return result, nil
		}
		m.ErrorMsg = ""
		result.Model = m
		result.StartScan = true
		result.Selection = selection
		result.SavePorts = true
	}

	// Check if form is aborted
	if m.Form.State == huh.StateAborted {
		result.Quit = true
	}

	return result, cmd
}

func validateAndBuild(m Model) (mainview.ScanSelection, error) {
	// Validate IP
	ipInput := m.IPInput
	if ipInput == "" {
		return mainview.ScanSelection{}, errors.New("IP address required")
	}

	// Validate CIDR
	cidrStr := m.CIDRInput
	if cidrStr == "" {
		cidrStr = "32" // Default to single host
	}
	cidrVal := 0
	_, err := fmt.Sscanf(cidrStr, "%d", &cidrVal)
	if err != nil || cidrVal < 16 || cidrVal > 32 {
		return mainview.ScanSelection{}, errors.New("CIDR must be 16-32")
	}

	// Build CIDR notation
	cidrNotation := ipInput + "/" + cidrStr

	// Parse CIDR
	_, ipnet, err := net.ParseCIDR(cidrNotation)
	if err != nil {
		return mainview.ScanSelection{}, errors.New("invalid IP address")
	}

	// Compute total hosts
	totalHosts := shared.TotalScanHosts(ipnet)

	// Build selection
	selection := mainview.ScanSelection{
		TargetAddr: cidrNotation,
		TotalHosts: totalHosts,
	}

	return selection, nil
}

// SaveTargetPortsConfig saves the target ports configuration to disk
func SaveTargetPortsConfig(portPack, customPorts string) error {
	cfg := ports.Config{
		Mode:   portPack,
		Custom: customPorts,
	}
	return ports.SaveConfig("target", cfg)
}
