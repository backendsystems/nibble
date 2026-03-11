package portsview

import (
	"strings"

	"github.com/backendsystems/nibble/internal/ports"
	tea "github.com/charmbracelet/bubbletea"
)

const portsTitleRows = 1 // "Configure Scan Ports"

// defaultRowCount returns how many terminal lines the default port list occupies
// at the given window width. Used to compute the Y position of the custom row.
func defaultRowCount(maxWidth int) int {
	line := wrapPortList("default: ", formatPortList(ports.DefaultPorts()), maxWidth)
	return strings.Count(line, "\n") + 1
}

// HandleMouse processes a mouse event for the ports view.
// Clicking the default or custom row selects it; clicking the active row applies.
func (m Model) HandleMouse(msg tea.MouseMsg, maxWidth int) Result {
	result := Result{Model: m}
	if msg.Button != tea.MouseButtonLeft || msg.Action != tea.MouseActionRelease {
		return result
	}
	if m.ShowHelp {
		result.Model.ShowHelp = false
		return result
	}

	defaultStart := portsTitleRows
	defaultEnd := defaultStart + defaultRowCount(maxWidth) - 1
	customRow := defaultEnd + 1

	switch {
	case msg.Y >= defaultStart && msg.Y <= defaultEnd:
		if m.PortPack == "default" {
			next, ok := applyConfig(result.Model)
			result.Model = next
			result.Done = ok
		} else {
			result.Model.PortPack = "default"
			var cmd tea.Cmd
			result.Model, cmd = Prepare(result.Model)
			result.Cmd = cmd
		}
	case msg.Y == customRow:
		if m.PortPack == "custom" {
			next, ok := applyConfig(result.Model)
			result.Model = next
			result.Done = ok
		} else {
			result.Model.PortPack = "custom"
			var cmd tea.Cmd
			result.Model, cmd = Prepare(result.Model)
			result.Cmd = cmd
		}
	}

	return result
}
