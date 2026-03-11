package historyview

import (
	"github.com/backendsystems/nibble/internal/tui/views/common"
	"github.com/backendsystems/nibble/internal/tui/views/history/delete"
	tea "github.com/charmbracelet/bubbletea"
)

// HandleMouse processes mouse events for the history list view.
// Scroll wheel scrolls the list; clicking a row selects or toggles it.
func (m Model) HandleMouse(msg tea.MouseMsg, maxWidth int) UpdateResult {
	result := UpdateResult{Model: m}

	helpLineY := m.HelpLineY
	helpLayout := common.BuildHelpLineLayout(historyHelpItems, historyHelpPrefix, maxWidth)
	helpLineEndY := helpLineY + helpLayout.LineCount - 1

	// Update hover state for all mouse events
	if helpLineY > 0 && msg.Y >= helpLineY && msg.Y <= helpLineEndY {
		result.Model.HoveredHelpItem = common.GetHelpItemAt(helpLayout, msg.X, msg.Y-helpLineY)
	} else {
		result.Model.HoveredHelpItem = -1
	}

	switch msg.Button {
	case tea.MouseButtonWheelUp:
		if result.Model.Cursor > 0 {
			result.Model.Cursor--
			result.Model = updateViewportContent(result.Model)
			saveViewState(result.Model.FlatList, result.Model.Cursor)
		}
		return result
	case tea.MouseButtonWheelDown:
		if result.Model.Cursor < len(result.Model.FlatList)-1 {
			result.Model.Cursor++
			result.Model = updateViewportContent(result.Model)
			saveViewState(result.Model.FlatList, result.Model.Cursor)
		}
		return result
	}

	if msg.Button != tea.MouseButtonLeft || msg.Action != tea.MouseActionRelease {
		return result
	}
	if m.DeleteDialog != nil {
		handled, action := result.Model.DeleteDialog.HandleMouseClick(msg.X, msg.Y, m.WindowW, m.WindowH)
		if !handled {
			return result
		}
		switch action {
		case delete.MouseActionConfirmYes:
			return handleDeleteDialog(result, ActionConfirmYes)
		case delete.MouseActionConfirmNo:
			return handleDeleteDialog(result, ActionConfirmNo)
		default:
			return result
		}
	}
	if m.ShowHelp {
		return result
	}

	// Check if clicking on helpline item
	if helpLineY > 0 && msg.Y >= helpLineY && msg.Y <= helpLineEndY {
		itemIndex := common.GetHelpItemAt(helpLayout, msg.X, msg.Y-helpLineY)
		if itemIndex >= 0 {
			switch Action(helpLayout.Items[itemIndex].Action) {
			case ActionDelete:
				result = handleListKey(result, ActionDelete)
			case ActionHelp:
				result.Model.ShowHelp = true
			case ActionQuit:
				result.Quit = true
			}
			return result
		}
	}

	titleRows := m.HelpLineY - m.Viewport.Height - 1
	if titleRows < 2 {
		titleRows = 2
	}
	contentY := msg.Y - titleRows
	if contentY < 0 {
		return result
	}

	index := m.Viewport.YOffset + contentY
	if index < 0 || index >= len(m.FlatList) {
		return result
	}

	if index == m.Cursor {
		// Second click on same row: activate (toggle/enter)
		result = handleListKey(result, ActionToggle)
		if !result.Quit && result.Model.Mode == ViewList {
			result.Model = updateViewportContent(result.Model)
		}
		return result
	}
	result.Model.Cursor = index
	result.Model = updateViewportContent(result.Model)
	saveViewState(result.Model.FlatList, result.Model.Cursor)
	return result
}
