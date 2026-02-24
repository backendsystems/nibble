package demo

import (
	"net"
	"time"

	"github.com/backendsystems/nibble/internal/ports"
	"github.com/backendsystems/nibble/internal/scanner/shared"
)

// Scanner simulates a scan with fake host data.
type Scanner struct {
	Ports []int
}

const (
	demoConcurrentPorts = 12000
	demoPortTimeout     = 70 * time.Millisecond
)

func (s *Scanner) ScanNetwork(ifaceName, subnet string, progressChan chan<- shared.ProgressUpdate) {
	_, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		close(progressChan)
		return
	}

	totalHosts := shared.TotalScanHosts(ipnet)
	selected := selectedPorts(s.Ports)
	selectedSet := make(map[int]struct{}, len(selected))
	for _, p := range selected {
		selectedSet[p] = struct{}{}
	}
	// nil means "use defaults"; explicit empty slice means host-only.
	hostOnly := s.Ports != nil && len(s.Ports) == 0
	hosts := allDemoHosts()
	neighborDelay, sweepDelay := demoDelaysForInterface(ifaceName)
	portDelay := demoPortDelay(len(selected), hostOnly)

	// Pick which demo hosts belong to this subnet.
	var subnetHosts []shared.HostResult
	for _, h := range hosts {
		ip := net.ParseIP(h.IP)
		if ip == nil || !ipnet.Contains(ip) {
			continue
		}
		resolved := shared.HostResult{
			IP:       h.IP,
			Hardware: shared.VendorFromMac(h.Hardware),
		}
		if !hostOnly {
			for _, p := range h.Ports {
				if _, ok := selectedSet[p.Port]; !ok {
					continue
				}
				resolved.Ports = append(resolved.Ports, shared.PortInfo{
					Port:   p.Port,
					Banner: p.Banner,
				})
			}
		}
		subnetHosts = append(subnetHosts, resolved)
	}

	// Skip neighbor discovery for target scans (when no interface specified)
	var neighbors, remaining []shared.HostResult
	if ifaceName != "" {
		// Emit nearby hosts first for interface scans
		neighborCount := 0
		if len(subnetHosts) > 0 {
			neighborCount = 1
			if len(subnetHosts) > 2 {
				neighborCount = 2
			}
		}

		neighbors = subnetHosts[:neighborCount]
		remaining = subnetHosts[neighborCount:]
		for i, h := range neighbors {
			time.Sleep(neighborDelay)
			time.Sleep(portDelay)
			progressChan <- shared.NeighborProgress{
				Host:       shared.FormatHost(h),
				TotalHosts: totalHosts,
				Seen:       i + 1,
				Total:      neighborCount,
			}
		}
		if neighborCount == 0 {
			progressChan <- shared.NeighborProgress{
				TotalHosts: totalHosts,
				Seen:       0,
				Total:      0,
			}
		}
	} else {
		// For target scans, skip neighbor phase and use all hosts
		remaining = subnetHosts
	}

	// Spread remaining hosts across the sweep.
	hostInterval := 0
	if len(remaining) > 0 {
		hostInterval = totalHosts / (len(remaining) + 1)
		if hostInterval == 0 {
			hostInterval = 1
		}
	}
	hostIdx := 0

	for i := 1; i <= totalHosts; i++ {
		time.Sleep(sweepDelay)

		host := ""
		if hostInterval > 0 && hostIdx < len(remaining) && i == hostInterval*(hostIdx+1) {
			time.Sleep(portDelay)
			host = shared.FormatHost(remaining[hostIdx])
			hostIdx++
		}

		progressChan <- shared.SweepProgress{
			Host:       host,
			TotalHosts: totalHosts,
			Scanned:    i,
		}
	}

	close(progressChan)
}

func demoPortDelay(selectedPorts int, hostOnly bool) time.Duration {
	if hostOnly || selectedPorts <= 0 {
		return 0
	}
	waves := (selectedPorts + demoConcurrentPorts - 1) / demoConcurrentPorts
	return time.Duration(waves) * demoPortTimeout
}

func allDemoHosts() []Host {
	hosts := make([]Host, 0, len(Hosts)+len(WiFiHosts))
	hosts = append(hosts, Hosts...)
	hosts = append(hosts, WiFiHosts...)
	return hosts
}

func demoDelaysForInterface(ifaceName string) (neighborDelay, sweepDelay time.Duration) {
	if ifaceName == "wlan0" {
		return 260 * time.Millisecond, 20 * time.Millisecond
	}
	return 180 * time.Millisecond, 10 * time.Millisecond
}

func selectedPorts(configured []int) []int {
	if configured != nil {
		return configured
	}
	return ports.DefaultPorts()
}
