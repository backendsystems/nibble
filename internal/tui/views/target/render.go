package targetview

import (
	"strings"

	"github.com/backendsystems/nibble/internal/tui/views/common"
)

func Render(m *Model, maxWidth int) string {
	var b strings.Builder

	if m.InCustomPortInput {
		// Stage 2: Custom port textinput
		b.WriteString(common.TitleStyle.Render("Custom Target - Custom ports") + "\n")
	} else {
		// Stage 1: Form
		b.WriteString(common.TitleStyle.Render("Custom Target") + "\n")
	}

	if m.InCustomPortInput {
		// Stage 2: Render custom port textinput
		b.WriteString("\n")
		input := m.PortInput.Input
		available := maxWidth - len("custom:  ")
		if available > 0 {
			input.Width = available
		}
		b.WriteString(common.HighlightStyle.Render("custom:  ") + input.View() + "\n")

		guide := "  • " + common.CustomPortsDescription
		b.WriteString(common.ItalicHelpStyle.Render(guide) + "\n")
	} else {
		// Stage 1: Form view
		b.WriteString("\n")
		if m.Form != nil {
			b.WriteString(m.Form.View())
		}
	}

	// Error message (if any)
	if m.ErrorMsg != "" {
		b.WriteString("\n" + common.ErrorStyle.Render("Error: "+m.ErrorMsg) + "\n")
	}

	m.HelpLineY = strings.Count(b.String(), "\n") + 1
	layout := common.BuildHelpLineLayout(targetHelpItems, targetHelpPrefix, maxWidth)
	b.WriteString("\n" + common.RenderHelpLine(layout, targetHelpPrefix, maxWidth, m.HoveredHelpItem))

	view := b.String()
	if m.ShowHelp {
		return renderHelpOverlay(view, *m, maxWidth)
	}

	return view
}
