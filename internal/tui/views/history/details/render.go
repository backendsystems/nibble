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
	b.WriteString(common.TitleStyle.Render(title) + "\n")

	// Content for viewport
	var content strings.Builder

	// Metadata
	// content.WriteString(fmt.Sprintf("Duration:     %.1fs\n", m.History.ScanMetadata.DurationSeconds))

	if m.History.ScanResults.PortsFound > 0 {
		content.WriteString(fmt.Sprintf("Ports found:  %d\n", m.History.ScanResults.PortsFound))
	}

	if !m.History.ScanMetadata.Updated.Equal(m.History.ScanMetadata.Created) {
		content.WriteString(fmt.Sprintf("Updated:      %s\n", m.History.ScanMetadata.Updated.Format("2006 Jan 2 15:04")))
	} else {
		content.WriteString(fmt.Sprintf("Created:      %s\n", m.History.ScanMetadata.Created.Format("2006 Jan 2 15:04")))
	}

	// content.WriteString("\n")

	// Hosts list
	if len(m.History.ScanResults.Hosts) == 0 {
		content.WriteString(common.MutedStyle.Render("No hosts found in this scan\n"))
	} else {
		for i, host := range m.History.ScanResults.Hosts {
			isSelected := i == m.Cursor
			allPortsScanned := len(host.PortsScanned) == 65535
			cursor := "  "
			if isSelected && !m.Scanning {
				cursor = "▶ "
			}

			// Host line
			hostLine := cursor + host.IP
			if host.Hardware != "" {
				hostLine += " - " + host.Hardware
			}
			if allPortsScanned {
				hostLine += " " + common.ProgressGreenStyle.Render("✓")
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

				// Use green if all ports scanned or newly found, normal otherwise
				if allPortsScanned {
					content.WriteString(common.ProgressGreenStyle.Render(portLine) + "\n")
				} else if isNewPort {
					content.WriteString(common.ProgressGreenStyle.Render(portLine) + "\n")
				} else {
					content.WriteString(portLine + "\n")
				}
			}

			// content.WriteString("\n")
		}
	}

	// Update viewport with content and dimensions
	m = m.UpdateViewportContent(content.String(), windowWidth, windowHeight)

	// Keep selected host visible and reveal some of its ports without large jumps.
	if len(m.History.ScanResults.Hosts) > 0 && m.Cursor >= 0 && m.Cursor < len(m.History.ScanResults.Hosts) {
		hostStart := 0

		// Count currently rendered metadata lines.
		if m.History.ScanResults.PortsFound > 0 {
			hostStart++
		}
		hostStart++ // Created/Updated line

		// Count all hosts before the selected one.
		for i := 0; i < m.Cursor; i++ {
			host := m.History.ScanResults.Hosts[i]
			hostStart++                  // Host line
			hostStart += len(host.Ports) // Port lines
		}

		selectedHost := m.History.ScanResults.Hosts[m.Cursor]
		hostEnd := hostStart + len(selectedHost.Ports) // inclusive

		top := m.Viewport.YOffset
		bottom := m.Viewport.YOffset + m.Viewport.Height - 1

		// If selection is above viewport, bring host line to top.
		if hostStart < top {
			m.Viewport.YOffset = hostStart
		} else {
			// Reveal additional selected-host ports as the cursor moves down.
			revealPorts := m.Viewport.Height / 2
			if revealPorts < 1 {
				revealPorts = 1
			}
			if revealPorts > len(selectedHost.Ports) {
				revealPorts = len(selectedHost.Ports)
			}

			targetBottom := hostStart + revealPorts
			if targetBottom > hostEnd {
				targetBottom = hostEnd
			}

			// If selected host (or its target ports) is below viewport, scroll minimally.
			if hostStart > bottom || targetBottom > bottom {
				m.Viewport.YOffset = targetBottom - m.Viewport.Height + 1
			}
		}

		if m.Viewport.YOffset < 0 {
			m.Viewport.YOffset = 0
		}
	}

	// Build final output with viewport and help text
	b.WriteString(m.Viewport.View())
	b.WriteString("\n")
	b.WriteString(common.HelpTextStyle.Render(common.WrapWords("↑/↓: select host • Enter/→: scan all ports • ←/q: back • ?: help", windowWidth)))

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
