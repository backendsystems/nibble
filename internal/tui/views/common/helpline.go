package common

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// HelpItem represents a clickable help item with its rendered position.
// Action is an opaque int — each view casts it to its own action type.
type HelpItem struct {
	Text   string
	Action int // view-specific action value
	StartX int // starting X position in the rendered line
	EndX   int // ending X position in the rendered line
	Line   int // line index in wrapped helpline (0-based)
}

// HelpLineLayout stores the computed layout for a helpline.
type HelpLineLayout struct {
	Items     []HelpItem
	LineCount int
}

const HelpSeparator = " • "

// BuildHelpLineLayout computes item positions for the given items and navPrefix,
// wrapping at maxWidth. navPrefix is the non-clickable leading text (e.g. "←/→/↑/↓").
func BuildHelpLineLayout(items []HelpItem, navPrefix string, maxWidth int) *HelpLineLayout {
	prefixLines := strings.Split(WrapWords(navPrefix, maxWidth), "\n")
	if len(prefixLines) == 0 {
		prefixLines = []string{navPrefix}
	}

	line := len(prefixLines) - 1
	currentX := len([]rune(prefixLines[line]))
	sepWidth := len([]rune(HelpSeparator))

	for i := range items {
		itemWidth := len([]rune(items[i].Text))
		candidateWidth := currentX + sepWidth + itemWidth

		if currentX == 0 {
			items[i].StartX = 0
			items[i].EndX = itemWidth
			items[i].Line = line
			currentX = itemWidth
			continue
		}

		if candidateWidth <= maxWidth {
			items[i].StartX = currentX + sepWidth
			items[i].EndX = items[i].StartX + itemWidth
			items[i].Line = line
			currentX = candidateWidth
			continue
		}

		line++
		currentX = 0
		items[i].StartX = 0
		items[i].EndX = itemWidth
		items[i].Line = line
		currentX = itemWidth
	}

	return &HelpLineLayout{
		Items:     items,
		LineCount: line + 1,
	}
}

// RenderHelpLine renders the helpline with hover highlighting on hoveredIndex (-1 = none).
func RenderHelpLine(layout *HelpLineLayout, navPrefix string, maxWidth int, hoveredIndex int) string {
	prefixLines := strings.Split(WrapWords(navPrefix, maxWidth), "\n")
	if len(prefixLines) == 0 {
		prefixLines = []string{navPrefix}
	}

	lines := make([][]string, layout.LineCount)
	for i, line := range prefixLines {
		if i >= len(lines) {
			break
		}
		lines[i] = append(lines[i], HelpTextStyle.Render(line))
	}

	for i, item := range layout.Items {
		style := HelpTextStyle
		if i == hoveredIndex {
			style = lipgloss.NewStyle().
				Foreground(Color.Selection).
				Bold(true)
		}

		if len(lines[item.Line]) > 0 {
			lines[item.Line] = append(lines[item.Line], HelpTextStyle.Render(HelpSeparator))
		}
		lines[item.Line] = append(lines[item.Line], style.Render(item.Text))
	}

	renderedLines := make([]string, 0, len(lines))
	for _, line := range lines {
		renderedLines = append(renderedLines, strings.Join(line, ""))
	}

	return strings.Join(renderedLines, "\n")
}

// GetHelpItemAt returns the item index at (x, relY) within the helpline,
// where relY=0 is the first helpline row. Returns -1 on miss.
func GetHelpItemAt(layout *HelpLineLayout, x, relY int) int {
	for i, item := range layout.Items {
		if item.Line == relY && x >= item.StartX && x < item.EndX {
			return i
		}
	}
	return -1
}
