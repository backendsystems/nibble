package historyview

import (
	"github.com/backendsystems/nibble/internal/history"
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

// performDeleteSync recursively deletes a node and all its children
func performDeleteSync(node *TreeNode) {
	if node == nil {
		return
	}

	switch node.Type {
	case NodeScan:
		if node.Path != "" {
			history.Delete(node.Path)
		}
	case NodeNetwork:
		for _, child := range node.Children {
			if child != nil && child.Path != "" {
				history.Delete(child.Path)
			}
		}
	case NodeInterface:
		for _, netNode := range node.Children {
			for _, scanNode := range netNode.Children {
				if scanNode != nil && scanNode.Path != "" {
					history.Delete(scanNode.Path)
				}
			}
		}
	}
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
	tree         []*TreeNode
	selectedPath string
}

// loadTreeCmd is a Bubble Tea command that loads the history tree asynchronously
func loadTreeCmd() tea.Cmd {
	return func() tea.Msg {
		tree, selectedPath, err := buildHistoryTree()
		if err != nil {
			return treeLoadedMsg{tree: []*TreeNode{}, selectedPath: ""}
		}
		return treeLoadedMsg{tree: tree, selectedPath: selectedPath}
	}
}
