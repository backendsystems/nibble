package historyview

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/backendsystems/nibble/internal/history"
)

// buildHistoryTree builds the tree structure from history files
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

		// Parse path: baseDir/interface/network/file.json
		relPath, _ := filepath.Rel(baseDir, path)
		parts := strings.Split(relPath, string(filepath.Separator))
		if len(parts) < 3 {
			return nil
		}

		interfaceName := parts[0]
		networkName := parts[1]

		// Parse timestamp from filename (scan_20060102_150405.json) — no file read needed.
		base := filepath.Base(path)
		created, parseErr := time.ParseInLocation("scan_20060102_150405.json", base, time.Local)
		if parseErr != nil {
			return nil
		}

		// Create scan node — ScanData is nil until the user opens the scan (lazy loaded).
		scanNode := &TreeNode{
			Type:    NodeScan,
			Name:    created.Format("2006 Jan 2 15:04"),
			Path:    path,
			Created: created,
			Level:   2,
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
				return scans[i].Created.After(scans[j].Created)
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

// flattenTree converts a tree to a flattened list based on expansion state
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
