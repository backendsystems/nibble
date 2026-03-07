package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func scanViewWidth(windowW int) int {
	maxWidth := 72
	if windowW > 8 {
		maxWidth = windowW - 4
	}
	return maxWidth
}

func enterAltScreenCmd() tea.Cmd {
	return func() tea.Msg { return tea.EnterAltScreen() }
}

func exitAltScreenCmd() tea.Cmd {
	return func() tea.Msg { return tea.ExitAltScreen() }
}
