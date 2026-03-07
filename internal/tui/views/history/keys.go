package historyview

import (
	"github.com/backendsystems/nibble/internal/tui/views/history/delete"
	detailsview "github.com/backendsystems/nibble/internal/tui/views/history/details"
	tea "github.com/charmbracelet/bubbletea"
)

func handleKeyMsg(m Model, key tea.KeyMsg) UpdateResult {
	result := UpdateResult{Model: m}

	inDeleteDialog := m.DeleteDialog != nil
	action := HandleKey(key.String(), inDeleteDialog)

	if inDeleteDialog {
		return handleDeleteDialog(result, action)
	}

	// Accept any key to close help overlay
	if m.ShowHelp {
		result.Model.ShowHelp = false
		return result
	}

	return handleListKey(result, action)
}

func handleDeleteDialog(result UpdateResult, action Action) UpdateResult {
	switch action {
	case ActionToggle:
		result.Model.DeleteDialog.Toggle()
	case ActionConfirmYes:
		if result.Model.DeleteDialog.IsDeleteSelected() {
			currentCursor := result.Model.Cursor
			if node, ok := result.Model.DeleteDialog.Target.(*TreeNode); ok {
				performDeleteSync(node)
			}
			tree, _, _ := buildHistoryTree()
			result.Model.Tree = tree
			result.Model.FlatList = flattenTree(tree)
			if currentCursor >= len(result.Model.FlatList) && len(result.Model.FlatList) > 0 {
				result.Model.Cursor = len(result.Model.FlatList) - 1
			} else if len(result.Model.FlatList) == 0 {
				result.Model.Cursor = 0
			} else {
				result.Model.Cursor = currentCursor
			}
		}
		result.Model.DeleteDialog = nil
	case ActionConfirmNo:
		result.Model.DeleteDialog = nil
	}
	return result
}

func handleListKey(result UpdateResult, action Action) UpdateResult {
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
				result.Model.Mode = ViewDetail
				result.Model.Details = detailsview.Model{
					History:      *node.ScanData,
					HistoryPath:  node.Path,
					NodePath:     node.Path,
					NodeName:     node.Name,
					NodeItemType: "scan",
					WindowW:      result.Model.WindowW,
					WindowH:      result.Model.WindowH,
				}
				return result
			}
			if node != nil {
				wasExpanded := node.Expanded
				node.Expanded = !node.Expanded
				result.Model.FlatList = flattenTree(result.Model.Tree)
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
			if node != nil && node.Expanded && (node.Type == NodeInterface || node.Type == NodeNetwork) {
				node.Expanded = false
				result.Model.FlatList = flattenTree(result.Model.Tree)
			} else if node != nil && node.Level > 0 {
				for i := result.Model.Cursor - 1; i >= 0; i-- {
					parent := result.Model.FlatList[i]
					if parent != nil && parent.Level == node.Level-1 {
						result.Model.Cursor = i
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
				result.Model.DeleteDialog = &delete.HistoryDeleteDialog{
					Target:      node,
					ItemType:    itemType,
					ItemName:    node.Name,
					CursorOnYes: true,
				}
			}
		}
	case ActionHelp:
		result.Model.ShowHelp = !result.Model.ShowHelp
	}

	if !result.Quit && result.Model.Mode == ViewList {
		saveViewState(result.Model.FlatList, result.Model.Cursor)
	}

	return result
}
