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

	// Add viewport content (already set in Update)
	b.WriteString(m.Viewport.View())
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓/←/→ • Enter • Del: delete • ?: help • q: back"))

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
