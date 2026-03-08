package tree

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/backendsystems/nibble/internal/history"
)

// Build constructs the history tree from on-disk scan files.
func Build() ([]*Node, error) {
	baseDir, err := history.HistoryDir()
	if err != nil {
		return nil, err
	}

	type networkData struct {
		scans []*Node
	}
	interfaceMap := make(map[string]map[string]*networkData)

	err = filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".json" {
			return nil
		}

		relPath, _ := filepath.Rel(baseDir, path)
		parts := strings.Split(relPath, string(filepath.Separator))
		if len(parts) < 3 {
			return nil
		}

		interfaceName := parts[0]
		networkName := parts[1]

		base := filepath.Base(path)
		created, parseErr := time.ParseInLocation("scan_20060102_150405.json", base, time.Local)
		if parseErr != nil {
			return nil
		}

		scanNode := &Node{
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
		interfaceMap[interfaceName][networkName].scans = append(interfaceMap[interfaceName][networkName].scans, scanNode)

		return nil
	})
	if err != nil {
		return nil, err
	}

	var result []*Node
	var interfaces []string
	for iface := range interfaceMap {
		interfaces = append(interfaces, iface)
	}
	sort.Strings(interfaces)

	for _, iface := range interfaces {
		ifaceNode := &Node{
			Type:     NodeInterface,
			Name:     iface,
			Path:     filepath.Join(baseDir, iface),
			Expanded: false,
			Level:    0,
		}

		var networks []string
		for net := range interfaceMap[iface] {
			networks = append(networks, net)
		}
		sort.Strings(networks)

		for _, net := range networks {
			scans := interfaceMap[iface][net].scans
			sort.Slice(scans, func(i, j int) bool {
				return scans[i].Created.After(scans[j].Created)
			})

			netNode := &Node{
				Type:     NodeNetwork,
				Name:     strings.ReplaceAll(net, "_", "/"),
				Path:     filepath.Join(baseDir, iface, net),
				Expanded: false,
				Level:    1,
				Children: scans,
			}

			ifaceNode.Children = append(ifaceNode.Children, netNode)
		}

		result = append(result, ifaceNode)
	}

	return result, nil
}

// Flatten converts an expanded tree into a visible list.
func Flatten(nodes []*Node) []*Node {
	var flat []*Node
	for _, node := range nodes {
		flat = append(flat, node)
		if node.Expanded {
			flat = append(flat, Flatten(node.Children)...)
		}
	}
	return flat
}

// FindCursorByPath finds the visible row for the given path.
func FindCursorByPath(flatList []*Node, path string) int {
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

func ExpandAncestorsForPath(nodes []*Node, path string) bool {
	for _, node := range nodes {
		if node.Path == path {
			return true
		}
		if ExpandAncestorsForPath(node.Children, path) {
			node.Expanded = true
			return true
		}
	}
	return false
}
