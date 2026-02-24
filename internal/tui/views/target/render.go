package targetview

import (
	"strings"

	"github.com/backendsystems/nibble/internal/tui/views/common"
	"github.com/charmbracelet/lipgloss"
)

const (
	guideText = "?: help • q: cancel"
)

func Render(m Model, maxWidth int) string {
	var b strings.Builder

	b.WriteString(common.TitleStyle.Render("Custom Target") + "\n\n")

	// Error message (if any)
	if m.ErrorMsg != "" {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
		b.WriteString(errorStyle.Render("Error: "+m.ErrorMsg) + "\n\n")
	}

	// Form view
	if m.Form != nil {
		b.WriteString(m.Form.View())
	}

	// Guide text
	b.WriteString("\n")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	b.WriteString(helpStyle.Render(common.WrapWords(guideText, maxWidth)))

	view := b.String()
	if m.ShowHelp {
		return renderHelpOverlay(view)
	}

	return view
}
