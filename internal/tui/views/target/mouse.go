package targetview

import (
	"github.com/backendsystems/nibble/internal/tui/views/common"
	tea "github.com/charmbracelet/bubbletea"
)

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

	if helpLineY > 0 && msg.Y >= helpLineY && msg.Y <= helpLineEndY {
		itemIndex := common.GetHelpItemAt(helpLayout, msg.X, msg.Y-helpLineY)
		if itemIndex >= 0 {
			switch helpLayout.Items[itemIndex].Action {
			case targetActionHelp:
				m.ShowHelp = true
			case targetActionQuit:
				return Result{Quit: true}, nil
			}
		}
	}

	return Result{}, nil
}
