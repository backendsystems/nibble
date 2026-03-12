package targetview

import (
	"strings"

	"github.com/backendsystems/nibble/internal/ports"
	"github.com/backendsystems/nibble/internal/tui/views/common"

	tea "github.com/charmbracelet/bubbletea"
)

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
		// Accept any key to close help overlay
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
		cmd := m.focusField(fieldPortMode)
		return result, cmd
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
