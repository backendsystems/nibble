package delete

import (
	"fmt"
	"strings"

	"github.com/backendsystems/nibble/internal/tui/views/common"
	"github.com/charmbracelet/lipgloss"
)

// Dialog is a reusable delete confirmation dialog
type Dialog struct {
	Target      any    // Generic target being deleted
	ItemType    string // "scan", "network", "interface", etc
	ItemName    string // Display name of the item
	CursorOnYes bool   // true = Delete selected, false = Cancel selected
}

// Render displays the delete confirmation dialog
func (d Dialog) Render(view string, viewWidth, viewHeight int) string {
	if d.Target == nil {
		return view
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
	warning := common.ErrorStyle.Render(fmt.Sprintf("Delete %s: %s?", d.ItemType, d.ItemName))
	note := common.HelpTextStyle.Render("This action cannot be undone.")

	// Buttons (Cancel on left, Delete on right)
	var cancelBtn, deleteBtn string
	if d.CursorOnYes {
		cancelBtn = unselectedButtonStyle.Render("Cancel")
		deleteBtn = selectedButtonStyle.Render("Delete")
	} else {
		cancelBtn = selectedButtonStyle.Render("Cancel")
		deleteBtn = unselectedButtonStyle.Render("Delete")
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, cancelBtn, deleteBtn)
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
	width = max(min(width, 60), 46)

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
