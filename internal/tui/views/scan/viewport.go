package scanview

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"
	"github.com/backendsystems/nibble/internal/scanner/shared"
	"github.com/backendsystems/nibble/internal/tui/views/common"
)

const (
	defaultResultsWidth  = 72
	defaultResultsHeight = 10
	minResultsHeight     = 3
)

func (m Model) SetViewportSize(maxWidth, windowHeight int) Model {
	width := maxWidth
	if width <= 0 {
		width = defaultResultsWidth
	}

	height := defaultResultsHeight
	if windowHeight > 0 {
		// Header (title+network+neighbor+sweep+progress) = 5 lines
		// Footer (blank+help) = 2 lines, "N active:" label = 1 line → total 8
		reserved := 8
		if !m.Scanning {
			// No help line when not scanning
			reserved = 7
		}
		height = windowHeight - reserved
	}
	if height < minResultsHeight {
		height = minResultsHeight
	}

	if m.Results.Width() == 0 || m.Results.Height() == 0 {
		m.Results = viewport.New(viewport.WithWidth(width), viewport.WithHeight(height))
	} else {
		m.Results.SetWidth(width)
		m.Results.SetHeight(height)
	}

	if m.Results.PastBottom() {
		m.Results.GotoBottom()
	}
	return m
}

func (m Model) RefreshResults(stickToBottom bool) Model {
	atBottom := m.Results.AtBottom()
	m.Results.SetContent(renderHostList(m.FoundHosts))
	if stickToBottom && atBottom {
		m.Results.GotoBottom()
	}
	return m
}

func renderHostList(hosts []shared.HostResult) string {
	hostStyle := lipgloss.NewStyle().Bold(true)
	portStyle := lipgloss.NewStyle()

	var b strings.Builder
	for i, host := range hosts {
		// First line: IP or "IP - Hardware"
		header := host.IP
		if host.Hardware != "" {
			header = fmt.Sprintf("%s - %s", host.IP, host.Hardware)
		}
		b.WriteString(hostStyle.Render("• " + header))

		if i < len(hosts)-1 || len(host.Ports) > 0 {
			b.WriteString("\n")
		}
		for j, p := range host.Ports {
			var line string
			if p.Banner != "" {
				line = fmt.Sprintf("port %d: %s", p.Port, common.MutedStyle.Render(p.Banner))
			} else {
				line = fmt.Sprintf("port %d", p.Port)
			}
			b.WriteString(portStyle.Render("    " + line))
			if i < len(hosts)-1 || j < len(host.Ports)-1 {
				b.WriteString("\n")
			}
		}
	}
	return b.String()
}
