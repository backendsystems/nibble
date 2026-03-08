package targetview

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
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
			// Accept any key to close help overlay
			m.ShowHelp = false
			return result, nil
		}

		// Check for backspace on IP field - prevent deletion before first dot
		if keyMsg.Type == tea.KeyBackspace {
			if m.Form != nil {
				focused := m.Form.GetFocusedField()
				if focused != nil && focused.GetKey() == "ip" {
					// Get the actual input field to check its current value
					if inputField, ok := focused.(*huh.Input); ok {
						if currentValue, ok := inputField.GetValue().(string); ok {
							// Find the position of the first dot
							firstDotPos := strings.Index(currentValue, ".")
							// If there's a dot and cursor is at or before it, block backspace
							if firstDotPos >= 0 && len(currentValue) <= firstDotPos+1 {
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
			m.PortInput.Value = m.CustomPorts // seed from saved value
			m.PortInput.Cursor = len(m.CustomPorts)
			m.PortInput.Ready = false // force re-init
			var prepCmd tea.Cmd
			m.PortInput, prepCmd = m.PortInput.Prepare(true)
			m.Form.State = huh.StateNormal // reset so form doesn't show as complete
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

