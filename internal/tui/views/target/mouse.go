package targetview

import (
	tea "charm.land/bubbletea/v2"
	"github.com/backendsystems/nibble/internal/tui/views/common"
)

// fieldHeights returns the number of lines each field occupies.
// IP: title + input + desc = 3, CIDR: same = 3, PortMode: title + 3 options = 4.
var fieldHeights = [fieldCount]int{3, 3, 4}

func (m *Model) HandleMouse(msg tea.Msg, maxWidth int) (Result, tea.Cmd) {
	mouse, ok := common.MouseXY(msg)
	if !ok {
		return Result{}, nil
	}

	hlr := common.HandleHelpLineMouse(mouse, msg, targetHelpItems, targetHelpPrefix, m.HelpLineY, maxWidth)
	m.HoveredHelpItem = hlr.NewHoveredItem

	if _, ok := msg.(tea.MouseReleaseMsg); !ok || mouse.Button != tea.MouseLeft {
		return Result{}, nil
	}
	if m.ShowHelp {
		m.ShowHelp = false
		return Result{}, nil
	}

	if hlr.Consumed {
		switch hlr.ClickedAction {
		case targetActionSubmit:
			if m.InCustomPortInput {
				return m.updateCustomPortInput(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
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
		if mouse.Y < startY || mouse.Y > endY {
			continue
		}
		// Clicked inside this field
		if field == fieldPortMode {
			// Which port option row was clicked?
			// Row 0 = title line, rows 1-3 = options
			relY := mouse.Y - startY
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
