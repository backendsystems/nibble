package portsview

import "github.com/backendsystems/nibble/internal/scanner/shared"

type Model struct {
	ShowHelp     bool
	PortPack     string
	CustomPorts  string
	CustomCursor int
	ErrorMsg     string
	NetworkScan  shared.Scanner
}
