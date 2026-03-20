package historyview

import (
	tea "charm.land/bubbletea/v2"
	"github.com/backendsystems/nibble/internal/tui/views/common"
	"github.com/backendsystems/nibble/internal/tui/views/history/delete"
)

// HandleMouse processes mouse events for the history list view.
// Scroll wheel scrolls the list; clicking a row selects or toggles it.
func (m Model) HandleMouse(msg tea.Msg, maxWidth int) UpdateResult {
	result := UpdateResult{Model: m}

	mouse, ok := mouseXY(msg)
	if !ok {
		return result
	}

	helpLineY := m.HelpLineY
	helpLayout := common.BuildHelpLineLayout(historyHelpItems, historyHelpPrefix, maxWidth)
	helpLineEndY := helpLineY + helpLayout.LineCount - 1

	// Update hover state for all mouse events
	if helpLineY > 0 && mouse.Y >= helpLineY && mouse.Y <= helpLineEndY {
		result.Model.HoveredHelpItem = common.GetHelpItemAt(helpLayout, mouse.X, mouse.Y-helpLineY)
	} else {
		result.Model.HoveredHelpItem = -1
	}

	switch msg.(type) {
	case tea.MouseWheelMsg:
		if mouse.Button == tea.MouseWheelUp {
			if result.Model.Cursor > 0 {
				result.Model.Cursor--
				result.Model = updateViewportContent(result.Model)
				saveViewState(result.Model.FlatList, result.Model.Cursor)
			}
			return result
		}
		if mouse.Button == tea.MouseWheelDown {
			if result.Model.Cursor < len(result.Model.FlatList)-1 {
				result.Model.Cursor++
				result.Model = updateViewportContent(result.Model)
				saveViewState(result.Model.FlatList, result.Model.Cursor)
			}
			return result
		}
	}

	if _, ok := msg.(tea.MouseReleaseMsg); !ok || mouse.Button != tea.MouseLeft {
		return result
	}
	if m.DeleteDialog != nil {
		handled, action := result.Model.DeleteDialog.HandleMouseClick(mouse.X, mouse.Y, m.WindowW, m.WindowH)
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
		result.Model.ShowHelp = false
		return result
	}

	// Check if clicking on helpline item
	if helpLineY > 0 && mouse.Y >= helpLineY && mouse.Y <= helpLineEndY {
		itemIndex := common.GetHelpItemAt(helpLayout, mouse.X, mouse.Y-helpLineY)
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

	titleRows := m.HelpLineY - m.Viewport.Height() - 1
	if titleRows < 2 {
		titleRows = 2
	}
	contentY := mouse.Y - titleRows
	if contentY < 0 {
		return result
	}

	index := m.Viewport.YOffset() + contentY
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

// mouseXY extracts Mouse data from any mouse message type.
func mouseXY(msg tea.Msg) (tea.Mouse, bool) {
	if mm, ok := msg.(tea.MouseMsg); ok {
		return mm.Mouse(), true
	}
	return tea.Mouse{}, false
}
