package portsview

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/backendsystems/nibble/internal/ports"
	"github.com/backendsystems/nibble/internal/tui/views/common"
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
func (m Model) HandleMouse(msg tea.Msg, maxWidth int) Result {
	result := Result{Model: m}

	mouse, ok := mouseXY(msg)
	if !ok {
		return result
	}

	helpLineY := m.HelpLineY
	helpLayout := common.BuildHelpLineLayout(portsHelpItems, portsHelpPrefix, maxWidth)
	helpLineEndY := helpLineY + helpLayout.LineCount - 1

	// Update hover state for all mouse events
	if helpLineY > 0 && mouse.Y >= helpLineY && mouse.Y <= helpLineEndY {
		result.Model.HoveredHelpItem = common.GetHelpItemAt(helpLayout, mouse.X, mouse.Y-helpLineY)
	} else {
		result.Model.HoveredHelpItem = -1
	}

	if _, ok := msg.(tea.MouseReleaseMsg); !ok || mouse.Button != tea.MouseLeft {
		return result
	}
	if m.ShowHelp {
		result.Model.ShowHelp = false
		return result
	}

	// Check if clicking on helpline item
	if helpLineY > 0 && mouse.Y >= helpLineY && mouse.Y <= helpLineEndY {
		itemIndex := common.GetHelpItemAt(helpLayout, mouse.X, mouse.Y-helpLineY)
		if itemIndex >= 0 {
			switch helpLayout.Items[itemIndex].Action {
			case portsActionBackspace:
				portAction := common.PortInputActionFromKey("backspace", false)
				var cmd tea.Cmd
				result.Model.PortInput, cmd = result.Model.PortInput.HandleKey(portAction, tea.KeyPressMsg(tea.Key{Code: tea.KeyBackspace}))
				result.Model.CustomPorts = result.Model.PortInput.Value
				result.Model.CustomCursor = result.Model.PortInput.Cursor
				result.Cmd = cmd
			case portsActionDeleteAll:
				portAction := common.PortInputActionFromKey("delete", false)
				var cmd tea.Cmd
				result.Model.PortInput, cmd = result.Model.PortInput.HandleKey(portAction, tea.KeyPressMsg(tea.Key{Code: tea.KeyDelete}))
				result.Model.CustomPorts = result.Model.PortInput.Value
				result.Model.CustomCursor = result.Model.PortInput.Cursor
				result.Cmd = cmd
			case portsActionApply:
				next, ok := applyConfig(result.Model)
				result.Model = next
				result.Done = ok
			case portsActionHelp:
				result.Model.ShowHelp = true
			case portsActionBack:
				result.Back = true
			}
			return result
		}
	}

	defaultStart := portsTitleRows
	defaultEnd := defaultStart + defaultRowCount(maxWidth) - 1
	customRow := defaultEnd + 1

	switch {
	case mouse.Y >= defaultStart && mouse.Y <= defaultEnd:
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
	case mouse.Y == customRow:
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

// mouseXY extracts Mouse data from any mouse message type.
func mouseXY(msg tea.Msg) (tea.Mouse, bool) {
	if mm, ok := msg.(tea.MouseMsg); ok {
		return mm.Mouse(), true
	}
	return tea.Mouse{}, false
}
