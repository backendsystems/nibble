package common

import (
	"github.com/backendsystems/nibble/internal/tui/views/common/colors"
	"github.com/charmbracelet/lipgloss"
)

// Color holds the global color palette used throughout the application
var Color = struct {
	Black     lipgloss.Color
	Selection lipgloss.Color
	Help      lipgloss.Color
	Info      lipgloss.Color
	Error     lipgloss.Color
	Scanned   lipgloss.Color
	Warning   lipgloss.Color
	Folder    lipgloss.Color
}{
	Black:     colors.Black,
	Selection: colors.Yellow,
	Help:      colors.BrightBlack,
	Info:      colors.White,
	Error:     colors.Red,
	Scanned:   colors.Green,
	Warning:   colors.BrightYellow,
	Folder:    colors.Blue,
}
