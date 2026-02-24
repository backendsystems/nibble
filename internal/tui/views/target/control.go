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

	// Only auto-transition to custom input if form was already shown
	// (i.e., returning from back/esc, not first open)
	if m.PortPack == "custom" && !m.InCustomPortInput && m.Form.State != huh.StateNormal {
		m.InCustomPortInput = true
		m.PortInput.Value = m.CustomPorts
		m.PortInput.Cursor = len(m.CustomPorts)
		m.PortInput.Ready = false
		_, cmd := m.PortInput.Prepare(true)
		return cmd
	}

	m.InCustomPortInput = false
	return m.Form.Init()
}

// Update handles tea.Msg and delegates to the form or custom port input
// The model is updated in place to preserve form bindings
func (m *Model) Update(msg tea.Msg) (Result, tea.Cmd) {
	result := Result{}

	// --- Stage 2: Custom port textinput is active ---
	if m.InCustomPortInput {
		return m.updateCustomPortInput(msg)
	}

	// --- Stage 1: huh form (ip/cidr/port_mode) ---
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if m.ShowHelp {
			m.ShowHelp = false
			return result, nil
		}

		// Check for backspace on IP field - prevent if it would leave < 4 characters
		if keyMsg.Type == tea.KeyBackspace {
			if m.Form != nil {
				focused := m.Form.GetFocusedField()
				if focused != nil && focused.GetKey() == "ip" {
					// Get the actual input field to check its current value
					if inputField, ok := focused.(*huh.Input); ok {
						if currentValue, ok := inputField.GetValue().(string); ok {
							// Block if backspace would result in fewer than 4 characters
							if len(currentValue) <= 4 {
								return result, nil
							}
						}
					}
				}
			}
		}

		switch keyMsg.String() {
		case "q", "esc":
			// Clear error state if present
			if m.ErrorMsg != "" {
				m.ErrorMsg = ""
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
				// Prevent deletion of first 4 characters in IP field
				if inputField, ok := focused.(*huh.Input); ok {
					if currentValue, ok := inputField.GetValue().(string); ok {
						if len(currentValue) < 5 {
							return result, nil
						}
					}
				}
				m.IPInput = ""
			case "cidr":
				m.CIDRInput = ""
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

	// Check if form is trying to complete
	if m.Form.State == huh.StateCompleted {
		// Extract values directly from form fields
		ipInput := m.Form.GetString("ip")
		cidrInput := m.Form.GetString("cidr")
		portPack := m.Form.GetString("port_mode")

		m.PortPack = portPack

		if portPack == "custom" {
			// Transition to stage 2 (custom port textinput)
			m.InCustomPortInput = true
			m.PortInput.Value = m.CustomPorts  // seed from saved value
			m.PortInput.Cursor = len(m.CustomPorts)
			m.PortInput.Ready = false           // force re-init
			var prepCmd tea.Cmd
			m.PortInput, prepCmd = m.PortInput.Prepare(true)
			m.Form.State = huh.StateNormal      // reset so form doesn't show as complete
			return result, prepCmd
		}

		// Non-custom: complete immediately
		m.IPInput = ipInput
		m.CIDRInput = cidrInput
		return m.finalizeScan(result)
	}

	// Check if form is aborted
	if m.Form.State == huh.StateAborted {
		result.Quit = true
	}

	return result, cmd
}

// updateCustomPortInput handles all messages when InCustomPortInput == true
func (m *Model) updateCustomPortInput(msg tea.Msg) (Result, tea.Cmd) {
	result := Result{}

	// Non-key: blink tick forwarded to textinput
	if _, ok := msg.(tea.KeyMsg); !ok {
		var cmd tea.Cmd
		m.PortInput, cmd = m.PortInput.UpdateNonKey(msg)
		return result, cmd
	}

	keyMsg := msg.(tea.KeyMsg)

	if m.ShowHelp {
		m.ShowHelp = false
		return result, nil
	}

	switch keyMsg.String() {
	case "ctrl+c":
		result.Quit = true
		return result, nil
	case "q", "esc":
		// Go back to stage 1 (port_mode select)
		m.InCustomPortInput = false
		m.initializeForm()
		// Navigate form to port_mode field
		m.Form.NextField() // ip -> cidr
		m.Form.NextField() // cidr -> port_mode
		return result, m.Form.Init()
	case "?":
		m.ShowHelp = true
		return result, nil
	case "delete":
		m.PortInput = m.PortInput.SetValue("")
		return result, nil
	case "enter":
		// Validate and finalize
		normalized, err := ports.NormalizeCustom(strings.TrimSpace(m.PortInput.Value))
		if err != nil {
			m.ErrorMsg = err.Error()
			return result, nil
		}
		m.PortInput = m.PortInput.SetValue(normalized)
		m.CustomPorts = normalized
		m.InCustomPortInput = false
		return m.finalizeScan(result)
	}

	// All other keys: delegate to PortInput
	portAction := common.PortInputActionFromKey(keyMsg.String(), keyMsg.Type == tea.KeyRunes)
	var cmd tea.Cmd
	m.PortInput, cmd = m.PortInput.HandleKey(portAction, keyMsg)
	return result, cmd
}

// finalizeScan validates IP/CIDR/ports and emits a StartScan result
func (m *Model) finalizeScan(result Result) (Result, tea.Cmd) {
	customPorts := ""
	if m.PortPack == "custom" {
		customPorts = m.CustomPorts // already normalized
	}

	targetAddr, totalHosts, resolvedPorts, err := buildScanConfig(m.IPInput, m.CIDRInput, m.PortPack, customPorts)
	if err != nil {
		m.ErrorMsg = err.Error()
		// re-show the form at stage 1
		m.InCustomPortInput = false
		m.initializeForm()
		return result, m.Form.Init()
	}
	if err := ports.SaveConfig("target", ports.Config{Mode: m.PortPack, Custom: m.CustomPorts}); err != nil {
		m.ErrorMsg = err.Error()
		return result, nil
	}

	m.ErrorMsg = ""
	result.StartScan = true
	result.TargetAddr = targetAddr
	result.TotalHosts = totalHosts
	result.Ports = resolvedPorts
	return result, nil
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
