package historyview

import (
	"fmt"
	"strings"

	"github.com/backendsystems/nibble/internal/tui/views/common"
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
		return renderDetail(m, maxWidth)
	}
	return renderList(m, maxWidth)
}

func renderList(m Model, maxWidth int) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Scan History") + "\n\n")

	if len(m.Tree) == 0 {
		b.WriteString(mutedStyle.Render("No scan history found\n"))
	} else {
		for i, node := range m.FlatList {
			isSelected := i == m.Cursor
			renderNode(&b, node, isSelected)
		}
	}

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

func renderDetail(m Model, maxWidth int) string {
	var b strings.Builder

	if m.DetailHistory == nil {
		return "No scan selected"
	}

	h := m.DetailHistory

	// Title
	title := fmt.Sprintf("%s - %s - %s",
		h.ScanMetadata.TargetCIDR,
		h.ScanMetadata.InterfaceName,
		h.ScanMetadata.Created.Format("2006 Jan 2 15:04"),
	)
	b.WriteString(titleStyle.Render(title) + "\n\n")

	// Metadata - only duration, updated (if different), and ports scanned (last)
	b.WriteString(fmt.Sprintf("Duration:     %.1fs\n", h.ScanMetadata.DurationSeconds))

	if !h.ScanMetadata.Updated.Equal(h.ScanMetadata.Created) {
		b.WriteString(fmt.Sprintf("Updated:      %s\n", h.ScanMetadata.Updated.Format("2006 Jan 2 15:04")))
	}

	if len(h.ScanMetadata.PortsScanned) > 0 {
		portsStr := formatDetailPorts(h.ScanMetadata.PortsScanned)
		b.WriteString(fmt.Sprintf("Ports:        %s\n", portsStr))
	}

	b.WriteString("\n")

	// Hosts list
	if len(h.ScanResults.Hosts) == 0 {
		b.WriteString(mutedStyle.Render("No hosts found in this scan\n"))
	} else {
		for i, host := range h.ScanResults.Hosts {
			isSelected := i == m.DetailCursor
			cursor := "  "
			if isSelected {
				cursor = "▶ "
			}

			// Host line
			hostLine := cursor + host.IP
			if host.Hardware != "" {
				hostLine += " - " + host.Hardware
			}

			if isSelected {
				b.WriteString(selectedStyle.Render(hostLine) + "\n")
			} else {
				b.WriteString(normalStyle.Render(hostLine) + "\n")
			}

			// Ports
			portStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("150"))
			for _, port := range host.Ports {
				portLine := "    port " + fmt.Sprintf("%d", port.Port)
				if port.Banner != "" {
					portLine += ": " + port.Banner
				}
				b.WriteString(portStyle.Render(portLine) + "\n")
			}

			// Show if all ports were scanned
			if len(host.PortsScanned) > 10000 {
				b.WriteString(mutedStyle.Render("    [All 65535 ports scanned]") + "\n")
			}

			b.WriteString("\n")
		}
	}

	b.WriteString(helpStyle.Render("↑/↓: select host • Enter: scan all ports • ?: help • q: back"))

	view := b.String()

	// Show overlays (help takes precedence over delete dialog)
	if m.ShowHelp {
		return renderDetailHelpOverlay(view, m.WindowW, m.WindowH)
	}

	if m.DeleteDialog != nil {
		return m.DeleteDialog.Render(view, m.WindowW, m.WindowH)
	}

	return view
}

func formatDetailPorts(ports []int) string {
	if len(ports) == 0 {
		return "none"
	}
	if len(ports) > 10000 {
		return "1-65535 (all ports)"
	}

	// Show actual ports with ranges
	ranges := buildPortRanges(ports)
	return strings.Join(ranges, ", ")
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
			name += fmt.Sprintf(" (%d networks)", len(node.Children))
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
			name += fmt.Sprintf(" (%d scans)", len(node.Children))
		}

	case NodeScan:
		icon = "📄"
		if node.ScanData != nil {
			name = fmt.Sprintf("%s (%d hosts)",
				node.Name,
				node.ScanData.ScanResults.HostsFound,
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

func buildPortRanges(ports []int) []string {
	if len(ports) == 0 {
		return nil
	}

	// Sort ports first (should already be sorted from config, but just in case)
	sorted := make([]int, len(ports))
	copy(sorted, ports)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	var ranges []string
	start := sorted[0]
	end := sorted[0]

	for i := 1; i < len(sorted); i++ {
		if sorted[i] == end+1 {
			// Continue the range
			end = sorted[i]
		} else {
			// End current range and start new one
			if start == end {
				ranges = append(ranges, fmt.Sprintf("%d", start))
			} else if end == start+1 {
				// Two consecutive ports, don't use range notation
				ranges = append(ranges, fmt.Sprintf("%d", start))
				ranges = append(ranges, fmt.Sprintf("%d", end))
			} else {
				ranges = append(ranges, fmt.Sprintf("%d-%d", start, end))
			}
			start = sorted[i]
			end = sorted[i]
		}
	}

	// Add final range
	if start == end {
		ranges = append(ranges, fmt.Sprintf("%d", start))
	} else if end == start+1 {
		ranges = append(ranges, fmt.Sprintf("%d", start))
		ranges = append(ranges, fmt.Sprintf("%d", end))
	} else {
		ranges = append(ranges, fmt.Sprintf("%d-%d", start, end))
	}

	return ranges
}
