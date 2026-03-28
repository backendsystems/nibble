package historydetailview

import (
	"charm.land/bubbles/v2/viewport"
)

// SetViewportSize initializes or updates the viewport with proper dimensions
// accounting for title, metadata, and help text that appear outside the viewport
func (m Model) SetViewportSize(windowWidth, windowHeight int) Model {
	m.Viewport = viewport.New(viewport.WithWidth(windowWidth))

	if windowWidth > 0 {
		m.Viewport.SetWidth(windowWidth)
	}

	if windowHeight > 0 {
		// Reserve space for:
		// - Title line (1)
		// - Help text at bottom (1)
		// Total reserved: 3 lines
		reservedHeight := 3
		viewportHeight := max(windowHeight-reservedHeight,
			// Minimum 3 lines for viewport content
			3)
		m.Viewport.SetHeight(viewportHeight)
	}

	return m
}

// UpdateViewportContent updates the viewport with new content and ensures proper dimensions.
// reservedLines is the total number of non-viewport lines (title + help, accounting for wrapping).
func (m Model) UpdateViewportContent(content string, windowWidth, windowHeight, reservedLines int) Model {
	m.Viewport.SetContent(content)

	if windowWidth > 0 {
		m.Viewport.SetWidth(windowWidth)
	}

	if windowHeight > 0 {
		viewportHeight := max(windowHeight-reservedLines, 3)
		m.Viewport.SetHeight(viewportHeight)
	}

	return m
}
