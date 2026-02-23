package targetview

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/backendsystems/nibble/internal/ports"
	"github.com/backendsystems/nibble/internal/scanner/shared"
	"github.com/backendsystems/nibble/internal/tui/views/common"
	"github.com/charmbracelet/lipgloss"
)

func Render(m Model, maxWidth int) string {
	var b strings.Builder

	b.WriteString(common.TitleStyle.Render("Custom Target") + "\n")

	// IP field
	ipLabel := "IP:   "
	ipContent := m.IPInput
	if m.FocusField == 0 {
		ipContent = withCursor(m.IPInput, m.IPCursor)
	}
	ipStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	if m.FocusField == 0 {
		ipStyle = ipStyle.Foreground(lipgloss.Color("226")).Bold(true)
	}
	b.WriteString(ipStyle.Render(ipLabel+ipContent) + "\n")

	// CIDR field with host count
	cidrLabel := "CIDR: "
	cidrContent := m.CIDRInput
	if m.CIDRInput == "" {
		cidrContent = "32"
	}
	if m.FocusField == 1 {
		cidrContent = withCursor(cidrContent, m.CIDRCursor)
	}
	cidrStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	if m.FocusField == 1 {
		cidrStyle = cidrStyle.Foreground(lipgloss.Color("226")).Bold(true)
	}

	// Calculate host count for display
	hostCount := getHostCount(m.IPInput, cidrContent)
	hostCountStr := ""
	if hostCount > 0 {
		hostCountStr = fmt.Sprintf(" (hosts: %d)", hostCount)
	}

	b.WriteString(cidrStyle.Render(cidrLabel+cidrContent+hostCountStr) + "\n")

	// Ports section
	b.WriteString("\n")
	defaultStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	allStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	customStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	if m.PortPack == "default" {
		defaultStyle = defaultStyle.Foreground(lipgloss.Color("226")).Bold(true)
	} else if m.PortPack == "all" {
		allStyle = allStyle.Foreground(lipgloss.Color("226")).Bold(true)
	} else {
		customStyle = customStyle.Foreground(lipgloss.Color("226")).Bold(true)
	}

	defaultLine := wrapPortList("default: ", formatPortList(ports.DefaultPorts()), maxWidth)
	b.WriteString(defaultStyle.Render(defaultLine) + "\n")
	allLine := wrapPortList("all:     ", "1-65535", maxWidth)
	b.WriteString(allStyle.Render(allLine) + "\n")
	customContent := m.CustomPorts
	if m.FocusField == 2 {
		customContent = withCursor(m.CustomPorts, m.PortCursor)
	}
	customLine := wrapPortList("custom:  ", customContent, maxWidth)
	b.WriteString(customStyle.Render(customLine) + "\n")

	b.WriteString("\n")

	// Error message
	if m.ErrorMsg != "" {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
		b.WriteString(errorStyle.Render("Error: "+m.ErrorMsg) + "\n")
		b.WriteString("\n")
	}

	// Help text
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	b.WriteString(helpStyle.Render(common.WrapWords(helpText, maxWidth)))

	view := b.String()
	if m.ShowHelp {
		return renderHelpOverlay(view)
	}
	return view
}

// getHostCount calculates the number of usable hosts for a given CIDR
func getHostCount(ip, cidrStr string) int {
	if ip == "" || cidrStr == "" {
		return 0
	}

	// Parse CIDR value
	cidr := 0
	_, err := fmt.Sscanf(cidrStr, "%d", &cidr)
	if err != nil || cidr < 16 || cidr > 32 {
		return 0
	}

	// Build full CIDR notation
	fullCIDR := ip + "/" + cidrStr
	_, ipnet, err := net.ParseCIDR(fullCIDR)
	if err != nil {
		return 0
	}

	return shared.TotalScanHosts(ipnet)
}

func withCursor(s string, cursor int) string {
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(s) {
		cursor = len(s)
	}
	return s[:cursor] + "|" + s[cursor:]
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

func wrapPortList(prefix, content string, maxWidth int) string {
	if content == "" {
		return prefix
	}
	if maxWidth <= len(prefix)+1 {
		return prefix + content
	}

	indent := strings.Repeat(" ", len(prefix))
	tokens := strings.Split(content, ",")
	lines := make([]string, 0, 4)
	current := prefix

	for i, token := range tokens {
		segment := token
		if i < len(tokens)-1 {
			segment += ","
		}
		if len(current)+len(segment) > maxWidth && len(current) > len(prefix) {
			lines = append(lines, current)
			current = indent + segment
			continue
		}
		current += segment
	}

	lines = append(lines, current)
	return strings.Join(lines, "\n")
}
