package historydetailview

import (
	"fmt"

	"github.com/backendsystems/nibble/internal/tui/views/common"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// hostAtViewportLine returns the host index that occupies the given line within
// the viewport content, or -1 if the line falls on the metadata line.
// Line 0 = metadata (ports found / created date); hosts start at line 1.
func hostAtViewportLine(m Model, viewportLine int) int {
	if viewportLine <= 0 {
		return -1
	}
	line := 1
	for i, host := range m.History.ScanResults.Hosts {
		end := line + 1 + len(host.Ports)
		if viewportLine < end {
			return i
		}
		line = end
	}
	return -1
}

// detailTitleLines returns how many terminal lines the detail title occupies.
// Mirrors the wrap logic in render.go exactly.
func detailTitleLines(m Model, windowWidth int) int {
	cidr := m.History.ScanMetadata.TargetCIDR
	iface := m.History.ScanMetadata.InterfaceName
	date := m.History.ScanMetadata.Created.Format("2006 Jan 2 15:04")
	titleFull := common.TitleStyle.Render(fmt.Sprintf("%s - %s - %s", cidr, iface, date))
	total := len(m.History.ScanResults.Hosts)
	counterW := 0
	if total > 0 {
		counterW = lipgloss.Width(common.MutedStyle.Render(fmt.Sprintf("%d/%d", m.Cursor+1, total)))
	}
	if lipgloss.Width(titleFull)+counterW+1 > windowWidth {
		return 2
	}
	return 1
}

// HandleMouse processes mouse events in the detail host list:
// scroll wheel scrolls the viewport; left-click selects or activates a host.
func (m Model) HandleMouse(msg tea.MouseMsg) UpdateResult {
	result := UpdateResult{Model: m}

	switch msg.Button {
	case tea.MouseButtonWheelUp:
		result.Model.Viewport.YOffset = max(0, result.Model.Viewport.YOffset-3)
		return result
	case tea.MouseButtonWheelDown:
		result.Model.Viewport.YOffset += 3
		return result
	}

	if msg.Button != tea.MouseButtonLeft || msg.Action != tea.MouseActionRelease {
		return result
	}
	if m.ShowHelp || m.DeleteDialog != nil || m.Scanning {
		return result
	}

	titleLines := detailTitleLines(m, m.WindowW)
	contentY := msg.Y - titleLines
	if contentY < 0 {
		return result
	}

	viewportLine := m.Viewport.YOffset + contentY
	index := hostAtViewportLine(m, viewportLine)
	if index < 0 {
		return result
	}

	if index == m.Cursor {
		// Second click on same host: scan all ports
		result.ScanAllPorts = true
		result.SelectedHostIP = m.History.ScanResults.Hosts[index].IP
		result.ScanHistoryPath = m.HistoryPath
		result.Model.ScanningHostIdx = index
		return result
	}

	result.Model.Cursor = index
	return result
}
