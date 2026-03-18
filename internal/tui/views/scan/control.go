package scanview

import (
	"github.com/backendsystems/nibble/internal/scanner/shared"
	"github.com/charmbracelet/bubbles/stopwatch"
	tea "github.com/charmbracelet/bubbletea"
)

type Result struct {
	Model   Model
	Handled bool
	Quit    bool
	Cmd     tea.Cmd
}

func (m Model) Update(msg tea.Msg) Result {
	result := Result{Model: m}

	switch typed := msg.(type) {
	case stopwatch.TickMsg, stopwatch.StartStopMsg:
		return handleStopwatchMsg(m, typed)
	case tea.KeyMsg:
		return handleKeyMsg(m, typed)
	case ProgressMsg:
		return handleProgressMsg(m, typed)
	case CompleteMsg:
		return handleCompleteMsg(m)
	case QuitMsg:
		result.Handled = true
		result.Quit = true
		return result
	default:
		return result
	}
}

func handleStopwatchMsg(m Model, msg tea.Msg) Result {
	result := Result{Model: m}
	var cmd tea.Cmd
	result.Model.Stopwatch, cmd = m.Stopwatch.Update(msg)
	result.Cmd = cmd
	result.Handled = true
	return result
}

func handleProgressMsg(m Model, msg ProgressMsg) Result {
	result := Result{Model: m}
	result.Handled = true
	hostAdded := false

	switch p := msg.Update.(type) {
	case shared.NeighborProgress:
		if p.TotalHosts > 0 {
			result.Model.TotalHosts = p.TotalHosts
		}
		result.Model.NeighborSeen = p.Seen
		result.Model.NeighborTotal = p.Total
		if p.Host != "" {
			before := len(result.Model.FoundHosts)
			result.Model.FoundHosts = appendIfNew(result.Model.FoundHosts, p.Host)
			hostAdded = len(result.Model.FoundHosts) > before
		}
	case shared.SweepProgress:
		if p.TotalHosts > 0 {
			result.Model.TotalHosts = p.TotalHosts
		}
		result.Model.ScannedCount = p.Scanned
		if p.Host != "" {
			before := len(result.Model.FoundHosts)
			result.Model.FoundHosts = appendIfNew(result.Model.FoundHosts, p.Host)
			hostAdded = len(result.Model.FoundHosts) > before
		}
	}

	if hostAdded {
		result.Model = result.Model.RefreshResults(true)
	}
	// Batch all scan operations together
	result.Cmd = continueScanLoop(result.Model)
	return result
}

func handleCompleteMsg(m Model) Result {
	result := Result{Model: m}
	result.Handled = true
	result.Model = prepareForExit(result.Model, true)
	result.Model.Scanning = false
	result.Model.ScanComplete = true

	// Save scan history
	if err := result.Model.SaveHistory(); err != nil {
		// Silently ignore history save errors - don't interrupt user flow
		_ = err
	}

	result.Cmd = tea.Batch(m.Stopwatch.Stop(), sendQuitMsg())
	return result
}
