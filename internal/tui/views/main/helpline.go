package mainview

import (
	"strings"

	"github.com/backendsystems/nibble/internal/tui/views/common"
	"github.com/charmbracelet/lipgloss"
)

// HelpItem represents a clickable help item with its position
type HelpItem struct {
	Text   string
	Key    string
	Action Action
	StartX int // Starting X position in the rendered line
	EndX   int // Ending X position in the rendered line
	Line   int // Line index in wrapped helpline (0-based)
}

// HelpLineLayout stores the complete helpline layout
type HelpLineLayout struct {
	Items     []HelpItem
	LineCount int
}

const (
	helpPrefixText = "←/→/↑/↓ a/d/w/s h/j/k/l"
	helpSeparator  = " • "
)

// BuildHelpLineLayout creates the helpline layout with positions calculated once
func BuildHelpLineLayout(maxWidth int) *HelpLineLayout {
	items := []HelpItem{
		{Text: "p: ports", Key: "p", Action: ActionOpenPorts},
		{Text: "r: history", Key: "r", Action: ActionOpenHistory},
		{Text: "t: target", Key: "t", Action: ActionOpenTarget},
		{Text: "?: help", Key: "?", Action: ActionOpenHelp},
		{Text: "q: quit", Key: "q", Action: ActionQuit},
	}

	prefixLines := strings.Split(common.WrapWords(helpPrefixText, maxWidth), "\n")
	if len(prefixLines) == 0 {
		prefixLines = []string{helpPrefixText}
	}

	line := len(prefixLines) - 1
	currentX := len([]rune(prefixLines[line]))
	sepWidth := len([]rune(helpSeparator))

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

// RenderHelpLine renders the helpline with hover effects
func RenderHelpLine(maxWidth int, hoveredIndex int) string {
	layout := BuildHelpLineLayout(maxWidth)
	prefixLines := strings.Split(common.WrapWords(helpPrefixText, maxWidth), "\n")
	if len(prefixLines) == 0 {
		prefixLines = []string{helpPrefixText}
	}

	lines := make([][]string, layout.LineCount)
	for i, line := range prefixLines {
		if i >= len(lines) {
			break
		}
		lines[i] = append(lines[i], common.HelpTextStyle.Render(line))
	}

	for i, item := range layout.Items {
		style := common.HelpTextStyle
		if i == hoveredIndex {
			style = lipgloss.NewStyle().
				Foreground(common.Color.Selection).
				Bold(true)
		}

		if len(lines[item.Line]) > 0 {
			lines[item.Line] = append(lines[item.Line], common.HelpTextStyle.Render(helpSeparator))
		}
		lines[item.Line] = append(lines[item.Line], style.Render(item.Text))
	}

	renderedLines := make([]string, 0, len(lines))
	for _, line := range lines {
		renderedLines = append(renderedLines, strings.Join(line, ""))
	}

	return strings.Join(renderedLines, "\n")
}

// GetHelpItemAt returns the help item index at the given X position, or -1
func GetHelpItemAt(x, y int, maxWidth int) int {
	layout := BuildHelpLineLayout(maxWidth)

	for i, item := range layout.Items {
		if item.Line == y && x >= item.StartX && x < item.EndX {
			return i
		}
	}

	return -1
}

// GetHelpItemAction returns the action for the help item at the given index
func GetHelpItemAction(index int, maxWidth int) Action {
	layout := BuildHelpLineLayout(maxWidth)

	if index >= 0 && index < len(layout.Items) {
		return layout.Items[index].Action
	}

	return ActionNone
}
