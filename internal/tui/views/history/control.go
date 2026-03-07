package historyview

import (
	"github.com/backendsystems/nibble/internal/tui/views/history/delete"
	detailsview "github.com/backendsystems/nibble/internal/tui/views/history/details"
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

	// Route detail view messages
	if m.Mode == ViewDetail {
		detailResult := m.Details.Update(msg)
		result.Model.Details = detailResult.Model
		result.Cmd = detailResult.Cmd
		if detailResult.Deleted {
			// Deletions must reload the tree so removed items disappear from list view.
			tree, selectedPath, _ := buildHistoryTree()
			result.Model.Tree = tree
			result.Model.FlatList = flattenTree(tree)
			if selectedPath != "" {
				result.Model.Cursor = findCursorByPath(result.Model.FlatList, selectedPath)
			}
			if result.Model.Cursor >= len(result.Model.FlatList) && len(result.Model.FlatList) > 0 {
				result.Model.Cursor = len(result.Model.FlatList) - 1
			}
			result.Model.Mode = ViewList
			result.Model.Details = detailsview.Model{}
			saveViewState(result.Model.FlatList, result.Model.Cursor)
		} else if detailResult.Quit {
			// Back from details should preserve current tree/list state.
			result.Model.Mode = ViewList
			result.Model.Details = detailsview.Model{}
			saveViewState(result.Model.FlatList, result.Model.Cursor)
		}
		if detailResult.ScanAllPorts {
			result.ScanAllPorts = true
			result.SelectedHostIP = detailResult.SelectedHostIP
			result.ScanHistoryPath = detailResult.ScanHistoryPath
		}
		return result
	}

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
	}

	// Update viewport sizes on window resize or initialization
	if result.Model.WindowW > 0 {
		oldListHeight := result.Model.Viewport.Height
		result.Model = result.Model.SetListViewportSize(result.Model.WindowW, result.Model.WindowH)
		// Reset scroll offset if viewport height changed significantly
		if oldListHeight != result.Model.Viewport.Height {
			result.Model.Viewport.YOffset = 0
		}
		// Also update details viewport size
		result.Model.Details.WindowW = result.Model.WindowW
		result.Model.Details.WindowH = result.Model.WindowH
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
				result.Model.Details = detailsview.Model{
					History:      *node.ScanData,
					HistoryPath:  node.Path,
					NodePath:     node.Path,
					NodeName:     node.Name,
					NodeItemType: "scan",
					WindowW:      m.WindowW,
					WindowH:      m.WindowH,
				}
				return result
			}
			// Toggle folder expansion
			if node != nil {
				wasExpanded := node.Expanded
				node.Expanded = !node.Expanded
				result.Model.FlatList = flattenTree(result.Model.Tree)
				// When opening a folder, move selection into the first child.
				if !wasExpanded && node.Expanded && len(node.Children) > 0 {
					if result.Model.Cursor+1 < len(result.Model.FlatList) {
						next := result.Model.FlatList[result.Model.Cursor+1]
						if next != nil && next.Level == node.Level+1 {
							result.Model.Cursor++
						}
					}
				}
			}
		}
	case ActionCollapse:
		if result.Model.Cursor >= 0 && result.Model.Cursor < len(result.Model.FlatList) {
			node := result.Model.FlatList[result.Model.Cursor]
			// If this is an expanded folder, collapse it
			if node != nil && node.Expanded && (node.Type == NodeInterface || node.Type == NodeNetwork) {
				node.Expanded = false
				result.Model.FlatList = flattenTree(result.Model.Tree)
			} else if node != nil && node.Level > 0 {
				// If already collapsed (or a leaf), move selection to parent node.
				for i := result.Model.Cursor - 1; i >= 0; i-- {
					parent := result.Model.FlatList[i]
					if parent != nil && parent.Level == node.Level-1 {
						result.Model.Cursor = i
						// Left on child should also close the parent we moved up into.
						if parent.Expanded && (parent.Type == NodeInterface || parent.Type == NodeNetwork) {
							parent.Expanded = false
							result.Model.FlatList = flattenTree(result.Model.Tree)
						}
						break
					}
				}
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
