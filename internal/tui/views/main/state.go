package mainview

import (
	"net"

	"charm.land/bubbles/v2/viewport"
)

type Model struct {
	Interfaces      []net.Interface
	InterfaceMap    map[string][]net.Addr
	Cursor          int
	CardsPerRow     int
	CardWidth       int // computed from content; includes border+padding
	ShowHelp        bool
	ErrorMsg        string
	Viewport        viewport.Model
	WindowH         int
	HoveredHelpItem int // -1 means no hover, otherwise index of helpline item
	HelpLineY       int // Y row where the helpline starts, set during render
}
