package historyview

import (
	"fmt"
	"strings"

	"github.com/backendsystems/nibble/internal/tui/views/common"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

// SetViewportSize initializes or updates the main viewport for list view
// accounting for title, spacing, and help text that appear outside the viewport
func (m Model) SetListViewportSize(windowWidth, windowHeight int) Model {
	m.Viewport = viewport.New(windowWidth, 0)

	if windowWidth > 0 {
		m.Viewport.Width = windowWidth
	}

	if windowHeight > 0 {
		// Reserve space for:
		// - Title line (1)
		// - Spacing after title (1)
		// - Help text at bottom (1)
		// - Buffer (1)
		// Total reserved: 4 lines
		reservedHeight := 4
		viewportHeight := windowHeight - reservedHeight
		if viewportHeight < 3 {
			// Minimum 3 lines for viewport content
			viewportHeight = 3
		}
		m.Viewport.Height = viewportHeight
	}

	return m
}

// SetDetailViewportSize initializes or updates the detail viewport
// accounting for title, metadata, and help text that appear outside the viewport
func (m Model) SetDetailViewportSize(windowWidth, windowHeight int) Model {
	m.DetailViewport = viewport.New(windowWidth, 0)

	if windowWidth > 0 {
		m.DetailViewport.Width = windowWidth
	}

	if windowHeight > 0 {
		// Reserve space for:
		// - Title line (1)
		// - Spacing after title (1)
		// - Help text at bottom (1)
		// - Buffer (1)
		// Total reserved: 4 lines
		reservedHeight := 4
		viewportHeight := windowHeight - reservedHeight
		if viewportHeight < 3 {
			// Minimum 3 lines for viewport content
			viewportHeight = 3
		}
		m.DetailViewport.Height = viewportHeight
	}

	return m
}

// UpdateViewportContent updates the main viewport with new content and dimensions
func (m Model) UpdateViewportContent(content string, windowWidth, windowHeight int) Model {
	// Initialize viewport if needed
	if m.Viewport.Width == 0 && windowWidth > 0 {
		m.Viewport = viewport.New(windowWidth, 0)
	}

	m.Viewport.SetContent(content)

	if windowWidth > 0 {
		m.Viewport.Width = windowWidth
	}

	if windowHeight > 0 {
		reservedHeight := 4
		viewportHeight := windowHeight - reservedHeight
		if viewportHeight < 3 {
			viewportHeight = 3
		}
		m.Viewport.Height = viewportHeight
	}

	return m
}

// UpdateDetailViewportContent updates the detail viewport with new content and dimensions
func (m Model) UpdateDetailViewportContent(content string, windowWidth, windowHeight int) Model {
	// Initialize viewport if needed
	if m.DetailViewport.Width == 0 && windowWidth > 0 {
		m.DetailViewport = viewport.New(windowWidth, 0)
	}

	m.DetailViewport.SetContent(content)

	if windowWidth > 0 {
		m.DetailViewport.Width = windowWidth
	}

	if windowHeight > 0 {
		reservedHeight := 4
		viewportHeight := windowHeight - reservedHeight
		if viewportHeight < 3 {
			viewportHeight = 3
		}
		m.DetailViewport.Height = viewportHeight
	}

	return m
}

// updateViewportContent pre-renders the appropriate content into the viewport
func updateViewportContent(m Model) Model {
	if m.Mode == ViewDetail {
		// Render detail view content
		if m.DetailHistory != nil {
			var content strings.Builder

			h := m.DetailHistory

			// Metadata
			content.WriteString(fmt.Sprintf("Duration:     %.1fs\n", h.ScanMetadata.DurationSeconds))

			if !h.ScanMetadata.Updated.Equal(h.ScanMetadata.Created) {
				content.WriteString(fmt.Sprintf("Updated:      %s\n", h.ScanMetadata.Updated.Format("2006 Jan 2 15:04")))
			}

			if len(h.ScanMetadata.PortsScanned) > 0 {
				portsStr := formatDetailPorts(h.ScanMetadata.PortsScanned)
				content.WriteString(fmt.Sprintf("Ports:        %s\n", portsStr))
			}

			content.WriteString("\n")

			// Hosts list
			if len(h.ScanResults.Hosts) == 0 {
				content.WriteString("No hosts found in this scan\n")
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
						content.WriteString(getSelectedStyle().Render(hostLine) + "\n")
					} else {
						content.WriteString(hostLine + "\n")
					}

					// Ports
					for _, port := range host.Ports {
						portLine := "    port " + fmt.Sprintf("%d", port.Port)
						if port.Banner != "" {
							portLine += ": " + port.Banner
						}
						content.WriteString(portLine + "\n")
					}

					// Show if all ports were scanned
					if len(host.PortsScanned) > 10000 {
						content.WriteString("    [All 65535 ports scanned]\n")
					}

					content.WriteString("\n")
				}
			}

			m = m.UpdateDetailViewportContent(content.String(), m.WindowW, m.WindowH)

			// Keep selected host visible by scrolling viewport
			// Count lines to selected host, accounting for all content
			lineToHost := 0

			// Count metadata lines (Duration, Updated, Ports)
			lineToHost++ // Duration line
			if !h.ScanMetadata.Updated.Equal(h.ScanMetadata.Created) {
				lineToHost++ // Updated line
			}
			if len(h.ScanMetadata.PortsScanned) > 0 {
				lineToHost++ // Ports line
			}
			lineToHost++ // Blank line after metadata

			// Count all hosts before the selected one
			for i := 0; i < m.DetailCursor; i++ {
				host := h.ScanResults.Hosts[i]
				lineToHost++ // Host line
				lineToHost += len(host.Ports) // Port lines
				// Check if "all ports scanned" message is shown
				if len(host.PortsScanned) > 10000 {
					lineToHost++
				}
				lineToHost++ // Blank line after host
			}

			// Ensure the selected host line is visible
			if lineToHost < m.DetailViewport.YOffset {
				m.DetailViewport.YOffset = lineToHost
			} else if lineToHost >= m.DetailViewport.YOffset+m.DetailViewport.Height {
				m.DetailViewport.YOffset = lineToHost - m.DetailViewport.Height + 1
			}
		}
	} else {
		// Render list view content
		var content strings.Builder
		if len(m.Tree) == 0 {
			content.WriteString("No scan history found\n")
		} else {
			for i, node := range m.FlatList {
				isSelected := i == m.Cursor
				renderNodeToString(&content, node, isSelected)
			}
		}

		m = m.UpdateViewportContent(content.String(), m.WindowW, m.WindowH)
	}

	return m
}

// renderNodeToString renders a node to a strings.Builder for viewport content
func renderNodeToString(b *strings.Builder, node *TreeNode, isSelected bool) {
	indent := strings.Repeat("  ", node.Level)
	cursor := "  "
	if isSelected {
		cursor = "▶ "
	}

	var icon string
	var name string

	switch node.Type {
	case NodeInterface:
		if node.Expanded {
			icon = "📂"
		} else {
			icon = "📁"
		}
		name = node.Name
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
	}

	line := indent + cursor + icon + " " + name
	if isSelected {
		b.WriteString(getSelectedStyle().Render(line) + "\n")
	} else {
		b.WriteString(line + "\n")
	}
}

// getSelectedStyle returns the style for selected items
func getSelectedStyle() lipgloss.Style {
	return common.HighlightStyle
}
