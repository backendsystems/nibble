package tree

import (
	"fmt"
	"strings"

	"github.com/backendsystems/nibble/internal/tui/views/common"
	"github.com/charmbracelet/lipgloss"
)

var (
	selectedStyle = common.HighlightStyle
	folderStyle   = common.FolderStyle
	scanStyle     = lipgloss.NewStyle().Foreground(common.Color.Info)
)

func RenderVisibleList(flatList []*Node, tree []*Node, cursor, offset, height int) string {
	if len(tree) == 0 {
		return "No scan history found\n"
	}

	start := offset
	if start < 0 {
		start = 0
	}
	if start > len(flatList) {
		start = len(flatList)
	}

	end := start + height
	if end > len(flatList) {
		end = len(flatList)
	}

	var b strings.Builder
	for i := start; i < end; i++ {
		node := flatList[i]
		if node == nil {
			continue
		}
		renderNode(&b, node, i == cursor)
	}
	return b.String()
}

func renderNode(b *strings.Builder, node *Node, isSelected bool) {
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
		name = withCount(node.Name, len(node.Children), "network", "networks")
		style = folderStyle
	case NodeNetwork:
		if node.Expanded {
			icon = "📂"
		} else {
			icon = "📁"
		}
		name = withCount(node.Name, len(node.Children), "scan", "scans")
		style = folderStyle
	case NodeScan:
		icon = "📄"
		name = node.Name
		if node.Counts != nil {
			name = fmt.Sprintf(
				"%s (%d %s, %d %s)",
				node.Name,
				node.Counts.Hosts,
				singularOrPlural(node.Counts.Hosts, "host", "hosts"),
				node.Counts.Ports,
				singularOrPlural(node.Counts.Ports, "port", "ports"),
			)
		}
		style = scanStyle
	}

	line := indent + cursor + icon + " " + name
	if isSelected {
		b.WriteString(selectedStyle.Render(line) + "\n")
		return
	}
	b.WriteString(style.Render(line) + "\n")
}

func withCount(name string, count int, singular, plural string) string {
	if count == 0 {
		return name
	}
	return fmt.Sprintf("%s (%d %s)", name, count, singularOrPlural(count, singular, plural))
}

func singularOrPlural(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}
