package delete

import (
	"fmt"

	"github.com/backendsystems/nibble/internal/tui/views/common"
	"github.com/charmbracelet/lipgloss"
)

type dialogButton int

const (
	dialogButtonNone dialogButton = iota
	dialogButtonCancel
	dialogButtonDelete
)

type MouseAction int

const (
	MouseActionNone MouseAction = iota
	MouseActionConfirmNo
	MouseActionConfirmYes
)

func clickedButton(d Dialog, x, y, viewWidth int) dialogButton {
	width := dialogWidth(viewWidth)
	overlayLeft := max(0, (viewWidth-width)/2)

	warning := lipgloss.NewStyle().Bold(true).Foreground(common.Color.Info).Render(fmt.Sprintf("Delete %s: %s?", d.ItemType, d.ItemName))
	note := common.HelpTextStyle.Render("This action cannot be undone.")
	cancelBtn, deleteBtn, buttons := dialogButtons(d.CursorOnYes)

	// HelpBoxStyle has top border (1 line), then content rows.
	buttonStartY := 1 + lipgloss.Height(warning) + lipgloss.Height(note)
	buttonEndY := buttonStartY + lipgloss.Height(buttons) - 1
	if y < buttonStartY || y > buttonEndY {
		return dialogButtonNone
	}

	// HelpBoxStyle has left border + left padding = 2 columns before content.
	buttonStartX := overlayLeft + 2
	cancelEndX := buttonStartX + lipgloss.Width(cancelBtn) - 1
	deleteStartX := cancelEndX + 1
	deleteEndX := deleteStartX + lipgloss.Width(deleteBtn) - 1

	if x >= buttonStartX && x <= cancelEndX {
		return dialogButtonCancel
	}
	if x >= deleteStartX && x <= deleteEndX {
		return dialogButtonDelete
	}
	return dialogButtonNone
}

// HandleMouseClick applies main-card click semantics to delete dialog buttons:
// first click selects a button, second click on the selected button confirms it.
func (d *HistoryDeleteDialog) HandleMouseClick(x, y, viewWidth, viewHeight int) (handled bool, action MouseAction) {
	if d == nil {
		return false, MouseActionNone
	}
	_ = viewHeight

	dialog := Dialog{
		Target:      d.Target,
		ItemType:    d.ItemType,
		ItemName:    d.ItemName,
		CursorOnYes: d.CursorOnYes,
	}
	switch clickedButton(dialog, x, y, viewWidth) {
	case dialogButtonCancel:
		if d.CursorOnYes {
			d.CursorOnYes = false
			return true, MouseActionNone
		}
		return true, MouseActionConfirmNo
	case dialogButtonDelete:
		if !d.CursorOnYes {
			d.CursorOnYes = true
			return true, MouseActionNone
		}
		return true, MouseActionConfirmYes
	default:
		return false, MouseActionNone
	}
}
