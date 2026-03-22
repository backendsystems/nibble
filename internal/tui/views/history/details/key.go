package historydetailview

import (
	tea "charm.land/bubbletea/v2"
	"github.com/backendsystems/nibble/internal/scanner/shared"
)

func HandleKey(key string) Action {
	switch key {
	case "q", "esc", "left", "a", "h":
		return ActionQuit
	case "up", "w", "k":
		return ActionMoveUp
	case "down", "s", "j":
		return ActionMoveDown
	case "enter", "right", "d", "l":
		return ActionScanAllPorts
	case "?":
		return ActionHelp
	default:
		return ActionNone
	}
}

func handleKeyMsg(m Model, key tea.KeyPressMsg) UpdateResult {
	result := UpdateResult{Model: m}

	// Accept any key to close help overlay (except ? which toggles help)
	if m.ShowHelp && key.String() != "?" {
		result.Model.ShowHelp = false
		// Update viewport for scrolling
		var cmd tea.Cmd
		result.Model.Viewport, cmd = m.Viewport.Update(key)
		_ = cmd
		return result
	}

	switch HandleKey(key.String()) {
	case ActionQuit:
		if m.Scanning {
			result.Model.Scanning = false
			result.Model.ProgressChan = nil
			result.Model.ScannedHostStr = ""
			result.Cmd = tea.Batch(
				m.Stopwatch.Stop(),
				drainProgressChan(m.ProgressChan),
			)
		}
		result.Quit = true
		return result
	case ActionMoveUp:
		if !m.Scanning && m.Cursor > 0 {
			result.Model.Cursor--
			result.Model = result.Model.ScrollToSelected()
		}
	case ActionMoveDown:
		if !m.Scanning && m.Cursor < len(m.History.ScanResults.Hosts)-1 {
			result.Model.Cursor++
			result.Model = result.Model.ScrollToSelected()
		}
	case ActionScanAllPorts:
		if m.Cursor < len(m.History.ScanResults.Hosts) {
			result.ScanAllPorts = true
			result.SelectedHostIP = m.History.ScanResults.Hosts[m.Cursor].IP
			result.ScanHistoryPath = m.HistoryPath
			result.Model.ScanningHostIdx = m.Cursor // Track which host is being scanned
		}
	case ActionHelp:
		result.Model.ShowHelp = !result.Model.ShowHelp
	}

	return result
}

func drainProgressChan(ch <-chan shared.ProgressUpdate) tea.Cmd {
	return func() tea.Msg {
		if ch == nil {
			return nil
		}
		go func() {
			for range ch {
			}
		}()
		return nil
	}
}

