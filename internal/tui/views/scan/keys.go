package scanview

import tea "charm.land/bubbletea/v2"

type Action int

const (
	ActionNone Action = iota
	ActionQuit
	ActionQuitAndComplete
)

func HandleKey(scanning bool, scanComplete bool, key string) Action {
	if !scanning && !scanComplete {
		return ActionNone
	}
	if key != "ctrl+c" && key != "q" && key != "esc" {
		return ActionNone
	}
	if scanning {
		return ActionQuitAndComplete
	}
	return ActionQuit
}

func handleKeyMsg(m Model, key tea.KeyPressMsg) Result {
	result := Result{Model: m}
	result.Handled = true

	switch HandleKey(m.Scanning, m.ScanComplete, key.String()) {
	case ActionQuitAndComplete:
		result.Model = prepareForExit(result.Model, true)
		result.Model.Scanning = false
		result.Model.ScanComplete = true
		result.Cmd = tea.Batch(m.Stopwatch.Stop(), sendQuitMsg())
		return result
	case ActionQuit:
		result.Cmd = tea.Batch(m.Stopwatch.Stop(), sendQuitMsg())
		return result
	}

	// Map w/s and h/l to arrow keys so the viewport scrolls with vim-style alternatives
	switch key.String() {
	case "w", "k":
		key = tea.KeyPressMsg(tea.Key{Code: tea.KeyUp})
	case "s", "j":
		key = tea.KeyPressMsg(tea.Key{Code: tea.KeyDown})
	}

	var cmd tea.Cmd
	result.Model.Results, cmd = m.Results.Update(key)
	if cmd != nil {
		result.Cmd = tea.Batch(cmd, continueScanLoop(result.Model))
	} else {
		result.Cmd = continueScanLoop(result.Model)
	}

	return result
}
