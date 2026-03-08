package historydetailview

import (
	"github.com/backendsystems/nibble/internal/history"
	"github.com/backendsystems/nibble/internal/scanner/shared"
	deletepkg "github.com/backendsystems/nibble/internal/tui/views/history/delete"
	tea "github.com/charmbracelet/bubbletea"
)

func HandleKey(key string) Action {
	switch key {
	case "q", "esc", "left", "a", "h":
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
		case "delete":
			// Delete key always confirms delete regardless of cursor position
			performDeleteSync(m.NodePath)
			result.Deleted = true
			result.Model.DeleteDialog = nil
			result.Quit = true
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
		if m.Scanning {
			result.Model.Scanning = false
			result.Model.ProgressChan = nil
			result.Model.ScannedHostStr = ""
			result.Cmd = tea.Batch(
				m.Stopwatch.Stop(),
				drainProgressChan(m.ProgressChan),
			)
		}
		result.Quit = true
		return result
	case ActionMoveUp:
		if !m.Scanning && m.Cursor > 0 {
			result.Model.Cursor--
		}
	case ActionMoveDown:
		if !m.Scanning && m.Cursor < len(m.History.ScanResults.Hosts)-1 {
			result.Model.Cursor++
		}
	case ActionScanAllPorts:
		if m.Cursor < len(m.History.ScanResults.Hosts) {
			result.ScanAllPorts = true
			result.SelectedHostIP = m.History.ScanResults.Hosts[m.Cursor].IP
			result.ScanHistoryPath = m.HistoryPath
			result.Model.ScanningHostIdx = m.Cursor // Track which host is being scanned
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

	return result
}

func drainProgressChan(ch <-chan shared.ProgressUpdate) tea.Cmd {
	return func() tea.Msg {
		if ch == nil {
			return nil
		}
		go func() {
			for range ch {
			}
		}()
		return nil
	}
}

func performDeleteSync(path string) {
	if path != "" {
		history.Delete(path)
	}
}
