package historyview

import (
	"context"

	"github.com/backendsystems/nibble/internal/history"
	tea "github.com/charmbracelet/bubbletea"
)

// cancelNodeLoads cancels any in-flight background count loads for a node and its children.
func cancelNodeLoads(node *TreeNode) {
	if node == nil {
		return
	}
	if node.cancelLoad != nil {
		node.cancelLoad()
		node.cancelLoad = nil
	}
	for _, child := range node.Children {
		cancelNodeLoads(child)
	}
}

// loadCountsForExpandedNodes collects scan children of expanded network nodes
// and schedules one background load per scan in visible order.
func loadCountsForExpandedNodes(tree []*TreeNode) tea.Cmd {
	var cmds []tea.Cmd
	for _, iface := range tree {
		if !iface.Expanded {
			continue
		}
		for _, net := range iface.Children {
			if !net.Expanded {
				continue
			}
			cmds = append(cmds, loadNetworkScanCountsCmd(net)...)
		}
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Sequence(cmds...)
}

func loadNetworkScanCountsCmd(node *TreeNode) []tea.Cmd {
	if node == nil || node.Type != NodeNetwork {
		return nil
	}

	var unloaded []*TreeNode
	for _, child := range node.Children {
		if child != nil && child.Type == NodeScan && child.Counts == nil {
			unloaded = append(unloaded, child)
		}
	}
	if len(unloaded) == 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	node.cancelLoad = cancel

	cmds := make([]tea.Cmd, 0, len(unloaded))
	for _, child := range unloaded {
		cmds = append(cmds, loadScanCountCmd(ctx, child.Path))
	}
	return cmds
}

// scanCountLoadedMsg carries host/port counts for a single scan path.
type scanCountLoadedMsg struct {
	path   string
	counts ScanCounts
}

// loadScanCountCmd reads counts for one scan node in the background.
// It respects ctx: if cancelled (e.g. folder collapsed), the result is discarded.
func loadScanCountCmd(ctx context.Context, path string) tea.Cmd {
	return func() tea.Msg {
		if ctx.Err() != nil {
			return nil
		}
		scanData, err := history.Load(path)
		if err != nil {
			return nil
		}
		ports := 0
		for _, host := range scanData.ScanResults.Hosts {
			ports += len(host.Ports)
		}
		if ctx.Err() != nil {
			return nil
		}
		return scanCountLoadedMsg{
			path: path,
			counts: ScanCounts{
				Hosts: len(scanData.ScanResults.Hosts),
				Ports: ports,
			},
		}
	}
}

func applyScanCountLoadedMsg(flatList []*TreeNode, msg scanCountLoadedMsg) {
	for _, node := range flatList {
		if node == nil || node.Path != msg.path {
			continue
		}
		counts := msg.counts
		node.Counts = &counts
		return
	}
}
