package historydetailview

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)
	normalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	mutedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	portStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("150"))
)

func Render(m Model, windowWidth, windowHeight int) string {
	var b strings.Builder

	// Title (outside viewport)
	title := fmt.Sprintf("%s - %s - %s",
		m.History.ScanMetadata.TargetCIDR,
		m.History.ScanMetadata.InterfaceName,
		m.History.ScanMetadata.Created.Format("Jan 2 15:04"),
	)
	b.WriteString(titleStyle.Render(title) + "\n\n")

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
		content.WriteString(mutedStyle.Render("No hosts found in this scan\n"))
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
				content.WriteString(selectedStyle.Render(hostLine) + "\n")
			} else {
				content.WriteString(normalStyle.Render(hostLine) + "\n")
			}

			// Ports
			for _, port := range host.Ports {
				portLine := "    port " + fmt.Sprintf("%d", port.Port)
				if port.Banner != "" {
					portLine += ": " + port.Banner
				}
				content.WriteString(portStyle.Render(portLine) + "\n")
			}

			// Show if all ports were scanned
			if len(host.PortsScanned) > 10000 {
				content.WriteString(mutedStyle.Render("    [All 65535 ports scanned]") + "\n")
			}

			content.WriteString("\n")
		}
	}

	// Update viewport with content and dimensions
	m = m.UpdateViewportContent(content.String(), windowWidth, windowHeight)

	// Build final output with viewport and help text
	b.WriteString(m.Viewport.View())
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓: select host • Enter: scan all ports • q: back • ?: help"))

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
