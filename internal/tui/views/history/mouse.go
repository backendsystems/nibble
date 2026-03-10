package historyview

import (
	tea "github.com/charmbracelet/bubbletea"
)

const (
	// "Scan History\n\n" = title line + blank line before the tree viewport
	historyListTitleRows = 2
)

// HandleMouse processes a mouse event for the history list view.
// Clicking a tree row selects it; clicking the already-selected row toggles it.
func (m Model) HandleMouse(msg tea.MouseMsg) UpdateResult {
	result := UpdateResult{Model: m}
	if msg.Button != tea.MouseButtonLeft || msg.Action != tea.MouseActionRelease {
		return result
	}
	if m.ShowHelp || m.DeleteDialog != nil {
		return result
	}

	contentY := msg.Y - historyListTitleRows
	if contentY < 0 {
		return result
	}

	index := m.Viewport.YOffset + contentY
	if index < 0 || index >= len(m.FlatList) {
		return result
	}

	if index == m.Cursor {
		// Second click on same row: activate (toggle/enter)
		return handleListKey(result, ActionToggle)
	}
	result.Model.Cursor = index
	saveViewState(result.Model.FlatList, result.Model.Cursor)
	return result
}
