package historydetailview

import (
	"fmt"
	"strings"

	"github.com/backendsystems/nibble/internal/tui/views/common"
)

func Render(m Model, windowWidth, windowHeight int) string {
	var b strings.Builder

	// Title (outside viewport)
	title := fmt.Sprintf("%s - %s - %s",
		m.History.ScanMetadata.TargetCIDR,
		m.History.ScanMetadata.InterfaceName,
		m.History.ScanMetadata.Created.Format("Jan 2 15:04"),
	)
	b.WriteString(common.TitleStyle.Render(title) + "\n\n")

	// Content for viewport
	var content strings.Builder

	// Metadata
	content.WriteString(fmt.Sprintf("Duration:     %.1fs\n", m.History.ScanMetadata.DurationSeconds))

	if len(m.History.ScanMetadata.PortsScanned) > 0 {
		portsStr := formatPorts(m.History.ScanMetadata.PortsScanned)
		content.WriteString(fmt.Sprintf("Ports:        %s\n", portsStr))
	}

	content.WriteString(fmt.Sprintf("Hosts found:  %d / %d\n",
		m.History.ScanResults.HostsFound,
		m.History.ScanResults.TotalHostsScanned,
	))

	if m.History.ScanResults.PortsFound > 0 {
		content.WriteString(fmt.Sprintf("Ports found:  %d\n", m.History.ScanResults.PortsFound))
	}

	if !m.History.ScanMetadata.Updated.Equal(m.History.ScanMetadata.Created) {
		content.WriteString(fmt.Sprintf("Updated:      %s\n", m.History.ScanMetadata.Updated.Format("Jan 2 15:04")))
	}

	content.WriteString("\n")

	// Hosts list
	if len(m.History.ScanResults.Hosts) == 0 {
		content.WriteString(common.MutedStyle.Render("No hosts found in this scan\n"))
	} else {
		for i, host := range m.History.ScanResults.Hosts {
			isSelected := i == m.Cursor
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
				content.WriteString(common.HighlightStyle.Render(hostLine) + "\n")
			} else {
				content.WriteString(common.InfoTextStyle.Render(hostLine) + "\n")
			}

			// Ports
			allPortsScanned := len(host.PortsScanned) == 65535
			for _, port := range host.Ports {
				portLine := "    port " + fmt.Sprintf("%d", port.Port)
				if port.Banner != "" {
					portLine += ": " + port.Banner
				}
				// Use green only if all ports were scanned
				if allPortsScanned {
					content.WriteString(common.ProgressGreenStyle.Render(portLine) + "\n")
				} else {
					content.WriteString(portLine + "\n")
				}
			}

			// Show if all ports were scanned
			if allPortsScanned {
				content.WriteString(common.MutedStyle.Render("    [All 65535 ports scanned]") + "\n")
			}

			content.WriteString("\n")
		}
	}

	// Update viewport with content and dimensions
	m = m.UpdateViewportContent(content.String(), windowWidth, windowHeight)

	// Build final output with viewport and help text
	b.WriteString(m.Viewport.View())
	b.WriteString("\n")
	b.WriteString(common.HelpTextStyle.Render("↑/↓: select host • Enter: scan all ports • q: back • ?: help"))

	if m.ErrorMsg != "" {
		b.WriteString("\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("Error: "+m.ErrorMsg))
	}

	view := b.String()
	if m.ShowHelp {
		return renderHelpOverlay(view, windowWidth)
	}
	return view
}

func formatPorts(ports []int) string {
	if len(ports) == 0 {
		return "none"
	}
	if len(ports) > 10000 {
		return "1-65535 (all ports)"
	}
	if len(ports) <= 10 {
		var portStrs []string
		for _, p := range ports {
			portStrs = append(portStrs, fmt.Sprintf("%d", p))
		}
		return strings.Join(portStrs, ", ")
	}
	return fmt.Sprintf("%d ports", len(ports))
}
