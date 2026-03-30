package targetview

import (
	tea "charm.land/bubbletea/v2"
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

	// Ensure we start at the form view, not custom port input
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
	if _, ok := msg.(tea.KeyPressMsg); !ok {
		return m.updateFocusedInput(msg)
	}

	keyMsg := msg.(tea.KeyPressMsg)

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
			// down/j/s moves through port options; tab wraps back to first field
			if keyMsg.String() == "tab" {
				cmd := m.focusField((fieldPortMode + 1) % fieldCount)
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
		// Move to previous field
		prev := (m.FocusedField - 1 + fieldCount) % fieldCount
		cmd := m.focusField(prev)
		return result, cmd
	case "h", "a":
		msg = tea.KeyPressMsg(tea.Key{Code: tea.KeyLeft})
		keyMsg = msg.(tea.KeyPressMsg)
	case "l", "d":
		msg = tea.KeyPressMsg(tea.Key{Code: tea.KeyRight})
		keyMsg = msg.(tea.KeyPressMsg)
	}

	switch keyMsg.String() {
	case "left":
		if m.FocusedField == fieldInterface {
			m.CycleInterfaceIP(false)
			return result, nil
		}
	case "right":
		if m.FocusedField == fieldInterface {
			m.CycleInterfaceIP(true)
			return result, nil
		}
	}

	// Delegate to focused textinput with character filtering
	switch m.FocusedField {
	case fieldIP:
		if keyMsg.Text != "" {
			ch := []rune(keyMsg.Text)[0]
			if !((ch >= '0' && ch <= '9') || ch == '.') {
				return result, nil
			}
		}
		var cmd tea.Cmd
		m.IPTextInput, cmd = m.IPTextInput.Update(msg)
		m.IPInput = m.IPTextInput.Value()
		m.IPIsCustom = !m.ipMatchesInterface(m.IPInput)
		if !m.IPIsCustom {
			// Sync IPIndex to the interface whose prefix matches
			prefix := ipPrefix(m.IPInput)
			for i, info := range m.InterfaceInfos {
				if ipPrefix(info.IP) == prefix {
					m.IPIndex = i
					break
				}
			}
		}
		result.Cmd = cmd
		return result, cmd
	case fieldCIDR:
		if keyMsg.Text != "" {
			ch := []rune(keyMsg.Text)[0]
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
