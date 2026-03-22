package scanview

import (
	tea "charm.land/bubbletea/v2"
	"github.com/backendsystems/nibble/internal/tui/views/common"
)

func (m Model) HandleMouse(msg tea.Msg, maxWidth int) Result {
	result := Result{Model: m}

	mouse, ok := common.MouseXY(msg)
	if !ok {
		return result
	}

	if !m.Scanning {
		return result
	}

	result.Handled = true

	hlr := common.HandleHelpLineMouse(mouse, msg, scanHelpItems, scanHelpPrefix, m.HelpLineY, maxWidth)
	result.Model.HoveredHelpItem = hlr.NewHoveredItem

	// Handle wheel scroll on the results viewport
	if _, ok := msg.(tea.MouseWheelMsg); ok {
		switch mouse.Button {
		case tea.MouseWheelUp:
			result.Model.Results.SetYOffset(max(0, result.Model.Results.YOffset()-1))
			result.Cmd = continueScanLoop(result.Model)
			return result
		case tea.MouseWheelDown:
			result.Model.Results.SetYOffset(result.Model.Results.YOffset() + 1)
			result.Cmd = continueScanLoop(result.Model)
			return result
		}
	}

	if hlr.Consumed {
		switch Action(hlr.ClickedAction) {
		case ActionQuit:
			result.Model = prepareForExit(result.Model, true)
			result.Model.Scanning = false
			result.Model.ScanComplete = true
			result.Cmd = tea.Batch(m.Stopwatch.Stop(), sendQuitMsg())
		}
		return result
	}

	return result
}
