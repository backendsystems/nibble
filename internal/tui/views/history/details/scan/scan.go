package scan

import (
	"strconv"
	"strings"
	"time"

	"github.com/backendsystems/nibble/internal/history"
)

type LiveUpdateInput struct {
	History          history.ScanHistory
	ScanningHostIdx  int
	ScanPortsScanned []int
	NewPortsByHost   map[string]map[int]bool
	HostStr          string
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
	updated := ParseHost(in.HostStr, current.IP, in.ScanPortsScanned)
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

// ParseHost converts a FormatHost string back to a HostResult.
func ParseHost(hostStr string, fallbackIP string, portsScanned []int) history.HostResult {
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

func PersistAndReload(historyPath string, hostStr string, hostIdx int, portsScanned []int, hosts []history.HostResult) (history.ScanHistory, error) {
	if hostIdx < len(hosts) {
		scannedHost := hosts[hostIdx]
		newHost := ParseHost(hostStr, scannedHost.IP, portsScanned)
		if hostStr == "" {
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
