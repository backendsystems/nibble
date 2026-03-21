package common

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Color holds the global color palette used throughout the application.
// Uses lipgloss v2 ANSI 4-bit constants so colors adapt to any terminal theme.
var Color = struct {
	Black     color.Color
	Selection color.Color
	Help      color.Color
	Info      color.Color
	Error     color.Color
	Scanned   color.Color
	Scanning  color.Color
	Folder    color.Color
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
