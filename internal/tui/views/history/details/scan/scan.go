package scan

import (
	"time"

	"github.com/backendsystems/nibble/internal/history"
	"github.com/backendsystems/nibble/internal/scanner/shared"
)

type LiveUpdateInput struct {
	History          history.ScanHistory
	ScanningHostIdx  int
	ScanPortsScanned []int
	NewPortsByHost   map[string]map[int]bool
	Host             *shared.HostResult
}

type LiveUpdateResult struct {
	History        history.ScanHistory
	NewPortsByHost map[string]map[int]bool
	Updated        bool
}

func ApplyLiveHostUpdate(in LiveUpdateInput) LiveUpdateResult {
	result := LiveUpdateResult{
		History:        in.History,
		NewPortsByHost: in.NewPortsByHost,
	}

	hostIdx := in.ScanningHostIdx
	if hostIdx < 0 || hostIdx >= len(result.History.ScanResults.Hosts) {
		return result
	}

	current := result.History.ScanResults.Hosts[hostIdx]
	updated := toHistoryHost(in.Host, current.IP, in.ScanPortsScanned)
	if updated.Hardware == "" {
		updated.Hardware = current.Hardware
	}
	updated.MAC = current.MAC

	existingPorts := make(map[int]struct{}, len(current.Ports))
	for _, port := range current.Ports {
		existingPorts[port.Port] = struct{}{}
	}

	if result.NewPortsByHost == nil {
		result.NewPortsByHost = make(map[string]map[int]bool)
	}
	if result.NewPortsByHost[current.IP] == nil {
		result.NewPortsByHost[current.IP] = make(map[int]bool)
	}
	for _, port := range updated.Ports {
		if _, ok := existingPorts[port.Port]; !ok {
			result.NewPortsByHost[current.IP][port.Port] = true
		}
	}

	result.History.ScanResults.Hosts[hostIdx] = updated
	result.History.ScanMetadata.Updated = time.Now()
	totalPorts := 0
	for _, host := range result.History.ScanResults.Hosts {
		totalPorts += len(host.Ports)
	}
	result.History.ScanResults.PortsFound = totalPorts
	result.Updated = true
	return result
}

// toHistoryHost converts a shared.HostResult to a history.HostResult.
func toHistoryHost(h *shared.HostResult, fallbackIP string, portsScanned []int) history.HostResult {
	if h == nil {
		return history.HostResult{
			IP:           fallbackIP,
			LastScanned:  time.Now(),
			PortsScanned: portsScanned,
		}
	}

	ip := h.IP
	if ip == "" {
		ip = fallbackIP
	}

	var ports []history.PortInfo
	for _, p := range h.Ports {
		ports = append(ports, history.PortInfo{Port: p.Port, Banner: p.Banner})
	}

	return history.HostResult{
		IP:           ip,
		Hardware:     h.Hardware,
		Ports:        ports,
		LastScanned:  time.Now(),
		PortsScanned: portsScanned,
	}
}

func PersistAndReload(historyPath string, host *shared.HostResult, hostIdx int, portsScanned []int, hosts []history.HostResult) (history.ScanHistory, error) {
	if hostIdx < len(hosts) {
		scannedHost := hosts[hostIdx]
		newHost := toHistoryHost(host, scannedHost.IP, portsScanned)
		if host == nil {
			newHost = history.HostResult{
				IP:           scannedHost.IP,
				Hardware:     scannedHost.Hardware,
				MAC:          scannedHost.MAC,
				Ports:        scannedHost.Ports,
				LastScanned:  time.Now(),
				PortsScanned: portsScanned,
			}
		}
		_ = history.UpdateHostInScan(historyPath, scannedHost.IP, newHost)
	}

	return history.Load(historyPath)
}
