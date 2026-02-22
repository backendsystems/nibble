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
	skipIPs := s.neighborDiscovery(ifaceName, ipnet, totalHosts, progressChan)
	s.subnetSweep(ifaceName, ipnet, totalHosts, skipIPs, progressChan)

	close(progressChan)
}

func (s *Scanner) ports() (out []int) {
	out = ports.DefaultPorts()
	if s.Ports != nil {
		out = s.Ports
	}
	return out
}
