package historyview

import (
	"github.com/backendsystems/nibble/internal/tui/views/history/delete"
	tea "github.com/charmbracelet/bubbletea"
)

type Action int

const (
	ActionNone Action = iota
	ActionQuit
	ActionMoveUp
	ActionMoveDown
	ActionToggle
	ActionCollapse
	ActionDelete
	ActionConfirmYes
	ActionConfirmNo
	ActionHelp
)

// HandleKey converts a key press into an action
func HandleKey(key string, inDeleteDialog bool) Action {
	if inDeleteDialog {
		switch key {
		case "left", "a", "h", "right", "d", "l":
			return ActionToggle // Toggle between Delete/Cancel
		case "enter":
			return ActionConfirmYes // Confirm selection
		case "esc", "q":
			return ActionConfirmNo // Cancel
		default:
			return ActionNone
		}
	}

	// Accept any key to close help overlay if in help mode (handled in Update logic)
	switch key {
	case "q", "esc":
		return ActionQuit
	case "up", "w", "k":
		return ActionMoveUp
	case "down", "s", "j":
		return ActionMoveDown
	case "enter", "right", "d", "l":
		return ActionToggle
	case "left", "a", "h":
		return ActionCollapse
	case "delete":
		return ActionDelete
	case "?":
		return ActionHelp
	default:
		return ActionNone
	}
}

func (m Model) Init() tea.Cmd {
	return loadTreeCmd()
}

func (m Model) Update(msg tea.Msg) UpdateResult {
	result := UpdateResult{Model: m}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		result = handleKeyMsg(m, msg)
	case treeLoadedMsg:
		result.Model.Tree = msg.tree
		result.Model.FlatList = flattenTree(result.Model.Tree)
		// Restore cursor to previously selected item
		if msg.selectedPath != "" {
			result.Model.Cursor = findCursorByPath(result.Model.FlatList, msg.selectedPath)
		}
		if result.Model.Cursor >= len(result.Model.FlatList) {
			result.Model.Cursor = 0
		}
	default:
		// Update viewports for scrolling
		var cmd tea.Cmd
		result.Model.Viewport, cmd = m.Viewport.Update(msg)
		_ = cmd
		result.Model.DetailViewport, cmd = m.DetailViewport.Update(msg)
		_ = cmd
	}

	// Update viewport sizes on window resize or initialization
	if result.Model.WindowW > 0 {
		oldListHeight := result.Model.Viewport.Height
		oldDetailHeight := result.Model.DetailViewport.Height
		result.Model = result.Model.SetListViewportSize(result.Model.WindowW, result.Model.WindowH)
		result.Model = result.Model.SetDetailViewportSize(result.Model.WindowW, result.Model.WindowH)
		// Reset scroll offset if viewport height changed significantly
		if oldListHeight != result.Model.Viewport.Height {
			result.Model.Viewport.YOffset = 0
		}
		if oldDetailHeight != result.Model.DetailViewport.Height {
			result.Model.DetailViewport.YOffset = 0
		}
	}

	// Pre-render viewport content based on current view mode
	// This must happen on all update paths to keep viewport in sync with model
	if !result.Quit {
		result.Model = updateViewportContent(result.Model)
	}

	return result
}

