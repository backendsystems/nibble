package mainview

import "github.com/backendsystems/nibble/internal/tui/views/common"

func renderHelpOverlay(view string, maxWidth int) string {
	return common.RenderHelpOverlay(view, common.HelpConfig{
		Title:     "Nibble Network Scanner",
		ViewWidth: maxWidth,
		Content: []string{
			"Scans local networks for active hosts.",
			"• Press p to configure ports",
			"• Press t for target mode scan",
			"• Press r to view scan history",
			"• Click a card to select, click again to confirm",
			"• Shift+drag to select text",
			"",
			"any key: close",
		},
	})
}
