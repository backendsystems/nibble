package targetview

import "github.com/backendsystems/nibble/internal/tui/views/common"

const targetHelpPrefix = ""

const (
	targetActionSubmit = iota
	targetActionHelp
	targetActionQuit
)

var targetHelpItems = []common.HelpItem{
	{Text: "enter: submit", Action: targetActionSubmit},
	{Text: "?: help", Action: targetActionHelp},
	{Text: "q: back", Action: targetActionQuit},
}
