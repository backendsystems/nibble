package historyview

import (
	"strings"

	"github.com/backendsystems/nibble/internal/tui/views/common"
	detailsview "github.com/backendsystems/nibble/internal/tui/views/history/details"
	historytree "github.com/backendsystems/nibble/internal/tui/views/history/tree"
)

var (
	titleStyle = common.TitleStyle
	helpStyle  = common.HelpTextStyle
)

func Render(m Model, maxWidth int) string {
	if m.Mode == ViewDetail {
		return detailsview.Render(m.Details, maxWidth, m.WindowH)
	}
	return renderList(m, maxWidth)
}

func renderList(m Model, maxWidth int) string {
	var b strings.Builder

	// Title (outside viewport)
	b.WriteString(titleStyle.Render("Scan History") + "\n\n")

	// Render only the visible rows instead of a fully pre-rendered viewport buffer.
	b.WriteString(historytree.RenderVisibleList(
		m.FlatList,
		m.Tree,
		m.Cursor,
		m.Viewport.YOffset,
		m.Viewport.Height,
	))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(common.WrapWords("↑/↓/←/→ • Del: delete • ?: help • q: back", maxWidth)))

	if m.ErrorMsg != "" {
		b.WriteString("\n\n" + common.ErrorStyle.Render("Error: "+m.ErrorMsg))
	}

	view := b.String()

	// Show overlays (help takes precedence over delete dialog)
	if m.ShowHelp {
		return renderListHelpOverlay(view, m.WindowW, m.WindowH)
	}

	if m.DeleteDialog != nil {
		return m.DeleteDialog.Render(view, m.WindowW, m.WindowH)
	}

	return view
}
