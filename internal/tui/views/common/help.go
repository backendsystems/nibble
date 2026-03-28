package common

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// HelpConfig defines the content and appearance of a help overlay
type HelpConfig struct {
	Title      string
	Content    []string
	Width      int // Width of help box (default 56)
	ViewWidth  int // Width of the view/terminal for centering
	ViewHeight int // Height of the view/terminal for placement
}

// RenderHelpOverlay renders a centered help overlay with consistent styling
func RenderHelpOverlay(view string, config HelpConfig) string {
	viewWidth := config.ViewWidth
	if viewWidth == 0 {
		viewWidth = lipgloss.Width(view)
	}

	viewHeight := config.ViewHeight
	if viewHeight == 0 {
		viewHeight = lipgloss.Height(view)
	}

	// Calculate help box width: use 80% of window width with min 46, max 80
	width := config.Width
	if width == 0 {
		width = min(80, max(46, int(float64(viewWidth)*0.8)))
	}

	titleRow := renderHelpTitle(config.Title, width-4) // -4 for border (2) + padding (2)
	content := append([]string{titleRow}, config.Content...)
	helpContent := strings.Join(content, "\n")

	helpOverlay := HelpBoxStyle.Width(width).Render(helpContent)
	return lipgloss.Place(
		viewWidth,
		viewHeight,
		lipgloss.Center,
		lipgloss.Top,
		helpOverlay,
		lipgloss.WithWhitespaceChars(" "),
	)
}

// renderHelpTitle creates a title row with icon
func renderHelpTitle(title string, width int) string {
	styledTitle := HelpTitleStyle.Render(title)
	icon := HelpIconStyle.Render("❓")
	gap := max(width-lipgloss.Width(styledTitle)-lipgloss.Width(icon), 1)
	spacer := strings.Repeat(" ", gap)
	return styledTitle + spacer + icon
}
