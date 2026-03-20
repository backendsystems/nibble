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

	mouse, ok := common.MouseXY(msg)
	if !ok {
		return result
	}

	hlr := common.HandleHelpLineMouse(mouse, msg, portsHelpItems, portsHelpPrefix, m.HelpLineY, maxWidth)
	result.Model.HoveredHelpItem = hlr.NewHoveredItem

	if _, ok := msg.(tea.MouseReleaseMsg); !ok || mouse.Button != tea.MouseLeft {
		return result
	}
	if m.ShowHelp {
		result.Model.ShowHelp = false
		return result
	}

	if hlr.Consumed {
		switch hlr.ClickedAction {
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
			result.Model, cmd = Init(result.Model)
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
			result.Model, cmd = Init(result.Model)
			result.Cmd = cmd
		}
	}

	return result
}
