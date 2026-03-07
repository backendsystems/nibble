package historydetailview

import "github.com/backendsystems/nibble/internal/tui/views/common"

func renderHelpOverlay(view string, maxWidth int) string {
	return common.RenderHelpOverlay(view, common.HelpConfig{
		Title:     "Scan Details",
		ViewWidth: maxWidth,
		Content: []string{
			"View scan results and rescan hosts with all ports.",
			"• ↑/↓: select host",
			"• enter: rescan all 65535 ports",
			"• del: delete scan",
			"• ←/q: back",
			"",
			common.ProgressGreenStyle.Render("✓") + " means all 65535 ports scanned on host",
			"any key: close",
		},
	})
}
