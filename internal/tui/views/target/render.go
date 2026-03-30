package targetview

import (
	"fmt"
	"net"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/backendsystems/nibble/internal/scanner/shared"
	"github.com/backendsystems/nibble/internal/tui/views/common"
)

var (
	fieldTitleFocused = lipgloss.NewStyle().Foreground(common.Color.Selection).Bold(true)
	fieldTitleBlurred = lipgloss.NewStyle().Foreground(common.Color.Folder)
	fieldDescStyle    = lipgloss.NewStyle().Foreground(common.Color.Help)
	optionSelected    = lipgloss.NewStyle().Foreground(common.Color.Selection).Bold(true)
	optionHovered     = lipgloss.NewStyle().Foreground(common.Color.Selection)
	optionNormal      = lipgloss.NewStyle().Foreground(common.Color.Help)
	selectorStyle     = lipgloss.NewStyle().Foreground(common.Color.Selection)
)

func Render(m *Model, maxWidth int) string {
	var b strings.Builder

	if m.InCustomPortInput {
		b.WriteString(common.TitleStyle.Render("Custom Target - Custom ports") + "\n")
	} else {
		b.WriteString(common.TitleStyle.Render("Custom Target") + "\n")
	}

	if m.InCustomPortInput {
		// Stage 2: Render custom port textinput
		b.WriteString("\n")
		input := m.PortInput.Input
		available := maxWidth - len("custom:  ")
		if available > 0 {
			input.SetWidth(available)
		}
		b.WriteString(common.HighlightStyle.Render("custom:  ") + input.View() + "\n")

		guide := "  • " + common.CustomPortsDescription
		b.WriteString(common.ItalicHelpStyle.Render(guide) + "\n")
	} else {
		// Stage 1: render all fields into viewport content.
		content := renderInterfaceField(m) +
			renderField(m, fieldIP, maxWidth) +
			renderField(m, fieldCIDR, maxWidth) +
			renderPortModeField(m)

		absRow := 0
		if m.WindowH > 0 {
			m.UpdateViewport(maxWidth)
			m.Viewport.SetContent(content)
			m.scrollToFocused()
			// FieldY = on-screen row: title(1) + fieldAbsRow - YOffset
			for f := range fieldCount {
				m.FieldY[f] = 1 + absRow - m.Viewport.YOffset()
				absRow += fieldHeights[f]
			}
			b.WriteString(m.Viewport.View())
		} else {
			// WindowH not yet known, render all fields directly
			for f := range fieldCount {
				m.FieldY[f] = 1 + absRow
				absRow += fieldHeights[f]
			}
			b.WriteString(content)
		}
	}

	// Error message (if any)
	if m.ErrorMsg != "" {
		b.WriteString("\n" + common.ErrorStyle.Render("Error: "+m.ErrorMsg) + "\n")
	}

	m.HelpLineY = strings.Count(b.String(), "\n") + 1
	layout := common.BuildHelpLineLayout(targetHelpItems, targetHelpPrefix, maxWidth)
	b.WriteString("\n" + common.RenderHelpLine(layout, targetHelpPrefix, maxWidth, m.HoveredHelpItem))

	view := b.String()
	if m.ShowHelp {
		return renderHelpOverlay(view, *m, maxWidth)
	}

	return view
}

func renderField(m *Model, field int, maxWidth int) string {
	var b strings.Builder
	focused := m.FocusedField == field

	titleStyle := fieldTitleBlurred
	if focused {
		titleStyle = fieldTitleFocused
	}

	switch field {
	case fieldIP:
		b.WriteString(titleStyle.Render("IP address") + "\n")
		b.WriteString(m.IPTextInput.View() + "\n")
	case fieldCIDR:
		b.WriteString(titleStyle.Render("CIDR (16-32)") + "\n")
		b.WriteString(m.CIDRTextInput.View() + "\n")
		b.WriteString(fieldDescStyle.Render(hostCountDesc(m.CIDRInput)) + "\n")
	}

	return b.String()
}

func renderPortModeField(m *Model) string {
	var b strings.Builder
	focused := m.FocusedField == fieldPortMode

	titleStyle := fieldTitleBlurred
	if focused {
		titleStyle = fieldTitleFocused
	}

	b.WriteString(titleStyle.Render("Ports to scan") + "\n")
	for i, opt := range portModeOptions {
		var line string
		if i == m.PortModeIndex {
			if focused {
				line = selectorStyle.Render("> ") + optionSelected.Render(opt.Label)
			} else {
				line = "  " + optionSelected.Render(opt.Label)
			}
		} else {
			line = "  " + optionNormal.Render(opt.Label)
		}
		b.WriteString(line + "\n")
	}

	return b.String()
}

func renderInterfaceField(m *Model) string {
	var b strings.Builder
	focused := m.FocusedField == fieldInterface

	titleStyle := fieldTitleBlurred
	if focused {
		titleStyle = fieldTitleFocused
	}

	b.WriteString(titleStyle.Render("Interface") + "\n")

	var label string
	if m.IPIsCustom {
		label = lipgloss.NewStyle().Foreground(common.Color.Help).Italic(true).Render("custom")
	} else if m.IPIndex >= 0 && m.IPIndex < len(m.InterfaceInfos) {
		info := m.InterfaceInfos[m.IPIndex]
		label = optionSelected.Render(info.Name)
	}

	if focused {
		b.WriteString(selectorStyle.Render("< ") + label + selectorStyle.Render(" >") + "\n")
	} else {
		b.WriteString("  " + label + "\n")
	}

	count := len(m.InterfaceInfos)
	desc := fmt.Sprintf("%d/%d — ←/→ to cycle", m.IPIndex+1, count)
	if count == 1 {
		desc = "1 interface"
	}
	b.WriteString(fieldDescStyle.Render(desc) + "\n")

	return b.String()
}

func hostCountDesc(cidrStr string) string {
	isDefault := false
	if cidrStr == "" {
		cidrStr = "32"
		isDefault = true
	}

	cidr := 0
	_, err := fmt.Sscanf(cidrStr, "%d", &cidr)
	if err != nil || cidr < 16 || cidr > 32 {
		return "targets: -"
	}

	_, ipnet, err := net.ParseCIDR("0.0.0.0/" + cidrStr)
	if err != nil {
		return "targets: -"
	}

	hosts := shared.TotalScanHosts(ipnet)
	if isDefault {
		return fmt.Sprintf("targets: %d (default /32)", hosts)
	}
	return fmt.Sprintf("targets: %d", hosts)
}
