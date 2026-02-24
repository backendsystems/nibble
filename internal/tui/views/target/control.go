package targetview

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/backendsystems/nibble/internal/ports"
	"github.com/backendsystems/nibble/internal/tui/views/common"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
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
		if m.ShowHelp {
			m.ShowHelp = false
			return result, nil
		}

		switch keyMsg.String() {
		case "q", "esc":
			// If in custom_ports field, convert to shift+tab to navigate back
			if m.Form != nil {
				focused := m.Form.GetFocusedField()
				if focused != nil && focused.GetKey() == "custom_ports" {
					// Convert to shift+tab and let form handle it
					msg = tea.KeyMsg{Type: tea.KeyShiftTab}
					break
				}
			}
			result.Quit = true
			return result, nil
		case "?":
			m.ShowHelp = true
			return result, nil
		case "delete":
			if m.Form == nil {
				return result, nil
			}
			focused := m.Form.GetFocusedField()
			if focused == nil {
				return result, nil
			}

			switch focused.GetKey() {
			case "ip":
				m.IPInput = ""
			case "cidr":
				m.CIDRInput = ""
			case "custom_ports":
				m.CustomPorts = ""
			default:
				return result, nil
			}

			// Rebuild form so the focused input reflects the cleared value.
			m.initializeForm()
			return result, m.Form.Init()
		case "left", "right":
			// Cycle through interface IPs when IP field is focused
			if m.Form != nil {
				focused := m.Form.GetFocusedField()
				if focused != nil && focused.GetKey() == "ip" {
					forward := keyMsg.String() == "right"
					m.CycleInterfaceIP(forward)
					// Recreate the form with the new IP value
					m.initializeForm()
					return result, m.Form.Init()
				}
			}
			// For other fields, let left/right fall through for normal behavior
		case "up", "k", "w":
			// Navigate form fields upward (like shift+tab)
			if m.Form != nil {
				focused := m.Form.GetFocusedField()
				// Block navigation in custom_ports field
				if focused != nil && focused.GetKey() == "custom_ports" {
					return result, nil
				}
				// Wrap from the first field back to ports selection instead of exiting the form.
				if focused != nil && focused.GetKey() == "ip" {
					m.Form.NextField() // ip -> cidr
					m.Form.NextField() // cidr -> port_mode
					return result, nil
				}
				// For port_mode select: if at first option (index 0), navigate up to CIDR
				if focused != nil && focused.GetKey() == "port_mode" {
					// Type assert to Select to access Hovered method
					if selectField, ok := focused.(*huh.Select[string]); ok {
						hovered, _ := selectField.Hovered()
						// Check if we're at the first option ("default")
						if hovered == "default" {
							// At first option, navigate to previous field
							m.Form.PrevField()
							return result, nil
						}
					}
					// Not at first option, convert k/w to up and let select handle it
					if keyMsg.String() == "k" || keyMsg.String() == "w" {
						msg = tea.KeyMsg{Type: tea.KeyUp}
					}
					break
				}
				m.Form.PrevField()
				return result, nil
			}
		case "down", "j", "s":
			// Navigate form fields downward (like tab)
			if m.Form != nil {
				focused := m.Form.GetFocusedField()
				// Block navigation in custom_ports field
				if focused != nil && focused.GetKey() == "custom_ports" {
					return result, nil
				}
				// For port_mode select: convert j/s to down and let it handle navigation
				if focused != nil && focused.GetKey() == "port_mode" {
					if keyMsg.String() == "j" || keyMsg.String() == "s" {
						msg = tea.KeyMsg{Type: tea.KeyDown}
					}
					break
				}
				m.Form.NextField()
				return result, nil
			}
		default:
			// Block invalid characters based on focused field
			if m.Form != nil {
				focused := m.Form.GetFocusedField()
				if focused != nil {
					key := focused.GetKey()
					if len(keyMsg.String()) == 1 {
						ch := keyMsg.String()[0]

						// IP field: only digits and dots
						if key == "ip" {
							if !((ch >= '0' && ch <= '9') || ch == '.') {
								return result, nil
							}
						}

						// CIDR field: only digits
						if key == "cidr" {
							if !(ch >= '0' && ch <= '9') {
								return result, nil
							}
						}

						// Custom ports field: block navigation keys w/s/k/j
						if key == "custom_ports" {
							if ch == 'w' || ch == 's' || ch == 'k' || ch == 'j' {
								return result, nil
							}
						}
					}

					// Custom ports field: use portinput validation
					if key == "custom_ports" && keyMsg.Type == tea.KeyRunes {
						// Get current value and cursor position from the input field
						currentValue := m.Form.GetString("custom_ports")

						// Filter runes through portinput
						filtered := make([]rune, 0, len(keyMsg.Runes))
						for _, r := range keyMsg.Runes {
							if r >= 32 {
								filtered = append(filtered, r)
							}
						}

						if len(filtered) > 0 {
							// Insert runes at the end (huh doesn't expose cursor position)
							newValue, _ := common.InsertRunes(currentValue, len(currentValue), filtered)

							// If the value changed, it means valid characters were added
							// Otherwise, block the input
							if newValue == currentValue {
								return result, nil
							}
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
		savedCustomPorts := customPorts
		if portPack == "custom" {
			normalized, err := ports.NormalizeCustom(strings.TrimSpace(customPorts))
			if err != nil {
				m.ErrorMsg = err.Error()
				return result, nil
			}
			savedCustomPorts = normalized
			customPorts = normalized
		}

		// Extract values from form and create scan config
		targetAddr, totalHosts, resolvedPorts, err := buildScanConfig(ipInput, cidrInput, portPack, customPorts)
		if err != nil {
			m.ErrorMsg = err.Error()
			return result, nil
		}
		if err := ports.SaveConfig("target", ports.Config{Mode: portPack, Custom: savedCustomPorts}); err != nil {
			m.ErrorMsg = err.Error()
			return result, nil
		}

		// Keep target model state in sync so reopening the view shows saved values.
		m.IPInput = ipInput
		m.CIDRInput = cidrInput
		m.PortPack = portPack
		m.CustomPorts = savedCustomPorts

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
