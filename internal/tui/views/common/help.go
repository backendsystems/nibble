package common

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// HelpConfig defines the content and appearance of a help overlay
type HelpConfig struct {
	Title        string
	Content      []string
	Width        int // Width of help box (default 56)
	ViewWidth    int // Width of the view/terminal for centering
	ViewHeight   int // Height of the view/terminal for placement
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

	// Calculate help box width: use 80% of window width with min 56, max 80
	width := config.Width
	if width == 0 {
		width = int(float64(viewWidth) * 0.8)
		if width < 56 {
			width = 56
		}
		if width > 80 {
			width = 80
		}
	}

	titleRow := renderHelpTitle(config.Title, width-2) // -2 for padding
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
	spacer := strings.Repeat(" ", width-lipgloss.Width(styledTitle)-lipgloss.Width(icon))
	return styledTitle + spacer + icon
}
