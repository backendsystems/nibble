package common

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

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
			Foreground(Color.Help).
			Italic(true)

	ItalicHelpStyle = lipgloss.NewStyle().
			Foreground(Color.Help).
			Italic(true)

	ProgressGreenStyle = lipgloss.NewStyle().
				Foreground(Color.Scanned)

	WarningStyle = lipgloss.NewStyle().
			Foreground(Color.Scanning).
			Bold(true)

	// Card styles - shared rounded-border card look used across views
	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Color.Help)

	SelectedCardStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Color.Selection)

	FolderStyle = lipgloss.NewStyle().
			Foreground(Color.Folder)
)

// FormTheme returns a custom huh theme using only ANSI 16 colors so it
// follows the user's terminal theme.
func FormTheme() *huh.Theme {
	theme := huh.ThemeBase()

	// Focused state
	theme.Focused.Base = theme.Focused.Base.BorderForeground(Color.Help)
	theme.Focused.Title = theme.Focused.Title.Foreground(Color.Selection).Bold(true)
	theme.Focused.NoteTitle = theme.Focused.NoteTitle.Foreground(Color.Selection).Bold(true)
	theme.Focused.Description = theme.Focused.Description.Foreground(Color.Help)
	theme.Focused.ErrorIndicator = theme.Focused.ErrorIndicator.Foreground(Color.Error)
	theme.Focused.ErrorMessage = theme.Focused.ErrorMessage.Foreground(Color.Error)
	theme.Focused.SelectSelector = theme.Focused.SelectSelector.Foreground(Color.Selection)
	theme.Focused.NextIndicator = theme.Focused.NextIndicator.Foreground(Color.Selection)
	theme.Focused.PrevIndicator = theme.Focused.PrevIndicator.Foreground(Color.Selection)
	theme.Focused.Option = theme.Focused.Option.Foreground(Color.Selection)
	theme.Focused.MultiSelectSelector = theme.Focused.MultiSelectSelector.Foreground(Color.Selection)
	theme.Focused.SelectedOption = theme.Focused.SelectedOption.Foreground(Color.Selection)
	theme.Focused.SelectedPrefix = theme.Focused.SelectedPrefix.Foreground(Color.Selection)
	theme.Focused.UnselectedPrefix = theme.Focused.UnselectedPrefix.Foreground(Color.Help)
	theme.Focused.UnselectedOption = theme.Focused.UnselectedOption.Foreground(Color.Help)
	theme.Focused.FocusedButton = theme.Focused.FocusedButton.Foreground(Color.Black).Background(Color.Selection)
	theme.Focused.BlurredButton = theme.Focused.BlurredButton.Foreground(Color.Info).Background(Color.Black)
	theme.Focused.TextInput.Cursor = theme.Focused.TextInput.Cursor.Foreground(Color.Selection)
	theme.Focused.TextInput.Placeholder = theme.Focused.TextInput.Placeholder.Foreground(Color.Help)
	theme.Focused.TextInput.Prompt = theme.Focused.TextInput.Prompt.Foreground(Color.Selection)

	// Blurred state inherits focused, just hide the border and dim the title
	theme.Blurred = theme.Focused
	theme.Blurred.Base = theme.Focused.Base.BorderStyle(lipgloss.HiddenBorder())
	theme.Blurred.Card = theme.Blurred.Base
	theme.Blurred.Title = theme.Focused.Title.Foreground(Color.Folder)
	theme.Blurred.NoteTitle = theme.Focused.NoteTitle.Foreground(Color.Folder)
	theme.Blurred.NextIndicator = lipgloss.NewStyle()
	theme.Blurred.PrevIndicator = lipgloss.NewStyle()

	// Group title/description
	theme.Group.Title = theme.Focused.Title
	theme.Group.Description = theme.Focused.Description

	return theme
}
