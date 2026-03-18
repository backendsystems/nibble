package portsview

import (
	"github.com/backendsystems/nibble/internal/scanner/shared"
	"github.com/backendsystems/nibble/internal/tui/views/common"
)

type Model struct {
	ShowHelp     bool
	PortPack     string
	CustomPorts  string
	CustomCursor int
	PortInput    common.CustomPortInput
	ErrorMsg        string
	NetworkScan     shared.Scanner
	HoveredHelpItem int // -1 means no hover, otherwise index of helpline item
	HelpLineY       int // Y row where the helpline starts, set during render
}
