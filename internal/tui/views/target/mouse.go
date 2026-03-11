package targetview

import (
	"github.com/backendsystems/nibble/internal/tui/views/common"
	tea "github.com/charmbracelet/bubbletea"
)

// fieldHeights returns the number of lines each field occupies.
// IP: title + input + desc = 3, CIDR: same = 3, PortMode: title + 3 options = 4.
var fieldHeights = [fieldCount]int{3, 3, 4}

func (m *Model) HandleMouse(msg tea.MouseMsg, maxWidth int) (Result, tea.Cmd) {
	helpLineY := m.HelpLineY
	helpLayout := common.BuildHelpLineLayout(targetHelpItems, targetHelpPrefix, maxWidth)
	helpLineEndY := helpLineY + helpLayout.LineCount - 1

	if helpLineY > 0 && msg.Y >= helpLineY && msg.Y <= helpLineEndY {
		m.HoveredHelpItem = common.GetHelpItemAt(helpLayout, msg.X, msg.Y-helpLineY)
	} else {
		m.HoveredHelpItem = -1
	}

	if msg.Button != tea.MouseButtonLeft || msg.Action != tea.MouseActionRelease {
		return Result{}, nil
	}
	if m.ShowHelp {
		m.ShowHelp = false
		return Result{}, nil
	}

	// Helpline clicks
	if helpLineY > 0 && msg.Y >= helpLineY && msg.Y <= helpLineEndY {
		itemIndex := common.GetHelpItemAt(helpLayout, msg.X, msg.Y-helpLineY)
		if itemIndex >= 0 {
			switch helpLayout.Items[itemIndex].Action {
			case targetActionSubmit:
				if m.InCustomPortInput {
					return m.updateCustomPortInput(tea.KeyMsg{Type: tea.KeyEnter})
				}
				result := Result{}
				return m.submitForm(result)
			case targetActionHelp:
				m.ShowHelp = true
			case targetActionQuit:
				if m.InCustomPortInput {
					m.InCustomPortInput = false
					cmd := m.focusField(fieldPortMode)
					return Result{Cmd: cmd}, cmd
				}
				return Result{Quit: true}, nil
			}
		}
		return Result{}, nil
	}

	// Stage 2 (custom port input): no field clicks
	if m.InCustomPortInput {
		return Result{}, nil
	}

	// Stage 1: field clicks
	for field := 0; field < fieldCount; field++ {
		startY := m.FieldY[field]
		endY := startY + fieldHeights[field] - 1
		if msg.Y < startY || msg.Y > endY {
			continue
		}
		// Clicked inside this field
		if field == fieldPortMode {
			// Which port option row was clicked?
			// Row 0 = title line, rows 1-3 = options
			relY := msg.Y - startY
			if relY >= 1 && relY <= len(portModeOptions) {
				optIndex := relY - 1
				if optIndex == m.PortModeIndex {
					// Second click on already-selected option: submit
					cmd := m.focusField(fieldPortMode)
					result := Result{}
					result, cmd = m.submitForm(result)
					return result, cmd
				}
				m.PortModeIndex = optIndex
				m.PortPack = portModeOptions[optIndex].Value
			}
		}
		cmd := m.focusField(field)
		return Result{Cmd: cmd}, cmd
	}

	return Result{}, nil
}
