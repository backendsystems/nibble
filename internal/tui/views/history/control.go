package historyview

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/backendsystems/nibble/internal/history"
	"github.com/charmbracelet/bubbles/textinput"
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
	ActionFilter
	ActionHelp
)

func HandleKey(key string, inDeleteDialog bool, inFilter bool) Action {
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

	if inFilter {
		switch key {
		case "esc":
			return ActionFilter // Toggle off
		default:
			return ActionNone // Let textinput handle it
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
	case "/":
		return ActionFilter
	case "?":
		return ActionHelp
	default:
		return ActionNone
	}
}

func (m Model) Init() tea.Cmd {
	// Initialize filter input
	ti := textinput.New()
	ti.Placeholder = "Filter scans..."
	ti.CharLimit = 50
	m.FilterInput = ti
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

	// If filter is active, handle textinput
	if m.FilterActive {
		var cmd tea.Cmd
		result.Model.FilterInput, cmd = m.FilterInput.Update(key)
		result.Model.FilterText = result.Model.FilterInput.Value()
		result.Model.FlatList = filterTree(result.Model.Tree, result.Model.FilterText)
		_ = cmd
		return result
	}

	inDeleteDialog := m.DeleteDialog != nil
	action := HandleKey(key.String(), inDeleteDialog, m.FilterActive)

	// Handle delete dialog actions
	if inDeleteDialog {
		switch action {
		case ActionToggle:
			// Toggle between Delete and Cancel
			result.Model.DeleteDialog.CursorOnYes = !result.Model.DeleteDialog.CursorOnYes
			return result
		case ActionConfirmYes:
			// User pressed Enter - execute the selected action
			if result.Model.DeleteDialog.CursorOnYes {
				// Delete was selected
				// Remember which nodes were expanded before deletion
				expandedState := collectExpandedState(result.Model.Tree)
				currentCursor := result.Model.Cursor

				performDeleteSync(result.Model.DeleteDialog.Target)

				// Reload tree and restore expanded state
				tree, _ := buildHistoryTree()
				restoreExpandedState(tree, expandedState)
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

	switch action {
	case ActionFilter:
		result.Model.FilterActive = !result.Model.FilterActive
		if result.Model.FilterActive {
			result.Model.FilterInput.Focus()
		} else {
			result.Model.FilterInput.Blur()
			result.Model.FilterText = ""
			result.Model.FilterInput.SetValue("")
			result.Model.FlatList = flattenTree(result.Model.Tree)
		}
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
		if result.Model.Cursor < len(result.Model.FlatList) {
			node := result.Model.FlatList[result.Model.Cursor]
			if node.Type == NodeScan && node.ScanData != nil {
				// Switch to detail view
				result.Model.Mode = ViewDetail
				result.Model.DetailHistory = node.ScanData
				result.Model.DetailPath = node.Path
				result.Model.DetailCursor = 0
				return result
			}
			// Toggle folder expansion
			node.Expanded = !node.Expanded
			result.Model.FlatList = flattenTree(result.Model.Tree)
		}
	case ActionCollapse:
		if result.Model.Cursor < len(result.Model.FlatList) {
			node := result.Model.FlatList[result.Model.Cursor]
			// If this is an expanded folder, collapse it
			if node.Expanded && (node.Type == NodeInterface || node.Type == NodeNetwork) {
				node.Expanded = false
				result.Model.FlatList = flattenTree(result.Model.Tree)
			}
		}
	case ActionDelete:
		if len(result.Model.FlatList) > 0 && result.Model.Cursor < len(result.Model.FlatList) {
			node := result.Model.FlatList[result.Model.Cursor]
			// Show delete dialog
			result.Model.DeleteDialog = &DeleteDialog{
				Target:      node,
				CursorOnYes: false, // Start on Cancel (safer default)
			}
		}
	case ActionHelp:
		result.Model.ShowHelp = !result.Model.ShowHelp
	}

	return result
}

func handleDetailKeyMsg(m Model, key tea.KeyMsg) UpdateResult {
	result := UpdateResult{Model: m}

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
	switch node.Type {
	case NodeScan:
		if node.Path != "" {
			_ = history.Delete(node.Path)
		}
	case NodeNetwork:
		for _, child := range node.Children {
			if child.Type == NodeScan && child.Path != "" {
				_ = history.Delete(child.Path)
			}
		}
	case NodeInterface:
		for _, netNode := range node.Children {
			for _, scanNode := range netNode.Children {
				if scanNode.Type == NodeScan && scanNode.Path != "" {
					_ = history.Delete(scanNode.Path)
				}
			}
		}
	}
}

type treeLoadedMsg struct {
	tree []*TreeNode
}

func loadTreeCmd() tea.Cmd {
	return func() tea.Msg {
		tree, err := buildHistoryTree()
		if err != nil {
			return treeLoadedMsg{tree: []*TreeNode{}}
		}
		return treeLoadedMsg{tree: tree}
	}
}


func buildHistoryTree() ([]*TreeNode, error) {
	baseDir, err := history.HistoryDir()
	if err != nil {
		return nil, err
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
		return nil, err
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

	return tree, nil
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

// restoreExpandedState restores the expanded state to matching nodes
func restoreExpandedState(tree []*TreeNode, state map[string]bool) {
	for _, node := range tree {
		if state[node.Name] {
			node.Expanded = true
		}
		// Recursively restore children
		restoreExpandedState(node.Children, state)
	}
}

func filterTree(tree []*TreeNode, filter string) []*TreeNode {
	if filter == "" {
		return flattenTree(tree)
	}

	filter = strings.ToLower(filter)
	var filtered []*TreeNode

	for _, node := range tree {
		if matchesFilter(node, filter) {
			filtered = append(filtered, node)
			if node.Expanded {
				filtered = append(filtered, filterTree(node.Children, filter)...)
			}
		}
	}

	return filtered
}

func matchesFilter(node *TreeNode, filter string) bool {
	// Match against node name
	if strings.Contains(strings.ToLower(node.Name), filter) {
		return true
	}

	// For scan nodes, also match against metadata
	if node.Type == NodeScan && node.ScanData != nil {
		// Match against CIDR
		if strings.Contains(strings.ToLower(node.ScanData.ScanMetadata.TargetCIDR), filter) {
			return true
		}
		// Match against interface name
		if strings.Contains(strings.ToLower(node.ScanData.ScanMetadata.InterfaceName), filter) {
			return true
		}
	}

	return false
}

