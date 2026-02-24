package ip4

import (
	"net"

	"github.com/backendsystems/nibble/internal/scanner/shared"
	"github.com/backendsystems/nibble/internal/ports"
)

// Scanner performs real network scanning (TCP connect, ARP, banner grab)
type Scanner struct {
	Ports []int
}

// ScanNetwork scans a real subnet with controlled concurrency for smooth progress
func (s *Scanner) ScanNetwork(ifaceName, subnet string, progressChan chan<- shared.ProgressUpdate) {
	_, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		return
	}

	totalHosts := shared.TotalScanHosts(ipnet)

	// Skip neighbor discovery for target scans (when no interface specified)
	var skipIPs map[string]struct{}
	if ifaceName != "" {
		skipIPs = s.neighborDiscovery(ifaceName, ipnet, totalHosts, progressChan)
	} else {
		skipIPs = make(map[string]struct{})
	}

	s.subnetSweep(ifaceName, ipnet, totalHosts, skipIPs, progressChan)

	close(progressChan)
}

func (s *Scanner) ports() (out []int) {
	// Use configured ports if explicitly set (even if empty for host-only scan)
	if s.Ports != nil {
		return s.Ports
	}
	// Otherwise use defaults
	return ports.DefaultPorts()
}
