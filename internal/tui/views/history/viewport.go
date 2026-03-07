package historyview

import (
	"github.com/charmbracelet/bubbles/viewport"
)

// SetViewportSize initializes or updates the main viewport for list view
// accounting for title, spacing, and help text that appear outside the viewport
func (m Model) SetListViewportSize(windowWidth, windowHeight int) Model {
	// Preserve existing viewport state (especially YOffset) instead of reinitializing every update.
	if m.Viewport.Width == 0 && m.Viewport.Height == 0 {
		m.Viewport = viewport.New(windowWidth, 0)
	}

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

// updateViewportContent keeps list viewport scroll/cursor state in sync.
// List rows are rendered on demand in render.go to avoid rebuilding full list content here.
func updateViewportContent(m Model) Model {
	m = m.SetListViewportSize(m.WindowW, m.WindowH)

	// Keep cursor visible by scrolling viewport
	cursorLine := m.Cursor
	if cursorLine < m.Viewport.YOffset {
		// Cursor is above viewport, scroll up to show it
		m.Viewport.YOffset = cursorLine
	} else if cursorLine >= m.Viewport.YOffset+m.Viewport.Height {
		// Cursor is below viewport, scroll down to keep it visible
		m.Viewport.YOffset = cursorLine - m.Viewport.Height + 1
	}

	if m.Viewport.YOffset < 0 {
		m.Viewport.YOffset = 0
	}

	maxOffset := len(m.FlatList) - m.Viewport.Height
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.Viewport.YOffset > maxOffset {
		m.Viewport.YOffset = maxOffset
	}

	return m
}
