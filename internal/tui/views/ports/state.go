package portsview

import (
	"github.com/backendsystems/nibble/internal/scanner/shared"
	"github.com/charmbracelet/bubbles/textinput"
)

type Model struct {
	ShowHelp     bool
	PortPack     string
	CustomPorts  string
	CustomCursor int
	CustomInput  textinput.Model
	InputReady   bool
	ErrorMsg     string
	NetworkScan  shared.Scanner
}
