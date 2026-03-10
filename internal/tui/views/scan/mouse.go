package scanview

import (
	"github.com/backendsystems/nibble/internal/tui/views/common"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) HandleMouse(msg tea.MouseMsg) Result {
	result := Result{Model: m}

	if common.IsRightClick(msg) {
		return handleKeyMsg(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	}

	return result
}
