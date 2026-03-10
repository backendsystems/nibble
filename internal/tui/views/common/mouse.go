package common

import tea "github.com/charmbracelet/bubbletea"

func IsRightClick(msg tea.MouseMsg) bool {
	return msg.Button == tea.MouseButtonRight && msg.Action == tea.MouseActionPress
}
