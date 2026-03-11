package targetview

import "github.com/backendsystems/nibble/internal/tui/views/common"

const targetHelpPrefix = ""

const (
	targetActionHelp = iota
	targetActionQuit
)

var targetHelpItems = []common.HelpItem{
	{Text: "enter: submit", Action: -1}, // non-clickable info
	{Text: "?: help", Action: targetActionHelp},
	{Text: "q: back", Action: targetActionQuit},
}
