package targetview

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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
	if m.IPTextInput.Value() == "" && m.CIDRTextInput.Value() == "" {
		m.initializeInputs()
	}

	// Auto-transition to custom input if returning from back/esc with custom selected
	if m.PortPack == "custom" && !m.InCustomPortInput {
		m.InCustomPortInput = true
		m.PortInput.Value = m.CustomPorts
		m.PortInput.Cursor = len(m.CustomPorts)
		m.PortInput.Ready = false
		_, cmd := m.PortInput.Prepare(true)
		return cmd
	}

	m.InCustomPortInput = false
	return m.focusField(m.FocusedField)
}

// Update handles tea.Msg and delegates to the custom inputs or port selection
func (m *Model) Update(msg tea.Msg) (Result, tea.Cmd) {
	result := Result{}

	// --- Stage 2: Custom port textinput is active ---
	if m.InCustomPortInput {
		return m.updateCustomPortInput(msg)
	}

	// Forward non-key messages to focused textinput
	if _, ok := msg.(tea.KeyMsg); !ok {
		return m.updateFocusedInput(msg)
	}

	keyMsg := msg.(tea.KeyMsg)

	if m.ShowHelp {
		m.ShowHelp = false
		return result, nil
	}

	switch keyMsg.String() {
	case "q", "esc":
		if m.ErrorMsg != "" {
			m.ErrorMsg = ""
		}
		result.Quit = true
		return result, nil
	case "?":
		m.ShowHelp = true
		return result, nil
	case "tab", "down", "j", "s", "enter":
		if m.FocusedField == fieldPortMode {
			if keyMsg.String() == "enter" {
				return m.submitForm(result)
			}
			// down/j/s moves through port options; tab wraps back to IP
			if keyMsg.String() == "tab" {
				cmd := m.focusField(fieldIP)
				return result, cmd
			}
			if m.PortModeIndex < len(portModeOptions)-1 {
				m.PortModeIndex++
				m.PortPack = portModeOptions[m.PortModeIndex].Value
			}
			return result, nil
		}
		// Move to next field
		cmd := m.focusField((m.FocusedField + 1) % fieldCount)
		return result, cmd
	case "shift+tab", "up", "k", "w":
		if m.FocusedField == fieldPortMode {
			if m.PortModeIndex > 0 {
				m.PortModeIndex--
				m.PortPack = portModeOptions[m.PortModeIndex].Value
			} else {
				// At top of port list, go to previous field
				cmd := m.focusField(fieldCIDR)
				return result, cmd
			}
			return result, nil
		}
		// Move to previous field (wrap from IP to port mode)
		prev := (m.FocusedField - 1 + fieldCount) % fieldCount
		cmd := m.focusField(prev)
		return result, cmd
	case "left":
		if m.FocusedField == fieldIP {
			m.CycleInterfaceIP(false)
			return result, nil
		}
	case "right":
		if m.FocusedField == fieldIP {
			m.CycleInterfaceIP(true)
			return result, nil
		}
	}

	// Delegate to focused textinput with character filtering
	switch m.FocusedField {
	case fieldIP:
		if keyMsg.Type == tea.KeyBackspace {
			// Block backspace if cursor is at or before the first dot
			val := m.IPTextInput.Value()
			firstDotPos := strings.Index(val, ".")
			if firstDotPos >= 0 && len(val) <= firstDotPos+1 {
				return result, nil
			}
		} else if keyMsg.Type == tea.KeyRunes {
			ch := keyMsg.Runes[0]
			if !((ch >= '0' && ch <= '9') || ch == '.') {
				return result, nil
			}
		}
		var cmd tea.Cmd
		m.IPTextInput, cmd = m.IPTextInput.Update(msg)
		m.IPInput = m.IPTextInput.Value()
		result.Cmd = cmd
		return result, cmd
	case fieldCIDR:
		if keyMsg.Type == tea.KeyRunes {
			ch := keyMsg.Runes[0]
			if !(ch >= '0' && ch <= '9') {
				return result, nil
			}
		}
		var cmd tea.Cmd
		m.CIDRTextInput, cmd = m.CIDRTextInput.Update(msg)
		m.CIDRInput = m.CIDRTextInput.Value()
		result.Cmd = cmd
		return result, cmd
	}

	return result, nil
}

// submitForm validates and submits the form
func (m *Model) submitForm(result Result) (Result, tea.Cmd) {
	m.IPInput = m.IPTextInput.Value()
	m.CIDRInput = m.CIDRTextInput.Value()
	m.PortPack = portModeOptions[m.PortModeIndex].Value

	if m.PortPack == "custom" {
		m.InCustomPortInput = true
		m.PortInput.Value = m.CustomPorts
		m.PortInput.Cursor = len(m.CustomPorts)
		m.PortInput.Ready = false
		var prepCmd tea.Cmd
		m.PortInput, prepCmd = m.PortInput.Prepare(true)
		return result, prepCmd
	}

	return m.finalizeScan(result)
}

// updateFocusedInput forwards non-key messages to the focused textinput
func (m *Model) updateFocusedInput(msg tea.Msg) (Result, tea.Cmd) {
	result := Result{}
	switch m.FocusedField {
	case fieldIP:
		var cmd tea.Cmd
		m.IPTextInput, cmd = m.IPTextInput.Update(msg)
		m.IPInput = m.IPTextInput.Value()
		result.Cmd = cmd
		return result, cmd
	case fieldCIDR:
		var cmd tea.Cmd
		m.CIDRTextInput, cmd = m.CIDRTextInput.Update(msg)
		m.CIDRInput = m.CIDRTextInput.Value()
		result.Cmd = cmd
		return result, cmd
	}
	return result, nil
}
