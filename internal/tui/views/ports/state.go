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
	ErrorMsg     string
	NetworkScan  shared.Scanner
}
