package historydetailview

import (
	"strconv"
	"strings"
	"time"

	"github.com/backendsystems/nibble/internal/history"
	"github.com/backendsystems/nibble/internal/scanner/shared"
	deletepkg "github.com/backendsystems/nibble/internal/tui/views/history/delete"
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

// Exported message types for scan progress
type ProgressMsg struct {
	Update shared.ProgressUpdate
}

type CompleteMsg struct{}

func HandleKey(key string) Action {
	switch key {
	case "q", "esc":
		return ActionQuit
	case "up", "w", "k":
		return ActionMoveUp
	case "down", "s", "j":
		return ActionMoveDown
	case "enter":
		return ActionScanAllPorts
	case "delete":
		return ActionDelete
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
	case stopwatch.TickMsg, stopwatch.StartStopMsg:
		var tickCmd tea.Cmd
		result.Model.Stopwatch, tickCmd = m.Stopwatch.Update(msg)
		result.Cmd = tea.Batch(tickCmd, continueScanLoop(result.Model))
		return result
	case ProgressMsg:
		return handleProgressMsg(m, msg)
	case CompleteMsg:
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

func handleKeyMsg(m Model, key tea.KeyMsg) UpdateResult {
	result := UpdateResult{Model: m}

	// Handle delete dialog in detail view
	if m.DeleteDialog != nil {
		switch key.String() {
		case "left", "a", "h", "right", "d", "l":
			// Toggle between Delete and Cancel
			result.Model.DeleteDialog.Toggle()
			return result
		case "enter":
			// User pressed Enter - execute the selected action
			if result.Model.DeleteDialog.IsDeleteSelected() {
				// Delete was selected
				performDeleteSync(m.NodePath)
				result.Deleted = true
			}
			// Close dialog (whether Delete or Cancel was selected)
			result.Model.DeleteDialog = nil
			if result.Deleted {
				result.Quit = true
			}
			return result
		default:
			// Any other key closes the dialog and returns to detail view
			result.Model.DeleteDialog = nil
			return result
		}
	}

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
		result.Quit = true
		return result
	case ActionMoveUp:
		if !m.Scanning && m.Cursor > 0 {
			result.Model.Cursor--
		}
	case ActionMoveDown:
		if !m.Scanning && m.Cursor < len(m.History.ScanResults.Hosts)-1 {
			result.Model.Cursor++
		}
	case ActionScanAllPorts:
		if m.Cursor < len(m.History.ScanResults.Hosts) {
			result.ScanAllPorts = true
			result.SelectedHostIP = m.History.ScanResults.Hosts[m.Cursor].IP
			result.ScanHistoryPath = m.HistoryPath
			result.Model.ScanningHostIdx = m.Cursor // Track which host is being scanned
		}
	case ActionDelete:
		if m.NodePath != "" {
			result.Model.DeleteDialog = &deletepkg.HistoryDeleteDialog{
				Target:      nil,
				ItemType:    m.NodeItemType,
				ItemName:    m.NodeName,
				CursorOnYes: true,
			}
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

func performDeleteSync(path string) {
	if path != "" {
		history.Delete(path)
	}
}

func handleProgressMsg(m Model, msg ProgressMsg) UpdateResult {
	result := UpdateResult{Model: m}

	if p, ok := msg.Update.(shared.SweepProgress); ok {
		if p.TotalHosts > 0 {
			result.Model.TotalHosts = p.TotalHosts
		}
		result.Model.ScannedCount = p.Scanned
		if p.Host != "" {
			result.Model.ScannedHostStr = p.Host
		}
	}

	result.Cmd = continueScanLoop(result.Model)
	return result
}

// SavedMsg is sent after the background save+reload finishes
type SavedMsg struct {
	Updated history.ScanHistory
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
			if hostStr != "" && hostIdx < len(hosts) {
				scannedHost := hosts[hostIdx]
				newHost := parseScanHostStr(hostStr, scannedHost.IP, portsScanned)
				_ = history.UpdateHostInScan(histPath, scannedHost.IP, newHost)
			}
			if updated, err := history.Load(histPath); err == nil {
				return SavedMsg{Updated: updated}
			}
			return SavedMsg{}
		})
	} else {
		result.Cmd = m.Stopwatch.Stop()
	}

	return result
}

// parseScanHostStr converts a FormatHost string back to a HostResult
func parseScanHostStr(hostStr string, fallbackIP string, portsScanned []int) history.HostResult {
	lines := strings.Split(hostStr, "\n")
	ip := fallbackIP
	hardware := ""
	if len(lines) > 0 {
		first := lines[0]
		if idx := strings.Index(first, " - "); idx != -1 {
			ip = strings.TrimSpace(first[:idx])
			hardware = strings.TrimSpace(first[idx+3:])
		} else {
			ip = strings.TrimSpace(first)
		}
	}

	var ports []history.PortInfo
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "port ") {
			continue
		}
		line = strings.TrimPrefix(line, "port ")
		portStr, banner, _ := strings.Cut(line, ":")
		portNum, err := strconv.Atoi(strings.TrimSpace(portStr))
		if err != nil {
			continue
		}
		ports = append(ports, history.PortInfo{Port: portNum, Banner: strings.TrimSpace(banner)})
	}

	return history.HostResult{
		IP:           ip,
		Hardware:     hardware,
		Ports:        ports,
		LastScanned:  time.Now(),
		PortsScanned: portsScanned,
	}
}

// ListenForProgress is exported for use by the main controller
func ListenForProgress(progressChan <-chan shared.ProgressUpdate) tea.Cmd {
	return func() tea.Msg {
		progress, ok := <-progressChan
		if !ok {
			return CompleteMsg{}
		}
		return ProgressMsg{Update: progress}
	}
}

func continueScanLoop(m Model) tea.Cmd {
	if m.ProgressChan == nil {
		return nil
	}
	return ListenForProgress(m.ProgressChan)
}
