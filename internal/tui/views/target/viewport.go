package targetview

import "charm.land/bubbles/v2/viewport"

// UpdateViewport resizes the viewport to fit the current WindowH and width.
func (m *Model) UpdateViewport(maxWidth int) {
	// title (1) + blank+helpline (2) = 3 reserved lines
	const reserved = 3
	h := max(m.WindowH-reserved, 1)
	if m.Viewport.Height() == 0 {
		m.Viewport = viewport.New(viewport.WithHeight(h), viewport.WithWidth(maxWidth))
	} else {
		m.Viewport.SetHeight(h)
		m.Viewport.SetWidth(maxWidth)
	}
}

// scrollToFocused scrolls the viewport so the focused field is fully visible.
func (m *Model) scrollToFocused() {
	h := m.Viewport.Height()
	if h == 0 {
		return
	}
	fieldStart := 0
	for i := range m.FocusedField {
		fieldStart += fieldHeights[i]
	}
	fieldEnd := fieldStart + fieldHeights[m.FocusedField] - 1

	// Scroll down to show field bottom, then up to show field top (top wins if field > viewport)
	if fieldEnd >= m.Viewport.YOffset()+h {
		m.Viewport.SetYOffset(fieldEnd - h + 1)
	}
	if fieldStart < m.Viewport.YOffset() {
		m.Viewport.SetYOffset(fieldStart)
	}
}
