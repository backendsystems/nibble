package targetview

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func renderHelpOverlay(view string) string {
	helpBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("226")).
		Padding(0, 1).
		Width(56).
		Foreground(lipgloss.Color("15"))

	helpTitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")).
		Bold(true).
		Render("Custom Target")

	iconStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")).
		Bold(true)

	titleWidth := 54
	icon := iconStyle.Render("🎯")
	spacer := strings.Repeat(" ", titleWidth-lipgloss.Width(helpTitle)-lipgloss.Width(icon))
	titleRow := helpTitle + spacer + icon

	helpContent := strings.Join([]string{
		titleRow,
		"Enter a custom target IP and CIDR to scan.",
		"CIDR range: 16-32 (default: 32 for single host)",
		"",
		"• tab/↑↓ or w/s k/j: switch between fields",
		"• ←/→ or a/d: move cursor, or toggle ports mode",
		"• type: enter IP, CIDR, or ports",
		"• backspace: remove character",
		"• delete: clear ports field",
		"• enter: start scan",
		"• q: cancel",
	}, "\n")

	helpOverlay := helpBox.Render(helpContent)
	return lipgloss.Place(
		lipgloss.Width(view),
		lipgloss.Height(view),
		lipgloss.Center,
		lipgloss.Top,
		helpOverlay,
		lipgloss.WithWhitespaceChars(" "),
	)
}
