package history

import (
	"strings"

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
		if p.Host != "" && selectedHostIdx < len(history.ScanResults.Hosts) {
			selectedHost := history.ScanResults.Hosts[selectedHostIdx]
			newPorts := parseHostData(p.Host)

			// Compare against existing ports in history
			for _, port := range newPorts {
				isNew := true
				for _, existingPort := range selectedHost.Ports {
					if port == existingPort.Port {
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
					progress.NewPortsByHost[selectedHost.IP][port] = true
				}
			}
		}
	}
}

// parseHostData extracts ports from host string format "IP:port IP:port ..."
func parseHostData(hostStr string) []int {
	var ports []int
	parts := strings.Fields(hostStr)
	for _, part := range parts {
		if strings.Contains(part, ":") {
			portStr := strings.Split(part, ":")[1]
			var port int
			for _, r := range portStr {
				if r >= '0' && r <= '9' {
					port = port*10 + int(r-'0')
				} else {
					break
				}
			}
			if port > 0 {
				ports = append(ports, port)
			}
		}
	}
	return ports
}
