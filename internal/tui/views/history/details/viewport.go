package historydetailview

import (
	"github.com/charmbracelet/bubbles/viewport"
)

// SetViewportSize initializes or updates the viewport with proper dimensions
// accounting for title, metadata, and help text that appear outside the viewport
func (m Model) SetViewportSize(windowWidth, windowHeight int) Model {
	m.Viewport = viewport.New(windowWidth, 0)

	if windowWidth > 0 {
		m.Viewport.Width = windowWidth
	}

	if windowHeight > 0 {
		// Reserve space for:
		// - Title line (1)
		// - Help text at bottom (1)
		// Total reserved: 3 lines
		reservedHeight := 3
		viewportHeight := windowHeight - reservedHeight
		if viewportHeight < 3 {
			// Minimum 3 lines for viewport content
			viewportHeight = 3
		}
		m.Viewport.Height = viewportHeight
	}

	return m
}

// UpdateViewportContent updates the viewport with new content and ensures proper dimensions
func (m Model) UpdateViewportContent(content string, windowWidth, windowHeight int) Model {
	m.Viewport.SetContent(content)

	if windowWidth > 0 {
		m.Viewport.Width = windowWidth
	}

	if windowHeight > 0 {
		// Reserve space for title and help text
		reservedHeight := 3
		viewportHeight := windowHeight - reservedHeight
		if viewportHeight < 3 {
			viewportHeight = 3
		}
		m.Viewport.Height = viewportHeight
	}

	return m
}
