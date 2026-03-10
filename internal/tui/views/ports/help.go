package portsview

import "github.com/backendsystems/nibble/internal/tui/views/common"

func renderHelpOverlay(view string, maxWidth int) string {
	return common.RenderHelpOverlay(view, common.HelpConfig{
		Title:     "Port Configuration",
		ViewWidth: maxWidth,
		Content: []string{
			"Configure which ports get scanned.",
			"• tab/↑↓: switch default/custom mode",
			"• ←/→: move cursor in custom list",
			"• type digits, commas, and ranges (e.g. 8000-9000)",
			"• backspace: remove",
			"• delete: clear all",
			"• enter: save and return",
			"• Click a mode to select, click again to apply",
			"• Shift+drag to select text",
			"",
			"any key: close",
		},
	})
}
