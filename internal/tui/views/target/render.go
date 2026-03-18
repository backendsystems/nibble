package targetview

import (
	"fmt"
	"net"
	"strings"

	"github.com/backendsystems/nibble/internal/scanner/shared"
	"github.com/backendsystems/nibble/internal/tui/views/common"
	"github.com/charmbracelet/lipgloss"
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
			input.Width = available
		}
		b.WriteString(common.HighlightStyle.Render("custom:  ") + input.View() + "\n")

		guide := "  • " + common.CustomPortsDescription
		b.WriteString(common.ItalicHelpStyle.Render(guide) + "\n")
	} else {
		// Stage 1: Custom fields
		b.WriteString("\n")
		m.FieldY[fieldIP] = strings.Count(b.String(), "\n")
		b.WriteString(renderField(m, fieldIP, maxWidth))
		m.FieldY[fieldCIDR] = strings.Count(b.String(), "\n")
		b.WriteString(renderField(m, fieldCIDR, maxWidth))
		m.FieldY[fieldPortMode] = strings.Count(b.String(), "\n")
		b.WriteString(renderPortModeField(m))
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
		b.WriteString(fieldDescStyle.Render(ipsDescription(m.IPIndex, m.InterfaceInfos)) + "\n")
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

func ipsDescription(ipIndex int, ifaces []InterfaceInfo) string {
	count := len(ifaces)
	if count == 0 {
		return "No interfaces found"
	}
	if ipIndex < 0 || ipIndex >= count {
		ipIndex = 0
	}
	return fmt.Sprintf("interfaces %d/%d ←/→ [%s]", ipIndex+1, count, ifaces[ipIndex].Name)
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
