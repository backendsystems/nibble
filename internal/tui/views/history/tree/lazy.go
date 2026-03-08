package tree

import (
	"context"

	"github.com/backendsystems/nibble/internal/history"
	tea "github.com/charmbracelet/bubbletea"
)

// CancelLoads cancels any in-flight background count loads for a node and its children.
func CancelLoads(node *Node) {
	if node == nil {
		return
	}
	if node.CancelLoad != nil {
		node.CancelLoad()
		node.CancelLoad = nil
	}
	for _, child := range node.Children {
		CancelLoads(child)
	}
}

// LoadCountsForExpandedNodes collects scan children of expanded network nodes
// and schedules one background load per scan in visible order.
func LoadCountsForExpandedNodes(tree []*Node) tea.Cmd {
	var cmds []tea.Cmd
	for _, iface := range tree {
		if !iface.Expanded {
			continue
		}
		for _, net := range iface.Children {
			if !net.Expanded {
				continue
			}
			cmds = append(cmds, LoadNetworkScanCountsCmd(net)...)
		}
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Sequence(cmds...)
}

func LoadNetworkScanCountsCmd(node *Node) []tea.Cmd {
	if node == nil || node.Type != NodeNetwork {
		return nil
	}

	var unloaded []*Node
	for _, child := range node.Children {
		if child != nil && child.Type == NodeScan && child.Counts == nil {
			unloaded = append(unloaded, child)
		}
	}
	if len(unloaded) == 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	node.CancelLoad = cancel

	cmds := make([]tea.Cmd, 0, len(unloaded))
	for _, child := range unloaded {
		cmds = append(cmds, loadScanCountCmd(ctx, child.Path))
	}
	return cmds
}

// ScanCountLoadedMsg carries host/port counts for a single scan path.
type ScanCountLoadedMsg struct {
	Path   string
	Counts ScanCounts
}

// loadScanCountCmd reads counts for one scan node in the background.
// It respects ctx: if cancelled (e.g. folder collapsed), the result is discarded.
func loadScanCountCmd(ctx context.Context, path string) tea.Cmd {
	return func() tea.Msg {
		if ctx.Err() != nil {
			return nil
		}
		summary, err := history.LoadSummary(path)
		if err != nil {
			return nil
		}
		if ctx.Err() != nil {
			return nil
		}
		return ScanCountLoadedMsg{
			Path: path,
			Counts: ScanCounts{
				Hosts: summary.ScanResults.HostsFound,
				Ports: summary.ScanResults.PortsFound,
			},
		}
	}
}

func ApplyScanCountLoadedMsg(flatList []*Node, msg ScanCountLoadedMsg) {
	for _, node := range flatList {
		if node == nil || node.Path != msg.Path {
			continue
		}
		counts := msg.Counts
		node.Counts = &counts
		return
	}
}
