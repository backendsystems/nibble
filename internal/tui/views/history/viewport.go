package historyview

import (
	"charm.land/bubbles/v2/viewport"
)

// SetViewportSize initializes or updates the main viewport for list view
// accounting for title, spacing, and help text that appear outside the viewport
func (m Model) SetListViewportSize(windowWidth, windowHeight int) Model {
	// Preserve existing viewport state (especially YOffset) instead of reinitializing every update.
	if m.Viewport.Width() == 0 && m.Viewport.Height() == 0 {
		m.Viewport = viewport.New(viewport.WithWidth(windowWidth))
	}

	if windowWidth > 0 {
		m.Viewport.SetWidth(windowWidth)
	}

	if windowHeight > 0 {
		// Reserve space for:
		// - Title line (1)
		// - Spacing after title (1)
		// - Help text at bottom (1)
		// - Buffer (1)
		// - Error message if present (2)
		// Total reserved: 4 lines (6 with error)
		reservedHeight := 4
		if m.ErrorMsg != "" {
			reservedHeight += 2
		}
		viewportHeight := windowHeight - reservedHeight
		if viewportHeight < 3 {
			// Minimum 3 lines for viewport content
			viewportHeight = 3
		}
		m.Viewport.SetHeight(viewportHeight)
	}

	return m
}

// updateViewportContent keeps list viewport scroll/cursor state in sync.
// List rows are rendered on demand in render.go to avoid rebuilding full list content here.
func updateViewportContent(m Model) Model {
	m = m.SetListViewportSize(m.WindowW, m.WindowH)

	maxOffset := len(m.FlatList) - m.Viewport.Height()
	if maxOffset < 0 {
		maxOffset = 0
	}

	// Keep cursor visible by scrolling viewport
	cursorLine := m.Cursor
	if cursorLine < m.ListOffset {
		m.ListOffset = cursorLine
	} else if cursorLine >= m.ListOffset+m.Viewport.Height() {
		m.ListOffset = cursorLine - m.Viewport.Height() + 1
	}

	if m.ListOffset < 0 {
		m.ListOffset = 0
	}
	if m.ListOffset > maxOffset {
		m.ListOffset = maxOffset
	}

	return m
}
