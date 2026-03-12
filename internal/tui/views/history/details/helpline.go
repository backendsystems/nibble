package historydetailview

import "github.com/backendsystems/nibble/internal/tui/views/common"

const detailHelpPrefix = "↑/↓: select host"

var detailHelpItems = []common.HelpItem{
	{Text: "Enter/→: scan all ports", Action: int(ActionScanAllPorts)},
	{Text: "←/q: back", Action: int(ActionQuit)},
	{Text: "?: help", Action: int(ActionHelp)},
}
