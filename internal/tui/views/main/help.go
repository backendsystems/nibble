package mainview

import "github.com/backendsystems/nibble/internal/tui/views/common"

func renderHelpOverlay(view string, maxWidth int) string {
	return common.RenderHelpOverlay(view, common.HelpConfig{
		Title:     "Nibble Network Scanner",
		ViewWidth: maxWidth,
		Content: []string{
			"Scans local networks for active hosts.",
			"• Scans TCP ports",
			"  • Press p to configure ports",
			"• Grabs service banners (SSH, HTTP Server)",
			"• Identifies hardware via MAC OUI (IEEE)",
			"",
			"any key: close",
		},
	})
}
