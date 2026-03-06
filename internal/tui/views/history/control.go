package historyview

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/backendsystems/nibble/internal/history"
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
		return handleKeyMsg(m, msg)
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

type treeLoadedMsg struct {
	tree         []*TreeNode
	selectedPath string
}

func loadTreeCmd() tea.Cmd {
	return func() tea.Msg {
		tree, selectedPath, err := buildHistoryTree()
		if err != nil {
			return treeLoadedMsg{tree: []*TreeNode{}, selectedPath: ""}
		}
		return treeLoadedMsg{tree: tree, selectedPath: selectedPath}
	}
}

func buildHistoryTree() ([]*TreeNode, string, error) {
	baseDir, err := history.HistoryDir()
	if err != nil {
		return nil, "", err
	}

	// Map of interface -> network -> scans
	type networkData struct {
		scans []*TreeNode
	}
	interfaceMap := make(map[string]map[string]*networkData)

	err = filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".json" {
			return nil
		}

		// Load scan to get metadata
		scanData, err := history.Load(path)
		if err != nil {
			return nil
		}

		// Parse path: baseDir/interface/network/file.json
		relPath, _ := filepath.Rel(baseDir, path)
		parts := strings.Split(relPath, string(filepath.Separator))
		if len(parts) < 3 {
			return nil
		}

		interfaceName := parts[0]
		networkName := parts[1]

		// Create scan node
		scanNode := &TreeNode{
			Type:     NodeScan,
			Name:     scanData.ScanMetadata.Created.Format("2006 Jan 2 15:04"),
			Path:     path,
			ScanData: &scanData,
			Level:    2,
		}

		if interfaceMap[interfaceName] == nil {
			interfaceMap[interfaceName] = make(map[string]*networkData)
		}
		if interfaceMap[interfaceName][networkName] == nil {
			interfaceMap[interfaceName][networkName] = &networkData{}
		}

		interfaceMap[interfaceName][networkName].scans = append(
			interfaceMap[interfaceName][networkName].scans,
			scanNode,
		)

		return nil
	})

	if err != nil {
		return nil, "", err
	}

	// Build tree structure
	var tree []*TreeNode

	// Sort interfaces
	var interfaces []string
	for iface := range interfaceMap {
		interfaces = append(interfaces, iface)
	}
	sort.Strings(interfaces)

	for _, iface := range interfaces {
		ifaceNode := &TreeNode{
			Type:     NodeInterface,
			Name:     iface,
			Path:     filepath.Join(baseDir, iface),
			Expanded: false,
			Level:    0,
		}

		// Sort networks
		var networks []string
		for net := range interfaceMap[iface] {
			networks = append(networks, net)
		}
		sort.Strings(networks)

		for _, net := range networks {
			// Convert network folder name back to CIDR (e.g., "192.168.1.0_24" -> "192.168.1.0/24")
			cidr := strings.ReplaceAll(net, "_", "/")

			netNode := &TreeNode{
				Type:     NodeNetwork,
				Name:     cidr,
				Path:     filepath.Join(baseDir, iface, net),
				Expanded: false,
				Level:    1,
			}

			// Sort scans by date (newest first)
			scans := interfaceMap[iface][net].scans
			sort.Slice(scans, func(i, j int) bool {
				return scans[i].ScanData.ScanMetadata.Created.After(scans[j].ScanData.ScanMetadata.Created)
			})

			netNode.Children = scans
			ifaceNode.Children = append(ifaceNode.Children, netNode)
		}

		tree = append(tree, ifaceNode)
	}

	// Load and expand tree to show selected path
	selectedPath, _ := history.LoadViewState()
	if selectedPath != "" {
		expandTreeForPath(tree, selectedPath)
	}

	return tree, selectedPath, nil
}

func flattenTree(tree []*TreeNode) []*TreeNode {
	var flat []*TreeNode
	for _, node := range tree {
		flat = append(flat, node)
		if node.Expanded {
			flat = append(flat, flattenTree(node.Children)...)
		}
	}
	return flat
}

// collectExpandedState collects the names of all expanded nodes
func collectExpandedState(tree []*TreeNode) map[string]bool {
	state := make(map[string]bool)
	for _, node := range tree {
		if node.Expanded {
			state[node.Name] = true
		}
		// Recursively collect from children
		for k, v := range collectExpandedState(node.Children) {
			state[k] = v
		}
	}
	return state
}

// saveViewState saves the selected item path to persistent storage
func saveViewState(flatList []*TreeNode, cursor int) {
	selectedPath := ""
	if cursor >= 0 && cursor < len(flatList) && flatList[cursor] != nil {
		selectedPath = flatList[cursor].Path
	}
	history.SaveViewState(selectedPath)
}

// findCursorByPath finds the cursor position for a given path in the flattened list
func findCursorByPath(flatList []*TreeNode, path string) int {
	if path == "" {
		return 0
	}
	for i, node := range flatList {
		if node != nil && node.Path == path {
			return i
		}
	}
	return 0
}

// expandTreeForPath expands all parent nodes needed to show a specific path
func expandTreeForPath(tree []*TreeNode, path string) {
	for _, node := range tree {
		if nodeContainsPath(node, path) {
			node.Expanded = true
			expandTreeForPath(node.Children, path)
		}
	}
}

// nodeContainsPath checks if a node or its children contain the given path
func nodeContainsPath(node *TreeNode, path string) bool {
	if node.Path == path {
		return true
	}
	for _, child := range node.Children {
		if nodeContainsPath(child, path) {
			return true
		}
	}
	return false
}
