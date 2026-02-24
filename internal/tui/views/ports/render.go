package portsview

import (
	"strconv"
	"strings"

	"github.com/backendsystems/nibble/internal/ports"
	"github.com/backendsystems/nibble/internal/tui/views/common"
)

func Render(m Model, maxWidth int) string {
	var b strings.Builder

	b.WriteString(common.TitleStyle.Render("Configure Scan Ports") + "\n")

	defaultStyle := common.InfoTextStyle
	customStyle := common.InfoTextStyle
	if m.PortPack == "default" {
		defaultStyle = common.HighlightStyle
	} else {
		customStyle = common.HighlightStyle
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
		b.WriteString(common.ItalicHelpStyle.Render(guide) + "\n")
	} else {
		b.WriteString("\n")
	}

	if m.ErrorMsg != "" {
		b.WriteString("\n" + common.ErrorStyle.Render("Error: "+m.ErrorMsg) + "\n")
	}

	b.WriteString("\n" + common.HelpTextStyle.Render(common.WrapWords(portsHelpText, maxWidth)))

	view := b.String()
	if m.ShowHelp {
		return renderHelpOverlay(view, maxWidth)
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
