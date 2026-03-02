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
			"• Press y to view scan history",
			"",
			"any key: close",
		},
	})
}
