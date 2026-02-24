package portsview

import (
	"strconv"
	"strings"

	"github.com/backendsystems/nibble/internal/ports"
	"github.com/backendsystems/nibble/internal/tui/views/common"
	"github.com/charmbracelet/lipgloss"
)

func Render(m Model, maxWidth int) string {
	var b strings.Builder

	b.WriteString(common.TitleStyle.Render("Configure Scan Ports") + "\n")

	defaultStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	customStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	if m.PortPack == "default" {
		defaultStyle = defaultStyle.Foreground(lipgloss.Color("226")).Bold(true)
	} else {
		customStyle = customStyle.Foreground(lipgloss.Color("226")).Bold(true)
	}

	defaultLine := wrapPortList("default: ", formatPortList(ports.DefaultPorts()), maxWidth)
	b.WriteString(defaultStyle.Render(defaultLine) + "\n")

	customLine := ""
	if m.PortPack == "custom" && m.InputReady {
		input := m.CustomInput
		available := maxWidth - len("custom:  ")
		if available > 0 {
			input.Width = available
		}
		customLine = "custom:  " + input.View()
	} else {
		customLine = wrapPortList("custom:  ", m.CustomPorts, maxWidth)
	}
	b.WriteString(customStyle.Render(customLine) + "\n")

	if m.PortPack == "custom" {
		guide := "  • " + common.CustomPortsDescription
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true).Render(guide) + "\n")
	} else {
		b.WriteString("\n")
	}

	if m.ErrorMsg != "" {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
		b.WriteString("\n" + errorStyle.Render("Error: "+m.ErrorMsg) + "\n")
	}

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	b.WriteString("\n" + helpStyle.Render(common.WrapWords(portsHelpText, maxWidth)))

	view := b.String()
	if m.ShowHelp {
		return renderHelpOverlay(view)
	}
	return view
}

func formatPortList(portList []int) string {
	if len(portList) == 0 {
		return ""
	}
	parts := make([]string, 0, len(portList))
	for _, p := range portList {
		parts = append(parts, strconv.Itoa(p))
	}
	return strings.Join(parts, ",")
}
