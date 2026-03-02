package historyview

import (
	"fmt"
	"strings"

	"github.com/backendsystems/nibble/internal/tui/views/common"
	"github.com/charmbracelet/lipgloss"
)

type DeleteDialog struct {
	Target      *TreeNode
	CursorOnYes bool // true = Yes, false = Cancel
}

func (d DeleteDialog) Render(view string, viewWidth, viewHeight int) string {
	if d.Target == nil {
		return view
	}

	var itemType string
	var itemName string

	switch d.Target.Type {
	case NodeScan:
		itemType = "scan"
		itemName = d.Target.Name
	case NodeNetwork:
		itemType = "all scans in network"
		itemName = d.Target.Name
	case NodeInterface:
		itemType = "all scans on interface"
		itemName = d.Target.Name
	}

	// Button styles
	buttonStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Margin(0, 1)

	selectedButtonStyle := buttonStyle.
		Foreground(lipgloss.Color("0")).
		Background(common.Color.Selection).
		Bold(true)

	unselectedButtonStyle := buttonStyle.
		Foreground(common.Color.Info).
		Background(lipgloss.Color("236"))

	// Build content
	warning := common.ErrorStyle.Render(fmt.Sprintf("Delete %s: %s?", itemType, itemName))
	note := common.HelpTextStyle.Render("This action cannot be undone.")

	// Buttons
	var deleteBtn, cancelBtn string
	if d.CursorOnYes {
		deleteBtn = selectedButtonStyle.Render("Delete")
		cancelBtn = unselectedButtonStyle.Render("Cancel")
	} else {
		deleteBtn = unselectedButtonStyle.Render("Delete")
		cancelBtn = selectedButtonStyle.Render("Cancel")
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Left, deleteBtn, cancelBtn)
	help := common.HelpTextStyle.Render("←/→: navigate • Enter: select • q: back")

	// Combine content
	content := strings.Join([]string{
		warning,
		note,
		"",
		buttons,
		"",
		help,
	}, "\n")

	// Calculate box width
	width := int(float64(viewWidth) * 0.6)
	if width < 46 {
		width = 46
	}
	if width > 60 {
		width = 60
	}

	// Create overlay with common style
	overlay := common.HelpBoxStyle.Width(width).Render(content)

	// Place overlay over the view (top-centered like help dialog)
	return lipgloss.Place(
		viewWidth,
		viewHeight,
		lipgloss.Center,
		lipgloss.Top,
		overlay,
		lipgloss.WithWhitespaceChars(" "),
	)
}
