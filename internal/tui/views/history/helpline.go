package historyview

import "github.com/backendsystems/nibble/internal/tui/views/common"

const historyHelpPrefix = "↑/↓/←/→"

var historyHelpItems = []common.HelpItem{
	{Text: "Del: delete", Action: int(ActionDelete)},
	{Text: "?: help", Action: int(ActionHelp)},
	{Text: "q: back", Action: int(ActionQuit)},
}
