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
		m.History.ScanMetadata.Created.Format("2006 Jan 2 15:04"),
	)
	b.WriteString(common.TitleStyle.Render(title) + "\n\n")

	// Content for viewport
	var content strings.Builder

	// Metadata
	content.WriteString(fmt.Sprintf("Duration:     %.1fs\n", m.History.ScanMetadata.DurationSeconds))
	portsStr := formatPorts(m.History.ScanMetadata.PortsScanned)
	content.WriteString(fmt.Sprintf("Ports:        %s\n", portsStr))

	content.WriteString(fmt.Sprintf("Hosts found:  %d / %d\n",
		m.History.ScanResults.HostsFound,
		m.History.ScanResults.TotalHostsScanned,
	))

	if m.History.ScanResults.PortsFound > 0 {
		content.WriteString(fmt.Sprintf("Ports found:  %d\n", m.History.ScanResults.PortsFound))
	}

	if !m.History.ScanMetadata.Updated.Equal(m.History.ScanMetadata.Created) {
		content.WriteString(fmt.Sprintf("Updated:      %s\n", m.History.ScanMetadata.Updated.Format("2006 Jan 2 15:04")))
	}

	content.WriteString("\n")

	// Hosts list
	if len(m.History.ScanResults.Hosts) == 0 {
		content.WriteString(common.MutedStyle.Render("No hosts found in this scan\n"))
	} else {
		for i, host := range m.History.ScanResults.Hosts {
			isSelected := i == m.Cursor
			cursor := "  "
			if isSelected && !m.Scanning {
				cursor = "▶ "
			}

			// Host line
			hostLine := cursor + host.IP
			if host.Hardware != "" {
				hostLine += " - " + host.Hardware
			}

			// Show spinner only for the host being scanned
			if i == m.ScanningHostIdx && m.Scanning {
				spinnerChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
				spinnerIdx := int(m.Stopwatch.Elapsed().Milliseconds()/100) % len(spinnerChars)
				hostLine += " " + common.ProgressGreenStyle.Render(spinnerChars[spinnerIdx])
			} else if i == m.ScanningHostIdx && !m.Scanning && m.ProgressChan != nil {
				// Show checkmark when that host's scan completes
				hostLine += " " + common.ProgressGreenStyle.Render("✓")
			}

			if isSelected {
				if m.Scanning {
					// Use warning color (orange) when scanning to indicate locked state
					content.WriteString(common.WarningStyle.Render(hostLine) + "\n")
				} else {
					// Normal yellow highlight when not scanning
					content.WriteString(common.HighlightStyle.Render(hostLine) + "\n")
				}
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

				// Check if this port is newly found
				isNewPort := false
				if m.NewPortsByHost != nil && m.NewPortsByHost[host.IP] != nil {
					isNewPort = m.NewPortsByHost[host.IP][port.Port]
				}

				// Use green if all ports scanned, yellow if newly found, normal otherwise
				if allPortsScanned {
					content.WriteString(common.ProgressGreenStyle.Render(portLine) + "\n")
				} else if isNewPort {
					// Yellow for newly found ports
					content.WriteString(common.HighlightStyle.Render(portLine) + "\n")
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

	// Keep selected host visible by scrolling viewport
	// Count lines to selected host
	lineToHost := 0

	// Count metadata lines
	lineToHost++ // Duration line
	lineToHost++ // Ports line
	lineToHost++ // Hosts found line
	if m.History.ScanResults.PortsFound > 0 {
		lineToHost++ // Ports found line
	}
	if !m.History.ScanMetadata.Updated.Equal(m.History.ScanMetadata.Created) {
		lineToHost++ // Updated line
	}
	lineToHost++ // Blank line after metadata

	// Count all hosts before the selected one
	for i := 0; i < m.Cursor; i++ {
		host := m.History.ScanResults.Hosts[i]
		lineToHost++                  // Host line
		lineToHost += len(host.Ports) // Port lines
		// Check if "all ports scanned" message is shown
		if len(host.PortsScanned) == 65535 {
			lineToHost++
		}
		lineToHost++ // Blank line after host
	}

	// Keep host visible, scrolling based on number of ports to show them
	if lineToHost < m.Viewport.YOffset {
		// Host is above viewport, scroll up to show metadata and host
		m.Viewport.YOffset = 0
	} else if lineToHost >= m.Viewport.YOffset+m.Viewport.Height-1 {
		// Host is at or past bottom, scroll to keep it visible with ports below
		selectedHost := m.History.ScanResults.Hosts[m.Cursor]
		// Reserve space: 1 for host line + 2 buffer lines
		portLines := len(selectedHost.Ports)
		if len(selectedHost.PortsScanned) == 65535 {
			portLines++ // "All ports" message
		}
		reserveLines := 1 + 2 // host line + buffer
		m.Viewport.YOffset = lineToHost - m.Viewport.Height + reserveLines + (portLines / 2)
		if m.Viewport.YOffset < 0 {
			m.Viewport.YOffset = 0
		}
	}

	// Build final output with viewport and help text
	b.WriteString(m.Viewport.View())
	b.WriteString("\n")
	b.WriteString(common.HelpTextStyle.Render("↑/↓: select host • Enter: scan all ports • q: back • ?: help"))

	if m.ErrorMsg != "" {
		b.WriteString("\n\n" + common.ErrorStyle.Render("Error: "+m.ErrorMsg))
	}

	view := b.String()
	if m.DeleteDialog != nil {
		return m.DeleteDialog.Render(view, windowWidth, windowHeight)
	}
	if m.ShowHelp {
		return renderHelpOverlay(view, windowWidth)
	}
	return view
}

func formatPorts(ports []int) string {
	if len(ports) == 0 {
		return "none"
	}
	if len(ports) == 65535 {
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
