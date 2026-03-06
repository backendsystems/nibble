package historydetailview

import (
	"github.com/backendsystems/nibble/internal/history"
	deletepkg "github.com/backendsystems/nibble/internal/tui/views/history/delete"
	tea "github.com/charmbracelet/bubbletea"
)

type Action int

const (
	ActionNone Action = iota
	ActionQuit
	ActionMoveUp
	ActionMoveDown
	ActionScanAllPorts
	ActionHelp
	ActionDelete
)

type UpdateResult struct {
	Model           Model
	Quit            bool
	ScanAllPorts    bool
	SelectedHostIP  string
	ScanHistoryPath string
	Deleted         bool
}

func HandleKey(key string) Action {
	switch key {
	case "q", "esc":
		return ActionQuit
	case "up", "w", "k":
		return ActionMoveUp
	case "down", "s", "j":
		return ActionMoveDown
	case "enter":
		return ActionScanAllPorts
	case "delete":
		return ActionDelete
	case "?":
		return ActionHelp
	default:
		return ActionNone
	}
}

func (m Model) Update(msg tea.Msg) UpdateResult {
	result := UpdateResult{Model: m}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return handleKeyMsg(m, msg)
	}

	// Update viewport for scrolling
	var cmd tea.Cmd
	result.Model.Viewport, cmd = m.Viewport.Update(msg)
	_ = cmd

	return result
}

func handleKeyMsg(m Model, key tea.KeyMsg) UpdateResult {
	result := UpdateResult{Model: m}

	// Handle delete dialog in detail view
	if m.DeleteDialog != nil {
		switch key.String() {
		case "left", "a", "h", "right", "d", "l":
			// Toggle between Delete and Cancel
			result.Model.DeleteDialog.Toggle()
			return result
		case "enter":
			// User pressed Enter - execute the selected action
			if result.Model.DeleteDialog.IsDeleteSelected() {
				// Delete was selected
				performDeleteSync(m.NodePath)
				result.Deleted = true
			}
			// Close dialog (whether Delete or Cancel was selected)
			result.Model.DeleteDialog = nil
			if result.Deleted {
				result.Quit = true
			}
			return result
		default:
			// Any other key closes the dialog and returns to detail view
			result.Model.DeleteDialog = nil
			return result
		}
	}

	// Accept any key to close help overlay (except ? which toggles help)
	if m.ShowHelp && key.String() != "?" {
		result.Model.ShowHelp = false
		// Update viewport for scrolling
		var cmd tea.Cmd
		result.Model.Viewport, cmd = m.Viewport.Update(key)
		_ = cmd
		return result
	}

	switch HandleKey(key.String()) {
	case ActionQuit:
		result.Quit = true
		return result
	case ActionMoveUp:
		if m.Cursor > 0 {
			result.Model.Cursor--
		}
	case ActionMoveDown:
		if m.Cursor < len(m.History.ScanResults.Hosts)-1 {
			result.Model.Cursor++
		}
	case ActionScanAllPorts:
		if m.Cursor < len(m.History.ScanResults.Hosts) {
			result.ScanAllPorts = true
			result.SelectedHostIP = m.History.ScanResults.Hosts[m.Cursor].IP
			result.ScanHistoryPath = m.HistoryPath
		}
	case ActionDelete:
		if m.NodePath != "" {
			result.Model.DeleteDialog = &deletepkg.HistoryDeleteDialog{
				Target:      nil,
				ItemType:    m.NodeItemType,
				ItemName:    m.NodeName,
				CursorOnYes: true,
			}
		}
	case ActionHelp:
		result.Model.ShowHelp = !result.Model.ShowHelp
	}

	// Update viewport for scrolling
	var cmd tea.Cmd
	result.Model.Viewport, cmd = m.Viewport.Update(key)
	_ = cmd

	return result
}

func performDeleteSync(path string) {
	if path != "" {
		history.Delete(path)
	}
}
