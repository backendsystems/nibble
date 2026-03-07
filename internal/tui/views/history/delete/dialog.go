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
	// Calculate box width early so help text can wrap to the dialog content width.
	width := int(float64(viewWidth) * 0.6)
	width = max(min(width, 60), 46)

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
	help := common.HelpTextStyle.Render(common.WrapWords("←/→: navigate • Enter: select • q: back", width-4))

	// Combine content
	content := strings.Join([]string{
		warning,
		note,
		"",
		buttons,
		"",
		help,
	}, "\n")

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

// HistoryDeleteDialog is a delete confirmation dialog specific to the history view
type HistoryDeleteDialog struct {
	Target      any   // *historyview.TreeNode
	CursorOnYes bool  // true = Delete selected, false = Cancel selected
	ItemType    string // "scan", "network", "interface"
	ItemName    string // Display name of the item
}

// Render displays the delete confirmation dialog for history view
func (d HistoryDeleteDialog) Render(view string, viewWidth, viewHeight int) string {
	// Use Dialog to render
	dialog := Dialog{
		Target:      d.Target,
		ItemType:    d.ItemType,
		ItemName:    d.ItemName,
		CursorOnYes: d.CursorOnYes,
	}
	return dialog.Render(view, viewWidth, viewHeight)
}

// Toggle toggles the cursor between Delete and Cancel
func (d *HistoryDeleteDialog) Toggle() {
	if d != nil {
		d.CursorOnYes = !d.CursorOnYes
	}
}

// IsDeleteSelected returns true if Delete is currently selected
func (d *HistoryDeleteDialog) IsDeleteSelected() bool {
	return d != nil && d.CursorOnYes
}
