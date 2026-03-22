package scanview

import "github.com/backendsystems/nibble/internal/tui/views/common"

const scanHelpPrefix = "↑/↓: scroll"

var scanHelpItems = []common.HelpItem{
	{Text: "q: quit", Action: int(ActionQuit)},
}
