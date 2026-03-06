package common

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// Color holds the global color palette used throughout the application
var Color = struct {
	Selection   lipgloss.Color // Primary highlight/selection color (yellow)
	Help        lipgloss.Color // Help text and secondary info (gray)
	Info        lipgloss.Color // Primary text and information (white)
	Error       lipgloss.Color // Error and danger messages (red)
	ProgressGreen lipgloss.Color // Progress bar and found ports color (green tint)
}{
	Selection:   lipgloss.Color("226"),
	Help:        lipgloss.Color("240"),
	Info:        lipgloss.Color("15"),
	Error:       lipgloss.Color("196"),
	ProgressGreen: lipgloss.Color("150"),
}

var (
	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(Color.Info)

	// Help overlay styles
	HelpBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Color.Selection).
			Padding(0, 1).
			Foreground(Color.Info)

	HelpTitleStyle = lipgloss.NewStyle().
			Foreground(Color.Selection).
			Bold(true)

	HelpIconStyle = lipgloss.NewStyle().
			Foreground(Color.Selection).
			Bold(true)

	// Common text styles
	HelpTextStyle = lipgloss.NewStyle().
			Foreground(Color.Help)

	InfoTextStyle = lipgloss.NewStyle().
			Foreground(Color.Help)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Color.Error).
			Bold(true)

	HighlightStyle = lipgloss.NewStyle().
			Foreground(Color.Selection).
			Bold(true)

	MutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("239")).
			Italic(true)

	ItalicHelpStyle = lipgloss.NewStyle().
			Foreground(Color.Help).
			Italic(true)

	ProgressGreenStyle = lipgloss.NewStyle().
			Foreground(Color.ProgressGreen)
)

// FormTheme returns a custom huh theme with yellow selection colors
// matching the rest of the application
func FormTheme() *huh.Theme {
	theme := huh.ThemeCharm()

	// Update focused field styles to use yellow selection color
	theme.Focused.SelectSelector = theme.Focused.SelectSelector.Foreground(Color.Selection)
	theme.Focused.Option = theme.Focused.Option.Foreground(Color.Selection)
	theme.Focused.MultiSelectSelector = theme.Focused.MultiSelectSelector.Foreground(Color.Selection)
	theme.Focused.SelectedOption = theme.Focused.SelectedOption.Foreground(Color.Selection)
	theme.Focused.SelectedPrefix = theme.Focused.SelectedPrefix.Foreground(Color.Selection)
	theme.Focused.FocusedButton = theme.Focused.FocusedButton.Foreground(Color.Selection)
	theme.Focused.Title = theme.Focused.Title.Foreground(Color.Selection)
	theme.Focused.TextInput.Prompt = theme.Focused.TextInput.Prompt.Foreground(Color.Selection)

	return theme
}
