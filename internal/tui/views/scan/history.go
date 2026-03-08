package scanview

import (
	"strconv"
	"strings"
	"time"

	"github.com/backendsystems/nibble/internal/history"
	"github.com/backendsystems/nibble/internal/scanner/demo"
	"github.com/backendsystems/nibble/internal/scanner/ip4"
	"github.com/backendsystems/nibble/internal/scanner/shared"
)

// parseHostString converts a formatted host string back to structured data
func parseHostString(hostStr string) shared.HostResult {
	lines := strings.Split(hostStr, "\n")
	if len(lines) == 0 {
		return shared.HostResult{}
	}

	// First line is "IP" or "IP - Hardware"
	firstLine := lines[0]
	ip, hardware, _ := strings.Cut(firstLine, " - ")
	ip = strings.TrimSpace(ip)
	hardware = strings.TrimSpace(hardware)

	var ports []shared.PortInfo
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "port ") {
			continue
		}
		// Format is "port 80" or "port 80: banner text"
		line = strings.TrimPrefix(line, "port ")
		portStr, banner, _ := strings.Cut(line, ":")
		portNum, err := strconv.Atoi(strings.TrimSpace(portStr))
		if err != nil {
			continue
		}
		ports = append(ports, shared.PortInfo{
			Port:   portNum,
			Banner: strings.TrimSpace(banner),
		})
	}

	return shared.HostResult{
		IP:       ip,
		Hardware: hardware,
		Ports:    ports,
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

	// Parse hosts from FinalHosts if available, otherwise FoundHosts
	hosts := m.FinalHosts
	if len(hosts) == 0 {
		hosts = m.FoundHosts
	}

	// Get ports that were scanned
	portsScanned := m.PortsScanned
	if portsScanned == nil {
		// Try to get from scanner
		switch s := m.NetworkScan.(type) {
		case *ip4.Scanner:
			portsScanned = s.Ports
		case *demo.Scanner:
			portsScanned = s.Ports
		}
	}

	var hostResults []history.HostResult
	now := time.Now()

	for _, hostStr := range hosts {
		h := parseHostString(hostStr)
		var ports []history.PortInfo
		for _, p := range h.Ports {
			ports = append(ports, history.PortInfo{
				Port:   p.Port,
				Banner: p.Banner,
			})
		}

		hostResults = append(hostResults, history.HostResult{
			IP:           h.IP,
			Hardware:     h.Hardware,
			MAC:          "", // MAC not currently tracked in display format
			Ports:        ports,
			LastScanned:  now,
			PortsScanned: portsScanned,
		})
	}

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
	// Parse hosts from FinalHosts if available, otherwise FoundHosts
	hosts := m.FinalHosts
	if len(hosts) == 0 {
		hosts = m.FoundHosts
	}

	portsScanned := m.PortsScanned
	if portsScanned == nil {
		portsScanned = getScannedPorts(m.NetworkScan)
	}

	now := time.Now()

	// If scanner returned no host text, still update the selected host with
	// the ports that were scanned and empty open-port results.
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

	// Parse the first (and should be only) host
	h := parseHostString(hosts[0])
	var ports []history.PortInfo
	for _, p := range h.Ports {
		ports = append(ports, history.PortInfo{
			Port:   p.Port,
			Banner: p.Banner,
		})
	}

	newHost := history.HostResult{
		IP:           h.IP,
		Hardware:     h.Hardware,
		MAC:          "",
		Ports:        ports,
		LastScanned:  now,
		PortsScanned: portsScanned,
	}

	// Update the specific host in the history file
	return history.UpdateHostInScan(m.RescanHistoryPath, h.IP, newHost)
}
