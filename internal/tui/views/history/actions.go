package historyview

import (
	"github.com/backendsystems/nibble/internal/history"
	historytree "github.com/backendsystems/nibble/internal/tui/views/history/tree"
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
	ActionConfirmDelete
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
		case "delete":
			return ActionConfirmDelete // Always delete
		case "esc", "q":
			return ActionConfirmNo // Cancel
		default:
			return ActionNone
		}
	}

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

// performDeleteSync recursively deletes a node and all its children,
// and cleans up any saved detail cursors for the removed scan files.
func performDeleteSync(node *TreeNode) {
	if node == nil {
		return
	}

	var scanPaths []string

	switch node.Type {
	case NodeScan:
		if node.Path != "" {
			history.Delete(node.Path)
			scanPaths = append(scanPaths, node.Path)
		}
	case NodeNetwork:
		for _, child := range node.Children {
			if child != nil && child.Path != "" {
				history.Delete(child.Path)
				scanPaths = append(scanPaths, child.Path)
			}
		}
	case NodeInterface:
		for _, netNode := range node.Children {
			for _, scanNode := range netNode.Children {
				if scanNode != nil && scanNode.Path != "" {
					history.Delete(scanNode.Path)
					scanPaths = append(scanPaths, scanNode.Path)
				}
			}
		}
	}

	history.DeleteDetailCursors(scanPaths)
}

// saveViewState saves the selected item path to persistent storage
func saveViewState(flatList []*TreeNode, cursor int) {
	selectedPath := ""
	if cursor >= 0 && cursor < len(flatList) && flatList[cursor] != nil {
		selectedPath = flatList[cursor].Path
	}
	history.SaveViewState(selectedPath)
}

// treeLoadedMsg is a message indicating the tree has been loaded
type treeLoadedMsg struct {
	tree          []*TreeNode
	selectedPath  string
	detailCursors map[string]int
}

// loadTreeCmd is a Bubble Tea command that loads the history tree asynchronously
func loadTreeCmd() tea.Cmd {
	return func() tea.Msg {
		selectedPath, detailCursors, _ := history.LoadViewState()
		tree, err := historytree.Build()
		if err != nil {
			return treeLoadedMsg{tree: []*TreeNode{}, selectedPath: ""}
		}
		if selectedPath != "" {
			historytree.ExpandAncestorsForPath(tree, selectedPath)
		}
		return treeLoadedMsg{tree: tree, selectedPath: selectedPath, detailCursors: detailCursors}
	}
}

func syncScanNode(tree []*TreeNode, path string, updated history.ScanHistory) {
	if path == "" || updated.Version == "" {
		return
	}

	for _, node := range tree {
		if node == nil {
			continue
		}
		if node.Path == path && node.Type == NodeScan {
			updatedCopy := updated
			node.ScanData = &updatedCopy
			node.Counts = &ScanCounts{
				Hosts: updated.ScanResults.HostsFound,
				Ports: updated.ScanResults.PortsFound,
			}
			return
		}
		syncScanNode(node.Children, path, updated)
	}
}

func nextSelectionPathAfterDelete(flatList []*TreeNode, deletedPath string) string {
	if deletedPath == "" {
		return ""
	}

	for i, node := range flatList {
		if node == nil || node.Path != deletedPath {
			continue
		}

		level := node.Level

		// Prefer the next sibling in the same parent folder.
		for j := i + 1; j < len(flatList); j++ {
			next := flatList[j]
			if next == nil {
				continue
			}
			if next.Level < level {
				break
			}
			if next.Level == level {
				return next.Path
			}
		}

		// Then prefer the previous sibling in the same parent folder.
		for j := i - 1; j >= 0; j-- {
			prev := flatList[j]
			if prev == nil {
				continue
			}
			if prev.Level < level {
				break
			}
			if prev.Level == level {
				return prev.Path
			}
		}

		// If there are no siblings left, stay anchored on the parent folder.
		for j := i - 1; j >= 0; j-- {
			parent := flatList[j]
			if parent != nil && parent.Level == level-1 {
				return parent.Path
			}
		}

		// Fallback to the next visible node, then the previous one.
		for j := i + 1; j < len(flatList); j++ {
			if flatList[j] != nil {
				return flatList[j].Path
			}
		}
		for j := i - 1; j >= 0; j-- {
			if flatList[j] != nil {
				return flatList[j].Path
			}
		}
		break
	}

	return ""
}
