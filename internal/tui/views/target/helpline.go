package targetview

import "github.com/backendsystems/nibble/internal/tui/views/common"

const targetHelpPrefix = "enter"

const (
	targetActionHelp = iota
	targetActionQuit
)

var targetHelpItems = []common.HelpItem{
	{Text: "?: help", Action: targetActionHelp},
	{Text: "q: back", Action: targetActionQuit},
}
