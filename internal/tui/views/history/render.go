package historyview

import (
	"fmt"
	"strings"

	"github.com/backendsystems/nibble/internal/tui/views/common"
	detailsview "github.com/backendsystems/nibble/internal/tui/views/history/details"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle    = common.TitleStyle
	selectedStyle = common.HighlightStyle
	normalStyle   = lipgloss.NewStyle().Foreground(common.Color.Info)
	mutedStyle    = common.HelpTextStyle
	helpStyle     = common.HelpTextStyle
	folderStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
	scanStyle     = lipgloss.NewStyle().Foreground(common.Color.Info)
)

func Render(m Model, maxWidth int) string {
	if m.Mode == ViewDetail {
		return detailsview.Render(m.Details, maxWidth, m.WindowH)
	}
	return renderList(m, maxWidth)
}

func renderList(m Model, maxWidth int) string {
	var b strings.Builder

	// Title (outside viewport)
	b.WriteString(titleStyle.Render("Scan History") + "\n\n")

	// Render only the visible rows instead of a fully pre-rendered viewport buffer.
	b.WriteString(renderVisibleList(m))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓/←/→ • Del: delete • ?: help • q: back"))

	if m.ErrorMsg != "" {
		b.WriteString("\n\n" + common.ErrorStyle.Render("Error: "+m.ErrorMsg))
	}

	view := b.String()

	// Show overlays (help takes precedence over delete dialog)
	if m.ShowHelp {
		return renderListHelpOverlay(view, m.WindowW, m.WindowH)
	}

	if m.DeleteDialog != nil {
		return m.DeleteDialog.Render(view, m.WindowW, m.WindowH)
	}

	return view
}

func renderVisibleList(m Model) string {
	if len(m.Tree) == 0 {
		return "No scan history found\n"
	}

	start := m.Viewport.YOffset
	if start < 0 {
		start = 0
	}
	if start > len(m.FlatList) {
		start = len(m.FlatList)
	}

	end := start + m.Viewport.Height
	if end > len(m.FlatList) {
		end = len(m.FlatList)
	}

	var b strings.Builder
	for i := start; i < end; i++ {
		node := m.FlatList[i]
		if node == nil {
			continue
		}
		renderNode(&b, node, i == m.Cursor)
	}
	return b.String()
}

func renderNode(b *strings.Builder, node *TreeNode, isSelected bool) {
	indent := strings.Repeat("  ", node.Level)
	cursor := "  "
	if isSelected {
		cursor = "▶ "
	}

	var icon string
	var name string
	var style lipgloss.Style

	switch node.Type {
	case NodeInterface:
		if node.Expanded {
			icon = "📂"
		} else {
			icon = "📁"
		}
		name = node.Name
		style = folderStyle
		if len(node.Children) > 0 {
			suffix := "networks"
			if len(node.Children) == 1 {
				suffix = "network"
			}
			name += fmt.Sprintf(" (%d %s)", len(node.Children), suffix)
		}

	case NodeNetwork:
		if node.Expanded {
			icon = "📂"
		} else {
			icon = "📁"
		}
		name = node.Name
		style = folderStyle
		if len(node.Children) > 0 {
			suffix := "scans"
			if len(node.Children) == 1 {
				suffix = "scan"
			}
			name += fmt.Sprintf(" (%d %s)", len(node.Children), suffix)
		}

	case NodeScan:
		icon = "📄"
		if node.ScanData != nil {
			hostCount := node.ScanData.ScanResults.HostsFound
			hostSuffix := "hosts"
			if hostCount == 1 {
				hostSuffix = "host"
			}
			portCount := node.ScanData.ScanResults.PortsFound
			portSuffix := "ports"
			if portCount == 1 {
				portSuffix = "port"
			}
			name = fmt.Sprintf("%s (%d %s, %d %s)",
				node.Name,
				hostCount,
				hostSuffix,
				portCount,
				portSuffix,
			)
		} else {
			name = node.Name
		}
		style = scanStyle
	}

	line := indent + cursor + icon + " " + name

	if isSelected {
		b.WriteString(selectedStyle.Render(line) + "\n")
	} else {
		b.WriteString(style.Render(line) + "\n")
	}
}