func handleKeyMsg(m Model, key tea.KeyMsg) UpdateResult {
	result := UpdateResult{Model: m}

	// If in detail view, handle detail view keys
	if m.Mode == ViewDetail {
		return handleDetailKeyMsg(m, key)
	}

	inDeleteDialog := m.DeleteDialog != nil
	action := HandleKey(key.String(), inDeleteDialog)

	// Handle delete dialog actions
	if inDeleteDialog {
		switch action {
		case ActionToggle:
			// Toggle between Delete and Cancel
			result.Model.DeleteDialog.Toggle()
			return result
		case ActionConfirmYes:
			// User pressed Enter - execute the selected action
			if result.Model.DeleteDialog.IsDeleteSelected() {
				// Delete was selected
				currentCursor := result.Model.Cursor

				if node, ok := result.Model.DeleteDialog.Target.(*TreeNode); ok {
					performDeleteSync(node)
				}

				// Reload tree
				tree, _, _ := buildHistoryTree()
				result.Model.Tree = tree
				result.Model.FlatList = flattenTree(tree)

				// Keep cursor at same position, or adjust if out of bounds
				if currentCursor >= len(result.Model.FlatList) && len(result.Model.FlatList) > 0 {
					result.Model.Cursor = len(result.Model.FlatList) - 1
				} else if len(result.Model.FlatList) == 0 {
					result.Model.Cursor = 0
				} else {
					result.Model.Cursor = currentCursor
				}
			}
			// Close dialog (whether Delete or Cancel was selected)
			result.Model.DeleteDialog = nil
			return result
		case ActionConfirmNo:
			// User pressed Esc - cancel
			result.Model.DeleteDialog = nil
			return result
		}
		return result
	}

	// Accept any key to close help overlay
	if m.ShowHelp {
		result.Model.ShowHelp = false
		return result
	}

	switch action {
	case ActionQuit:
		result.Quit = true
	case ActionMoveUp:
		if result.Model.Cursor > 0 {
			result.Model.Cursor--
		}
	case ActionMoveDown:
		if result.Model.Cursor < len(result.Model.FlatList)-1 {
			result.Model.Cursor++
		}
	case ActionToggle:
		if result.Model.Cursor >= 0 && result.Model.Cursor < len(result.Model.FlatList) {
			node := result.Model.FlatList[result.Model.Cursor]
			if node != nil && node.Type == NodeScan && node.ScanData != nil {
				// Switch to detail view
				result.Model.Mode = ViewDetail
				result.Model.DetailHistory = node.ScanData
				result.Model.DetailPath = node.Path
				result.Model.DetailCursor = 0
				return result
			}
			// Toggle folder expansion
			if node != nil {
				node.Expanded = !node.Expanded
				result.Model.FlatList = flattenTree(result.Model.Tree)
			}
		}
	case ActionCollapse:
		if result.Model.Cursor >= 0 && result.Model.Cursor < len(result.Model.FlatList) {
			node := result.Model.FlatList[result.Model.Cursor]
			// If this is an expanded folder, collapse it
			if node != nil && node.Expanded && (node.Type == NodeInterface || node.Type == NodeNetwork) {
				node.Expanded = false
				result.Model.FlatList = flattenTree(result.Model.Tree)
			}
		}
	case ActionDelete:
		if result.Model.Cursor >= 0 && result.Model.Cursor < len(result.Model.FlatList) {
			node := result.Model.FlatList[result.Model.Cursor]
			// Only allow deletion of actual nodes with valid data
			if node != nil && node.Type >= NodeInterface && node.Type <= NodeScan {
				var itemType string
				switch node.Type {
				case NodeScan:
					itemType = "scan"
				case NodeNetwork:
					itemType = "all scans in network"
				case NodeInterface:
					itemType = "all scans on interface"
				}

				// Show delete dialog
				result.Model.DeleteDialog = &delete.HistoryDeleteDialog{
					Target:      node,
					ItemType:    itemType,
					ItemName:    node.Name,
					CursorOnYes: true, // Default to Delete
				}
			}
		}
	case ActionHelp:
		result.Model.ShowHelp = !result.Model.ShowHelp
	}

	// Save state once at the end if not quitting or switching views
	if !result.Quit && result.Model.Mode == ViewList {
		saveViewState(result.Model.FlatList, result.Model.Cursor)
	}

	return result
}

func handleDetailKeyMsg(m Model, key tea.KeyMsg) UpdateResult {
	result := UpdateResult{Model: m}

	// Handle delete dialog in detail view
	if m.DeleteDialog != nil {
		switch key.String() {
		case "left", "a", "h", "right", "d", "l":
			// Toggle between Delete/Cancel
			result.Model.DeleteDialog.Toggle()
			return result
		case "enter":
			// User pressed Enter - execute the selected action
			if result.Model.DeleteDialog.IsDeleteSelected() {
				// Delete was selected
				if node, ok := result.Model.DeleteDialog.Target.(*TreeNode); ok {
					performDeleteSync(node)
				}

				// Reload tree
				tree, _, _ := buildHistoryTree()
				result.Model.Tree = tree
				result.Model.FlatList = flattenTree(tree)
				// Save state after deletion
				saveViewState(result.Model.FlatList, result.Model.Cursor)
			}
			// Close dialog (whether Delete or Cancel was selected)
			result.Model.DeleteDialog = nil
			return result
		default:
			// Any other key closes the dialog and returns to detail view
			result.Model.DeleteDialog = nil
			return result
		}
	}

	// Accept any key to close help overlay
	if m.ShowHelp {
		result.Model.ShowHelp = false
		return result
	}

	switch key.String() {
	case "q", "esc":
		// Go back to list view
		result.Model.Mode = ViewList
		result.Model.DetailHistory = nil
		result.Model.DetailPath = ""
	case "up", "w", "k":
		if result.Model.DetailCursor > 0 {
			result.Model.DetailCursor--
		}
	case "down", "s", "j":
		if result.Model.DetailHistory != nil && result.Model.DetailCursor < len(result.Model.DetailHistory.ScanResults.Hosts)-1 {
			result.Model.DetailCursor++
		}
	case "enter":
		// Trigger all-port scan on selected host
		if result.Model.DetailHistory != nil && result.Model.DetailCursor < len(result.Model.DetailHistory.ScanResults.Hosts) {
			result.ScanAllPorts = true
			result.SelectedHostIP = result.Model.DetailHistory.ScanResults.Hosts[result.Model.DetailCursor].IP
			result.ScanHistoryPath = result.Model.DetailPath
		}
	case "?":
		result.Model.ShowHelp = !result.Model.ShowHelp
	}

	return result
}
