package historydetailview

import (
	tea "charm.land/bubbletea/v2"
	"github.com/backendsystems/nibble/internal/tui/views/common"
)

// hostAtViewportLine returns the host index that occupies the given line within
// the viewport content, or -1 if the line falls on a metadata line.
// Metadata lines (1-2 lines): optional "Ports found" + "Created/Updated" date
func hostAtViewportLine(m Model, viewportLine int) int {
	// Calculate metadata lines (matches render.go and scroll.go logic)
	metadataLines := 0
	if m.History.ScanResults.PortsFound > 0 {
		metadataLines++
	}
	metadataLines++ // Created/Updated line

	if viewportLine < metadataLines {
		return -1
	}

	line := metadataLines
	for i, host := range m.History.ScanResults.Hosts {
		end := line + 1 + len(host.Ports)
		if viewportLine < end {
			return i
		}
		line = end
	}
	return -1
}

// HandleMouse processes mouse events in the detail host list:
// scroll wheel scrolls the viewport; left-click selects or activates a host.
func (m Model) HandleMouse(msg tea.Msg) UpdateResult {
	result := UpdateResult{Model: m}

	mouse, ok := common.MouseXY(msg)
	if !ok {
		return result
	}

	hlr := common.HandleHelpLineMouse(mouse, msg, detailHelpItems, detailHelpPrefix, m.HelpLineY, m.WindowW)
	result.Model.HoveredHelpItem = hlr.NewHoveredItem

	switch msg.(type) {
	case tea.MouseWheelMsg:
		if mouse.Button == tea.MouseWheelUp {
			result.Model.Viewport.SetYOffset(max(0, result.Model.Viewport.YOffset()-3))
			return result
		}
		if mouse.Button == tea.MouseWheelDown {
			// Calculate total lines: metadata + all hosts and their ports
			totalLines := 0
			// Metadata lines
			if m.History.ScanResults.PortsFound > 0 {
				totalLines++
			}
			totalLines++ // Created/Updated line
			// Host lines
			for _, host := range m.History.ScanResults.Hosts {
				totalLines += 1 + len(host.Ports)
			}
			maxOffset := max(0, totalLines-m.Viewport.Height())
			result.Model.Viewport.SetYOffset(min(result.Model.Viewport.YOffset()+3, maxOffset))
			return result
		}
	}

	if _, ok := msg.(tea.MouseReleaseMsg); !ok || mouse.Button != tea.MouseLeft {
		return result
	}
	if m.ShowHelp {
		result.Model.ShowHelp = false
		return result
	}
	if m.DeleteDialog != nil || m.Scanning {
		return result
	}

	if hlr.Consumed {
		switch Action(hlr.ClickedAction) {
		case ActionScanAllPorts:
			if len(m.History.ScanResults.Hosts) > 0 {
				result.ScanAllPorts = true
				result.SelectedHostIP = m.History.ScanResults.Hosts[m.Cursor].IP
				result.ScanHistoryPath = m.HistoryPath
				result.Model.ScanningHostIdx = m.Cursor
			}
		case ActionQuit:
			result.Quit = true
		case ActionHelp:
			result.Model.ShowHelp = true
		}
		return result
	}

	titleLines := m.HelpLineY - m.Viewport.Height()
	if titleLines < 1 {
		titleLines = 1
	}
	contentY := mouse.Y - titleLines
	if contentY < 0 {
		return result
	}

	viewportLine := m.Viewport.YOffset() + contentY
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
	result.Model = result.Model.ScrollToSelected()
	return result
}
