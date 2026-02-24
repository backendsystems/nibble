package targetview

import "github.com/backendsystems/nibble/internal/tui/views/common"

func renderHelpOverlay(view string, m Model, maxWidth int) string {
	if m.InCustomPortInput {
		return renderPortsHelpOverlay(view, maxWidth)
	}
	return renderMainHelpOverlay(view, maxWidth)
}

func renderMainHelpOverlay(view string, maxWidth int) string {
	return common.RenderHelpOverlay(view, common.HelpConfig{
		Title:      "Custom Target",
		ViewWidth:  maxWidth,
		Content: []string{
			"This mode is usefull to scan all ports on a single host, or to scan a custom list of ports on a subnet",
			"",
			"• tab/↑↓: move through options",
			"• ←/→: select interface ip",
			"• delete: clear current field",
			"• enter: submit",
			"• q: cancel",
			"• enter: submit",
		},
	})
}

func renderPortsHelpOverlay(view string, maxWidth int) string {
	return common.RenderHelpOverlay(view, common.HelpConfig{
		Title:      "Port Configuration",
		ViewWidth:  maxWidth,
		Content: []string{
			"Configure which ports get scanned.",
			"• ←/→ or a/d or h/l: move cursor",
			"• type digits, commas, and ranges (e.g. 8000-9000)",
			"• backspace: remove",
			"• delete: clear all",
			"• q: cancel",
			"• enter: save and return",
		},
	})
}
