package mainview

import "github.com/backendsystems/nibble/internal/tui/views/common"

const helpPrefixText = "←/→/↑/↓ a/d/w/s h/j/k/l"

var mainHelpItems = []common.HelpItem{
	{Text: "p: ports", Action: int(ActionOpenPorts)},
	{Text: "r: history", Action: int(ActionOpenHistory)},
	{Text: "t: target", Action: int(ActionOpenTarget)},
	{Text: "?: help", Action: int(ActionOpenHelp)},
	{Text: "q: quit", Action: int(ActionQuit)},
}
