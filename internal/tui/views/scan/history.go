package scanview

import (
	"strings"
	"time"

	"github.com/backendsystems/nibble/internal/history"
	"github.com/backendsystems/nibble/internal/scanner/demo"
	"github.com/backendsystems/nibble/internal/scanner/ip4"
	"github.com/backendsystems/nibble/internal/scanner/shared"
)

// toHistoryHost converts a shared.HostResult to a history.HostResult.
func toHistoryHost(h shared.HostResult, portsScanned []int) history.HostResult {
	var ports []history.PortInfo
	for _, p := range h.Ports {
		ports = append(ports, history.PortInfo{
			Port:   p.Port,
			Banner: p.Service,
		})
	}
	return history.HostResult{
		IP:           h.IP,
		Hardware:     h.Hardware,
		Ports:        ports,
		LastScanned:  time.Now(),
		PortsScanned: portsScanned,
	}
}

// SaveHistory saves the scan results to history or updates existing history
func (m Model) SaveHistory() error {
	if !m.ScanComplete {
		return nil
	}

	// If this is a rescan from history, update the existing file
	if m.IsRescan && m.RescanHistoryPath != "" {
		return m.updateHistoryRescan()
	}

	hosts := m.FinalHosts
	if len(hosts) == 0 {
		hosts = m.FoundHosts
	}

	portsScanned := m.PortsScanned
	if portsScanned == nil {
		switch s := m.NetworkScan.(type) {
		case *ip4.Scanner:
			portsScanned = s.Ports
		case *demo.Scanner:
			portsScanned = s.Ports
		}
	}

	hostResults := make([]history.HostResult, 0, len(hosts))
	for _, h := range hosts {
		hostResults = append(hostResults, toHistoryHost(h, portsScanned))
	}

	now := time.Now()
	duration := m.Stopwatch.Elapsed().Seconds()

	scanHistory := history.ScanHistory{
		Version: "1.0",
		ScanMetadata: history.ScanMetadata{
			Created:         now,
			Updated:         now,
			DurationSeconds: duration,
			InterfaceName:   interfaceName(m),
			TargetCIDR:      m.TargetCIDR,
			PortsScanned:    portsScanned,
		},
		ScanResults: history.ScanResults{
			TotalHostsScanned: m.TotalHosts,
			HostsFound:        len(hostResults),
			PortsFound:        totalPortsFound(hostResults),
			Hosts:             hostResults,
		},
	}

	return history.Save(scanHistory)
}

func interfaceName(m Model) string {
	if m.SelectedIface.Name != "" {
		return m.SelectedIface.Name
	}
	return "target"
}

func totalPortsFound(hosts []history.HostResult) int {
	total := 0
	for _, h := range hosts {
		total += len(h.Ports)
	}
	return total
}

// updateHistoryRescan updates an existing history file with rescan results
func (m Model) updateHistoryRescan() error {
	hosts := m.FinalHosts
	if len(hosts) == 0 {
		hosts = m.FoundHosts
	}

	portsScanned := m.PortsScanned
	if portsScanned == nil {
		portsScanned = getScannedPorts(m.NetworkScan)
	}

	now := time.Now()

	if len(hosts) == 0 {
		hostIP, _, _ := strings.Cut(m.TargetCIDR, "/")
		hostIP = strings.TrimSpace(hostIP)
		if hostIP == "" {
			return nil
		}

		existingHost := history.HostResult{
			IP: hostIP,
		}
		if existing, err := history.Load(m.RescanHistoryPath); err == nil {
			for _, host := range existing.ScanResults.Hosts {
				if host.IP == hostIP {
					existingHost = host
					break
				}
			}
		}

		return history.UpdateHostInScan(m.RescanHistoryPath, hostIP, history.HostResult{
			IP:           hostIP,
			Hardware:     existingHost.Hardware,
			MAC:          existingHost.MAC,
			Ports:        existingHost.Ports,
			LastScanned:  now,
			PortsScanned: portsScanned,
		})
	}

	h := hosts[0]
	newHost := toHistoryHost(h, portsScanned)

	return history.UpdateHostInScan(m.RescanHistoryPath, h.IP, newHost)
}
