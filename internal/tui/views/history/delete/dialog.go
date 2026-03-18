package delete

import (
	"fmt"
	"strings"

	"github.com/backendsystems/nibble/internal/tui/views/common"
	"github.com/charmbracelet/lipgloss"
)

const (
	dialogMinWidth = 46
	dialogMaxWidth = 60
)

// Dialog is a reusable delete confirmation dialog
type Dialog struct {
	Target      any    // Generic target being deleted
	ItemType    string // "scan", "network", "interface", etc
	ItemName    string // Display name of the item
	CursorOnYes bool   // true = Delete selected, false = Cancel selected
}

func dialogWidth(viewWidth int) int {
	width := int(float64(viewWidth) * 0.6)
	return max(min(width, dialogMaxWidth), dialogMinWidth)
}

func dialogButtonStyles() (lipgloss.Style, lipgloss.Style) {
	// Button styles - same card style used across views
	selectedButtonStyle := common.SelectedCardStyle.
		Padding(0, 1).
		MarginRight(1).
		Foreground(common.Color.Selection).
		Bold(true)

	unselectedButtonStyle := common.CardStyle.
		Padding(0, 1).
		MarginRight(1).
		Foreground(common.Color.Info)

	return selectedButtonStyle, unselectedButtonStyle
}

func dialogButtons(cursorOnYes bool) (string, string, string) {
	selectedButtonStyle, unselectedButtonStyle := dialogButtonStyles()

	// Buttons (Cancel on left, Delete on right)
	var cancelBtn, deleteBtn string
	if cursorOnYes {
		cancelBtn = unselectedButtonStyle.Render("Cancel")
		deleteBtn = selectedButtonStyle.Render("🔥 Delete")
	} else {
		cancelBtn = selectedButtonStyle.Render("Cancel")
		deleteBtn = unselectedButtonStyle.Render("🔥 Delete")
	}

	return cancelBtn, deleteBtn, lipgloss.JoinHorizontal(lipgloss.Center, cancelBtn, deleteBtn)
}

// Render displays the delete confirmation dialog
func (d Dialog) Render(view string, viewWidth, viewHeight int) string {
	if d.Target == nil {
		return view
	}

	// Calculate box width early so help text can wrap to the dialog content width.
	width := dialogWidth(viewWidth)

	// Build content
	warning := lipgloss.NewStyle().Bold(true).Foreground(common.Color.Info).Render(fmt.Sprintf("Delete %s: %s?", d.ItemType, d.ItemName))
	note := common.HelpTextStyle.Render("This action cannot be undone.")
	_, _, buttons := dialogButtons(d.CursorOnYes)
	help := common.HelpTextStyle.Render(common.WrapWords("←/→: navigate • Enter • Del: delete • q: back", width-4))

	// Combine content
	content := strings.Join([]string{
		warning,
		note,
		buttons,
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
	Target      any    // *historyview.TreeNode
	CursorOnYes bool   // true = Delete selected, false = Cancel selected
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
