package ip4

import (
	"net"

	"github.com/backendsystems/nibble/internal/ports"
	"github.com/backendsystems/nibble/internal/scanner/shared"
)

// Scanner performs real network scanning (TCP connect, ARP, banner grab)
type Scanner struct {
	Ports         []int
	dockerIfaces map[string]struct{} // display names of Docker network interfaces
}

// DockerIfaces returns the set of interface display names that correspond to
// Docker networks. Only populated after GetInterfaces has been called.
func (s *Scanner) DockerIfaces() map[string]struct{} {
	return s.dockerIfaces
}

// ScanNetwork scans a real subnet with controlled concurrency for smooth progress
func (s *Scanner) ScanNetwork(ifaceName, subnet string, progressChan chan<- shared.ProgressUpdate) {
	defer close(progressChan)

	_, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		return
	}

	totalHosts := shared.TotalScanHosts(ipnet)

	// Skip neighbor discovery for target scans (when no interface specified)
	var skipIPs map[string]struct{}
	var exhaustive bool
	if ifaceName != "" {
		skipIPs, exhaustive = s.neighborDiscovery(ifaceName, ipnet, totalHosts, progressChan)
	} else {
		skipIPs = make(map[string]struct{})
	}

	if exhaustive {
		// Docker socket gave us the complete container list — no sweep needed.
		// Emit a final SweepProgress so the progress bar reaches 100%.
		progressChan <- shared.SweepProgress{TotalHosts: totalHosts, Scanned: totalHosts}
		return
	}

	s.subnetSweep(ifaceName, ipnet, totalHosts, skipIPs, progressChan)
}

func (s *Scanner) ports() (out []int) {
	// Use configured ports if explicitly set (even if empty for host-only scan)
	if s.Ports != nil {
		return s.Ports
	}
	// Otherwise use defaults
	return ports.DefaultPorts()
}
