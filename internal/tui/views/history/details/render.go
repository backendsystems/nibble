package historydetailview

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/backendsystems/nibble/internal/tui/views/common"
)

func Render(m Model, windowWidth, windowHeight int) string {
	var b strings.Builder

	// Title (outside viewport)
	cidr := m.History.ScanMetadata.TargetCIDR
	iface := m.History.ScanMetadata.InterfaceName
	date := m.History.ScanMetadata.Created.Format("2006 Jan 2 15:04")

	var counter string
	if total := len(m.History.ScanResults.Hosts); total > 0 {
		counter = common.MutedStyle.Render(fmt.Sprintf("%d/%d", m.Cursor+1, total))
	}

	titleFull := common.TitleStyle.Render(fmt.Sprintf("%s - %s - %s", cidr, iface, date))
	titleShort := common.TitleStyle.Render(fmt.Sprintf("%s - %s", cidr, iface))

	// Wrap date to second line when full title + counter won't fit.
	// Always show cidr - iface and counter regardless of width.
	counterW := lipgloss.Width(counter)
	needsWrap := lipgloss.Width(titleFull)+counterW+1 > windowWidth

	titleLine := titleFull
	if needsWrap {
		titleLine = titleShort
	}
	if counter != "" {
		gap := windowWidth - lipgloss.Width(titleLine) - counterW
		if gap > 0 {
			titleLine += strings.Repeat(" ", gap) + counter
		} else {
			titleLine += " " + counter
		}
	}
	b.WriteString(titleLine + "\n")
	if needsWrap {
		b.WriteString(common.TitleStyle.Render(date) + "\n")
	}

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

	// Compute how many lines the title and help areas occupy at this width.
	titleLines := 1
	if needsWrap {
		titleLines = 2
	}
	const helpText = "↑/↓: select host • Enter/→: scan all ports • ←/q: back • ?: help"
	helpWrapped := common.WrapWords(helpText, windowWidth)
	helpLines := strings.Count(helpWrapped, "\n") + 1 // +1 blank line before help
	m = m.UpdateViewportContent(content.String(), windowWidth, windowHeight, titleLines+helpLines)

	// Only auto-scroll during active scanning (to follow new ports).
	// Normal cursor-driven scrolling is handled by the controller via ScrollToSelected.
	if m.Scanning {
		m = m.ScrollToSelected()
	}

	// Build final output with viewport and help text
	b.WriteString(m.Viewport.View())
	b.WriteString("\n")
	b.WriteString(common.HelpTextStyle.Render(helpWrapped))

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
