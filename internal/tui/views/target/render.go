package targetview

import (
	"strings"

	"github.com/backendsystems/nibble/internal/tui/views/common"
)

const (
	guideText = "?: help • q: cancel"
)

func Render(m Model, maxWidth int) string {
	var b strings.Builder

	b.WriteString(common.TitleStyle.Render("Custom Target") + "\n\n")

	// Error message (if any)
	if m.ErrorMsg != "" {
		b.WriteString(common.ErrorStyle.Render("Error: "+m.ErrorMsg) + "\n\n")
	}

	// Form view
	if m.Form != nil {
		b.WriteString(m.Form.View())
	}

	// Guide text
	b.WriteString("\n")
	b.WriteString(common.HelpTextStyle.Render(common.WrapWords(guideText, maxWidth)))

	view := b.String()
	if m.ShowHelp {
		return renderHelpOverlay(view, m, maxWidth)
	}

	return view
}
