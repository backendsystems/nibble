package common

import (
	"charm.land/lipgloss/v2"
)

// Color holds the global color palette used throughout the application.
// Uses lipgloss v2 ANSI 4-bit constants so colors adapt to any terminal theme.
var Color = struct {
	Black     lipgloss.BasicColor
	Selection lipgloss.BasicColor
	Help      lipgloss.BasicColor
	Info      lipgloss.BasicColor
	Error     lipgloss.BasicColor
	Scanned   lipgloss.BasicColor
	Scanning  lipgloss.BasicColor
	Folder    lipgloss.BasicColor
}{
	Black:     lipgloss.Black,
	Selection: lipgloss.Yellow,
	Help:      lipgloss.BrightBlack,
	Info:      lipgloss.White,
	Error:     lipgloss.Red,
	Scanned:   lipgloss.Green,
	Scanning:  lipgloss.Cyan,
	Folder:    lipgloss.Blue,
}
