package historyview

import "github.com/backendsystems/nibble/internal/tui/views/common"

func renderListHelpOverlay(view string, viewWidth, viewHeight int) string {
	return common.RenderHelpOverlay(view, common.HelpConfig{
		Title:      "Scan History",
		ViewWidth:  viewWidth,
		ViewHeight: viewHeight,
		Content: []string{
			"Browse and manage your network scan history.",
			"• ↑/↓: navigate",
			"• →: expand folder or view scan",
			"• ←: collapse folder",
			"• Del: delete",
			"",
			"any key: close",
		},
	})
}

func renderDetailHelpOverlay(view string, viewWidth, viewHeight int) string {
	return common.RenderHelpOverlay(view, common.HelpConfig{
		Title:      "Scan Details",
		ViewWidth:  viewWidth,
		ViewHeight: viewHeight,
		Content: []string{
			"View scan results and rescan hosts with all ports.",
			"• ↑/↓: select host",
			"• enter: rescan all 65535 ports",
			"• q: back",
		},
	})
}
