package targetview

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func renderHelpOverlay(view string, m Model) string {
	if m.Form != nil {
		if focused := m.Form.GetFocusedField(); focused != nil && focused.GetKey() == "custom_ports" {
			return renderPortsHelpOverlay(view)
		}
	}
	return renderMainHelpOverlay(view)
}

func renderMainHelpOverlay(view string) string {
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
	icon := iconStyle.Render("❓")
	spacer := strings.Repeat(" ", titleWidth-lipgloss.Width(helpTitle)-lipgloss.Width(icon))
	titleRow := helpTitle + spacer + icon

	helpContent := strings.Join([]string{
		titleRow,
		"This mode is usefull to scan all ports on a single host, or to scan a custom list of ports on a subnet",
		"",
		"• tab/↑↓: move through options",
		"• ←/→: select interface ip",
		"• delete: clear current field",
		"• enter: submit",
		"• q: cancel",
		"• enter: submit",
	}, "\n")

	return placeHelpOverlay(view, helpBox.Render(helpContent))
}

func renderPortsHelpOverlay(view string) string {
	helpBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("226")).
		Padding(0, 1).
		Width(56).
		Foreground(lipgloss.Color("15"))

	helpTitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")).
		Bold(true).
		Render("Port Configuration")

	iconStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")).
		Bold(true)

	titleWidth := 54
	icon := iconStyle.Render("❓")
	spacer := strings.Repeat(" ", titleWidth-lipgloss.Width(helpTitle)-lipgloss.Width(icon))
	titleRow := helpTitle + spacer + icon

	helpContent := strings.Join([]string{
		titleRow,
		"Configure which ports get scanned.",
		"• ←/→ or a/d or h/l: move cursor",
		"• type digits, commas, and ranges (e.g. 8000-9000)",
		"• backspace: remove",
		"• delete: clear all",
		"• q: cancel",
		"• enter: save and return",
	}, "\n")

	return placeHelpOverlay(view, helpBox.Render(helpContent))
}

func placeHelpOverlay(view, helpOverlay string) string {
	return lipgloss.Place(
		lipgloss.Width(view),
		lipgloss.Height(view),
		lipgloss.Center,
		lipgloss.Top,
		helpOverlay,
		lipgloss.WithWhitespaceChars(" "),
	)
}
