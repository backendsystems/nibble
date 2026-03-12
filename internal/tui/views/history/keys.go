package historyview

import (
	"github.com/backendsystems/nibble/internal/history"
	"github.com/backendsystems/nibble/internal/tui/views/history/delete"
	detailsview "github.com/backendsystems/nibble/internal/tui/views/history/details"
	historytree "github.com/backendsystems/nibble/internal/tui/views/history/tree"
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

// toggleFolder expands or collapses a folder, scrolling to show all children without losing the cursor.
func toggleFolder(result UpdateResult, node *historytree.Node) UpdateResult {
	// Cancel background loads before collapsing
	if node.Expanded {
		historytree.CancelLoads(node)
	}
	node.Expanded = !node.Expanded
	result.Model.FlatList = historytree.Flatten(result.Model.Tree)
	if !node.Expanded || len(node.Children) == 0 {
		return result
	}
	// Move cursor to first child
	if result.Model.Cursor+1 < len(result.Model.FlatList) {
		if next := result.Model.FlatList[result.Model.Cursor+1]; next != nil && next.Level == node.Level+1 {
			result.Model.Cursor++
		}
	}
	// Scroll down to reveal the last child, capped so the cursor stays on screen
	lastChildIdx := result.Model.Cursor + len(node.Children) - 1
	if lastChildIdx >= len(result.Model.FlatList) {
		lastChildIdx = len(result.Model.FlatList) - 1
	}
	if h := result.Model.Viewport.Height; h > 0 {
		wantOffset := min(lastChildIdx-h+1, result.Model.Cursor)
		if wantOffset > result.Model.Viewport.YOffset {
			result.Model.Viewport.YOffset = wantOffset
		}
	}
	// Kick off background host/port count loads for newly visible scans
	if node.Type == NodeNetwork {
		if cmds := historytree.LoadNetworkScanCountsCmd(node); len(cmds) > 0 {
			result.Cmd = tea.Sequence(cmds...)
		}
	}
	return result
}

func handleDeleteDialog(result UpdateResult, action Action) UpdateResult {
	switch action {
	case ActionToggle:
		result.Model.DeleteDialog.Toggle()
	case ActionConfirmYes:
		if result.Model.DeleteDialog.IsDeleteSelected() {
			result = executeDelete(result)
		}
		result.Model.DeleteDialog = nil
	case ActionConfirmDelete:
		result = executeDelete(result)
		result.Model.DeleteDialog = nil
	case ActionConfirmNo:
		result.Model.DeleteDialog = nil
	}
	return result
}

func executeDelete(result UpdateResult) UpdateResult {
	nextPath := ""
	if node, ok := result.Model.DeleteDialog.Target.(*TreeNode); ok {
		nextPath = nextSelectionPathAfterDelete(result.Model.FlatList, node.Path)
		performDeleteSync(node)
	}
	tree, _ := historytree.Build()
	if nextPath != "" {
		historytree.ExpandAncestorsForPath(tree, nextPath)
	}
	result.Model.Tree = tree
	result.Model.FlatList = historytree.Flatten(tree)
	if nextPath != "" {
		result.Model.Cursor = historytree.FindCursorByPath(result.Model.FlatList, nextPath)
	} else if result.Model.Cursor >= len(result.Model.FlatList) && len(result.Model.FlatList) > 0 {
		result.Model.Cursor = len(result.Model.FlatList) - 1
	} else if len(result.Model.FlatList) == 0 {
		result.Model.Cursor = 0
	}
	result.Cmd = historytree.LoadCountsForExpandedNodes(result.Model.Tree)
	saveViewState(result.Model.FlatList, result.Model.Cursor)
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
			if node != nil && node.Type == NodeScan {
				// Lazy-load: ScanData is nil until first open to avoid reading all files at startup.
				if node.ScanData == nil {
					scanData, err := history.Load(node.Path)
					if err != nil {
						result.Model.ErrorMsg = "failed to load scan: " + err.Error()
						return result
					}
					node.ScanData = &scanData
				}
				result.Model.Mode = ViewDetail
				savedCursor := 0
				if result.Model.DetailCursors != nil {
					savedCursor = result.Model.DetailCursors[node.Path]
				}
				details := detailsview.Model{
					History:      *node.ScanData,
					HistoryPath:  node.Path,
					NodePath:     node.Path,
					NodeName:     node.Name,
					NodeItemType: "scan",
					WindowW:      result.Model.WindowW,
					WindowH:      result.Model.WindowH,
					Cursor:       savedCursor,
					HoveredHelpItem: -1,
				}
				details = details.SetViewportSize(result.Model.WindowW, result.Model.WindowH)
				details = details.ScrollToSelected()
				result.Model.Details = details
				return result
			}
			if node != nil {
				result = toggleFolder(result, node)
			}
		}
	case ActionCollapse:
		if result.Model.Cursor >= 0 && result.Model.Cursor < len(result.Model.FlatList) {
			node := result.Model.FlatList[result.Model.Cursor]
			if node != nil && node.Expanded && (node.Type == NodeInterface || node.Type == NodeNetwork) {
				historytree.CancelLoads(node)
				node.Expanded = false
				result.Model.FlatList = historytree.Flatten(result.Model.Tree)
			} else if node != nil && node.Level > 0 {
				for i := result.Model.Cursor - 1; i >= 0; i-- {
					parent := result.Model.FlatList[i]
					if parent != nil && parent.Level == node.Level-1 {
						result.Model.Cursor = i
						if parent.Expanded && (parent.Type == NodeInterface || parent.Type == NodeNetwork) {
							historytree.CancelLoads(parent)
							parent.Expanded = false
							result.Model.FlatList = historytree.Flatten(result.Model.Tree)
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
