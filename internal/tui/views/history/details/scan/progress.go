package scan

import (
	"github.com/backendsystems/nibble/internal/history"
	"github.com/backendsystems/nibble/internal/scanner/shared"
	tea "github.com/charmbracelet/bubbletea"
)

type ProgressMsg struct {
	Update shared.ProgressUpdate
}

type CompleteMsg struct{}

type UpdateInput struct {
	Update           shared.ProgressUpdate
	History          history.ScanHistory
	ScanningHostIdx  int
	ScanPortsScanned []int
	NewPortsByHost   map[string]map[int]bool
	TotalHosts       int
	ScannedCount     int
	ScannedHostStr   string
}

type UpdateResult struct {
	History        history.ScanHistory
	NewPortsByHost map[string]map[int]bool
	TotalHosts     int
	ScannedCount   int
	ScannedHostStr string
}

func ApplyProgressUpdate(in UpdateInput) UpdateResult {
	result := UpdateResult{
		History:        in.History,
		NewPortsByHost: in.NewPortsByHost,
		TotalHosts:     in.TotalHosts,
		ScannedCount:   in.ScannedCount,
		ScannedHostStr: in.ScannedHostStr,
	}

	if p, ok := in.Update.(shared.SweepProgress); ok {
		if p.TotalHosts > 0 {
			result.TotalHosts = p.TotalHosts
		}
		result.ScannedCount = p.Scanned
		if p.Host != "" {
			result.ScannedHostStr = p.Host
			live := ApplyLiveHostUpdate(LiveUpdateInput{
				History:          result.History,
				ScanningHostIdx:  in.ScanningHostIdx,
				ScanPortsScanned: in.ScanPortsScanned,
				NewPortsByHost:   result.NewPortsByHost,
				HostStr:          p.Host,
			})
			if live.Updated {
				result.History = live.History
				result.NewPortsByHost = live.NewPortsByHost
			}
		}
	}

	return result
}

func ListenForProgress(progressChan <-chan shared.ProgressUpdate) tea.Cmd {
	return func() tea.Msg {
		progress, ok := <-progressChan
		if !ok {
			return CompleteMsg{}
		}
		return ProgressMsg{Update: progress}
	}
}

func Continue(progressChan <-chan shared.ProgressUpdate) tea.Cmd {
	if progressChan == nil {
		return nil
	}
	return ListenForProgress(progressChan)
}
