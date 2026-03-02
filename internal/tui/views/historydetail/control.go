package historydetailview

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Action int

const (
	ActionNone Action = iota
	ActionQuit
	ActionMoveUp
	ActionMoveDown
	ActionScanAllPorts
	ActionHelp
)

type UpdateResult struct {
	Model          Model
	Quit           bool
	ScanAllPorts   bool
	SelectedHostIP string
}

func HandleKey(key string) Action {
	switch key {
	case "q", "esc":
		return ActionQuit
	case "up", "k":
		return ActionMoveUp
	case "down", "j":
		return ActionMoveDown
	case "enter":
		return ActionScanAllPorts
	case "?":
		return ActionHelp
	default:
		return ActionNone
	}
}

func (m Model) Update(msg tea.Msg) UpdateResult {
	result := UpdateResult{Model: m}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return handleKeyMsg(m, msg)
	}

	// Update viewport for scrolling
	var cmd tea.Cmd
	result.Model.Viewport, cmd = m.Viewport.Update(msg)
	_ = cmd

	return result
}

func handleKeyMsg(m Model, key tea.KeyMsg) UpdateResult {
	result := UpdateResult{Model: m}

	switch HandleKey(key.String()) {
	case ActionQuit:
		result.Quit = true
		return result
	case ActionMoveUp:
		if m.Cursor > 0 {
			result.Model.Cursor--
		}
	case ActionMoveDown:
		if m.Cursor < len(m.History.ScanResults.Hosts)-1 {
			result.Model.Cursor++
		}
	case ActionScanAllPorts:
		if m.Cursor < len(m.History.ScanResults.Hosts) {
			result.ScanAllPorts = true
			result.SelectedHostIP = m.History.ScanResults.Hosts[m.Cursor].IP
		}
	case ActionHelp:
		result.Model.ShowHelp = !result.Model.ShowHelp
	}

	// Update viewport for scrolling
	var cmd tea.Cmd
	result.Model.Viewport, cmd = m.Viewport.Update(key)
	_ = cmd

	return result
}
