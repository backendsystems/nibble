package targetview

import (
	"github.com/backendsystems/nibble/internal/tui/views/common"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) HandleMouse(msg tea.MouseMsg) (Result, tea.Cmd) {
	if common.IsRightClick(msg) {
		if m.InCustomPortInput {
			return m.updateCustomPortInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
		}
		return m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	}

	return Result{}, nil
}
