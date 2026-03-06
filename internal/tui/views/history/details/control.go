package historydetailview

import (
	"github.com/backendsystems/nibble/internal/history"
	detailsscan "github.com/backendsystems/nibble/internal/tui/views/history/details/scan"
	"github.com/charmbracelet/bubbles/stopwatch"
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
	ActionDelete
)

type UpdateResult struct {
	Model           Model
	Quit            bool
	ScanAllPorts    bool
	SelectedHostIP  string
	ScanHistoryPath string
	Deleted         bool
	Cmd             tea.Cmd
}

// SavedMsg is sent after the background save+reload finishes
type SavedMsg struct {
	Updated history.ScanHistory
}

func (m Model) Update(msg tea.Msg) UpdateResult {
	result := UpdateResult{Model: m}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return handleKeyMsg(m, msg)
	case stopwatch.TickMsg, stopwatch.StartStopMsg:
		var tickCmd tea.Cmd
		result.Model.Stopwatch, tickCmd = m.Stopwatch.Update(msg)
		result.Cmd = tea.Batch(tickCmd, detailsscan.Continue(result.Model.ProgressChan))
		return result
	case detailsscan.ProgressMsg:
		progress := detailsscan.ApplyProgressUpdate(detailsscan.UpdateInput{
			Update:           msg.Update,
			History:          result.Model.History,
			ScanningHostIdx:  result.Model.ScanningHostIdx,
			ScanPortsScanned: result.Model.ScanPortsScanned,
			NewPortsByHost:   result.Model.NewPortsByHost,
			TotalHosts:       result.Model.TotalHosts,
			ScannedCount:     result.Model.ScannedCount,
			ScannedHostStr:   result.Model.ScannedHostStr,
		})
		result.Model.History = progress.History
		result.Model.NewPortsByHost = progress.NewPortsByHost
		result.Model.TotalHosts = progress.TotalHosts
		result.Model.ScannedCount = progress.ScannedCount
		result.Model.ScannedHostStr = progress.ScannedHostStr
		result.Cmd = detailsscan.Continue(result.Model.ProgressChan)
		return result
	case detailsscan.CompleteMsg:
		return handleScanComplete(m)
	case SavedMsg:
		if msg.Updated.Version != "" {
			result.Model.History = msg.Updated
		}
		return result
	}

	// Update viewport for scrolling
	var cmd tea.Cmd
	result.Model.Viewport, cmd = m.Viewport.Update(msg)
	_ = cmd

	return result
}

func handleScanComplete(m Model) UpdateResult {
	result := UpdateResult{Model: m}
	result.Model.Scanning = false
	result.Model.ProgressChan = nil
	result.Model.ScannedHostStr = ""

	// Kick off background save + reload so disk I/O doesn't block the UI
	if m.HistoryPath != "" {
		histPath := m.HistoryPath
		hostStr := m.ScannedHostStr
		hostIdx := m.ScanningHostIdx
		portsScanned := m.ScanPortsScanned
		hosts := m.History.ScanResults.Hosts

		result.Cmd = tea.Batch(m.Stopwatch.Stop(), func() tea.Msg {
			if updated, err := detailsscan.PersistAndReload(histPath, hostStr, hostIdx, portsScanned, hosts); err == nil {
				return SavedMsg{Updated: updated}
			}
			return SavedMsg{}
		})
	} else {
		result.Cmd = m.Stopwatch.Stop()
	}

	return result
}
