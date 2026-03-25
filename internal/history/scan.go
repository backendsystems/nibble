package history

import (
	"github.com/backendsystems/nibble/internal/scanner/shared"
)

// ScanProgress tracks the state of a live scan operation
type ScanProgress struct {
	TotalHosts     int
	ScannedCount   int
	NewPortsByHost map[string]map[int]bool // Track newly found ports per host IP
}

// UpdateScanProgress processes a scanner progress update and tracks new ports
func UpdateScanProgress(progress *ScanProgress, update shared.ProgressUpdate, history ScanHistory, selectedHostIdx int) {
	switch p := update.(type) {
	case shared.SweepProgress:
		if p.TotalHosts > 0 {
			progress.TotalHosts = p.TotalHosts
		}
		progress.ScannedCount = p.Scanned

		// Track newly found ports for the selected host
		if p.Host != nil && selectedHostIdx < len(history.ScanResults.Hosts) {
			selectedHost := history.ScanResults.Hosts[selectedHostIdx]

			// Compare against existing ports in history
			for _, port := range p.Host.Ports {
				isNew := true
				for _, existingPort := range selectedHost.Ports {
					if port.Port == existingPort.Port {
						isNew = false
						break
					}
				}
				if isNew {
					if progress.NewPortsByHost == nil {
						progress.NewPortsByHost = make(map[string]map[int]bool)
					}
					if progress.NewPortsByHost[selectedHost.IP] == nil {
						progress.NewPortsByHost[selectedHost.IP] = make(map[int]bool)
					}
					progress.NewPortsByHost[selectedHost.IP][port.Port] = true
				}
			}
		}
	}
}
