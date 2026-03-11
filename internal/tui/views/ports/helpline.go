package portsview

import "github.com/backendsystems/nibble/internal/tui/views/common"

const portsHelpPrefix = "tab"

const (
	portsActionBackspace = iota
	portsActionDeleteAll
	portsActionApply
	portsActionHelp
	portsActionBack
)

var portsHelpItems = []common.HelpItem{
	{Text: "backspace", Action: portsActionBackspace},
	{Text: "delete: clear all", Action: portsActionDeleteAll},
	{Text: "enter", Action: portsActionApply},
	{Text: "?: help", Action: portsActionHelp},
	{Text: "q: back", Action: portsActionBack},
}
