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

// updateViewportContent pre-renders the list view content into the viewport
func updateViewportContent(m Model) Model {
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

	// Keep cursor visible by scrolling viewport
	cursorLine := m.Cursor
	if cursorLine < m.Viewport.YOffset {
		// Cursor is above viewport, scroll up to show it
		m.Viewport.YOffset = 0
	} else if cursorLine >= m.Viewport.YOffset+m.Viewport.Height {
		// Cursor is below viewport, scroll down to keep it visible
		m.Viewport.YOffset = cursorLine - m.Viewport.Height + 1
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
	var style lipgloss.Style

	switch node.Type {
	case NodeInterface:
		if node.Expanded {
			icon = "📂"
		} else {
			icon = "📁"
		}
		name = node.Name
		style = getFolderStyle()
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
		style = getFolderStyle()
		if len(node.Children) > 0 {
			suffix := "scans"
			if len(node.Children) == 1 {
				suffix = "scan"
			}
			name += fmt.Sprintf(" (%d %s)", len(node.Children), suffix)
		}

	case NodeScan:
		icon = "📄"
		style = getScanStyle()
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
	}

	line := indent + cursor + icon + " " + name
	if isSelected {
		b.WriteString(getSelectedStyle().Render(line) + "\n")
	} else {
		b.WriteString(style.Render(line) + "\n")
	}
}

// getSelectedStyle returns the style for selected items
func getSelectedStyle() lipgloss.Style {
	return common.HighlightStyle
}

// getFolderStyle returns the style for folder nodes
func getFolderStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
}

// getScanStyle returns the style for scan nodes
func getScanStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(common.Color.Info)
}
