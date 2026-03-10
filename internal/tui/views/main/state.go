package mainview

import (
	"net"

	"github.com/charmbracelet/bubbles/viewport"
)

type Model struct {
	Interfaces   []net.Interface
	InterfaceMap map[string][]net.Addr
	Cursor       int
	CardsPerRow  int
	ShowHelp     bool
	ErrorMsg     string
	Viewport     viewport.Model
	WindowH      int
}
