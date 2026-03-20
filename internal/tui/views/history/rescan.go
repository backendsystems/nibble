package historyview

import (
	"time"

	"charm.land/bubbles/v2/stopwatch"
	tea "charm.land/bubbletea/v2"
	scannerconfig "github.com/backendsystems/nibble/internal/scanner/config"
	"github.com/backendsystems/nibble/internal/scanner/shared"
	detailsscan "github.com/backendsystems/nibble/internal/tui/views/history/details/scan"
)

func StartDetailRescan(m Model, baseScanner shared.Scanner, hostIP string) (Model, tea.Cmd) {
	targetCIDR := hostIP + "/32"

	allPorts := make([]int, 65535)
	for i := 0; i < 65535; i++ {
		allPorts[i] = i + 1
	}

	detailScanner := scannerconfig.WithPorts(baseScanner, allPorts)

	m.Details.Scanning = true
	m.Details.ProgressChan = make(chan shared.ProgressUpdate, 256)
	m.Details.NewPortsByHost = make(map[string]map[int]bool)
	m.Details.ScanPortsScanned = allPorts
	m.Details.Stopwatch = stopwatch.New(stopwatch.WithInterval(10 * time.Millisecond))
	m.Details.Stopwatch.Start()

	scanCmd := tea.Batch(
		m.Details.Stopwatch.Init(),
		detailsscan.StartNetworkScan(detailScanner, targetCIDR, m.Details.ProgressChan),
	)
	return m, scanCmd
}
