package ip4

import (
	"net"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/backendsystems/nibble/internal/scanner/shared"
)

var maxWorkers = newMaxWorkers()

const (
	maxConcurrentHostScans = 24
)

func newMaxWorkers() int {
	switch runtime.GOOS {
	case "windows":
		return 6 * 1024
	case "darwin":
		return 2 * 1024
	default:
		return 12 * 1024
	}
}

// neighborDiscovery emits hosts already visible in neighbor tables.
// Returns the set of IPs to skip in the sweep, and whether discovery was
// exhaustive (true when the Docker socket was used and no sweep is needed).
func (s *Scanner) neighborDiscovery(ifaceName string, subnet *net.IPNet, totalHosts int, progressChan chan<- shared.ProgressUpdate) (map[string]struct{}, bool) {
	neighbors, exhaustive := s.visibleNeighbors(ifaceName, subnet)
	skipIPs := buildSkipMap(neighbors)
	if len(neighbors) == 0 {
		emitNeighborProgress(progressChan, shared.NeighborProgress{TotalHosts: totalHosts})
		return skipIPs, exhaustive
	}

	ports := s.ports()
	workerCount := min(len(neighbors), max(1, maxWorkers/max(1, len(ports))))

	jobs := make(chan NeighborEntry)
	var wg sync.WaitGroup
	var seenCount atomic.Int64

	for range workerCount {
		wg.Go(func() {
			for neighbor := range jobs {
				processNeighborJob(ifaceName, neighbor, ports, totalHosts, len(neighbors), &seenCount, progressChan)
			}
		})
	}

	for _, neighbor := range neighbors {
		jobs <- neighbor
	}
	close(jobs)

	wg.Wait()
	return skipIPs, exhaustive
}

// subnetSweep scans the subnet and skips hosts found in neighbor discovery
func (s *Scanner) subnetSweep(ifaceName string, subnet *net.IPNet, totalHosts int, skipIPs map[string]struct{}, progressChan chan<- shared.ProgressUpdate) {
	ports := s.ports()
	workerCount := min(maxConcurrentHostScans, max(1, maxWorkers/max(1, len(ports))))
	jobs := make(chan string, workerCount)
	var wg sync.WaitGroup
	var scanned atomic.Int64

	for range workerCount {
		wg.Go(func() {
			for currentIP := range jobs {
				processSweepJob(ifaceName, currentIP, ports, skipIPs, totalHosts, &scanned, progressChan)
			}
		})
	}

	for ip := subnet.IP.Mask(subnet.Mask); subnet.Contains(ip); incrementIP(ip) {
		if skipIp4(ip, subnet) {
			continue
		}

		jobs <- ip.String()
	}
	close(jobs)

	wg.Wait()
}

func processNeighborJob(ifaceName string, neighbor NeighborEntry, ports []int, totalHosts, totalNeighbors int, seenCount *atomic.Int64, progressChan chan<- shared.ProgressUpdate) {
	host := scanHostMac(neighbor.IP, neighbor.MAC, ports)
	if host == nil {
		host = neighborToHost(neighbor)
	}

	currentSeen := int(seenCount.Add(1))

	progress := shared.NeighborProgress{
		Host:       host,
		TotalHosts: totalHosts,
		Seen:       currentSeen,
		Total:      totalNeighbors,
	}

	// Always send blocking if we have host info - must not drop discoveries
	// Otherwise use non-blocking to avoid blocking workers when channel is full
	if progress.Host != nil {
		progressChan <- progress
	} else {
		emitNeighborProgress(progressChan, progress)
	}
}

func processSweepJob(ifaceName, currentIP string, ports []int, skipIPs map[string]struct{}, totalHosts int, scanned *atomic.Int64, progressChan chan<- shared.ProgressUpdate) {
	var host *shared.HostResult
	if len(ports) > 0 {
		if _, alreadyFound := skipIPs[currentIP]; !alreadyFound {
			host = scanHost(currentIP, ports)
		}
	}

	currentScanned := int(scanned.Add(1))

	progress := shared.SweepProgress{
		Host:       host,
		TotalHosts: totalHosts,
		Scanned:    currentScanned,
	}

	// Always send blocking if we found a host - must not drop discoveries
	// Otherwise use non-blocking to avoid blocking workers when channel is full
	if host != nil {
		progressChan <- progress
	} else {
		select {
		case progressChan <- progress:
		default:
		}
	}
}

func buildSkipMap(neighbors []NeighborEntry) map[string]struct{} {
	skipIPs := make(map[string]struct{}, len(neighbors))
	for _, neighbor := range neighbors {
		skipIPs[neighbor.IP] = struct{}{}
	}
	return skipIPs
}

func neighborToHost(neighbor NeighborEntry) *shared.HostResult {
	return &shared.HostResult{
		IP:       neighbor.IP,
		MAC:      neighbor.MAC,
		Hardware: shared.VendorFromMac(neighbor.MAC),
	}
}

func emitNeighborProgress(progressChan chan<- shared.ProgressUpdate, progress shared.NeighborProgress) {
	select {
	case progressChan <- progress:
	default:
	}
}

func incrementIP(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] > 0 {
			return
		}
	}
}

func skipIp4(ip net.IP, subnet *net.IPNet) bool {
	ip4 := ip.To4()
	base := subnet.IP.To4()
	mask := subnet.Mask
	if ip4 == nil || base == nil || len(mask) != net.IPv4len {
		return false
	}

	// For /31 and /32, don't skip any addresses - all are usable hosts
	ones, _ := subnet.Mask.Size()
	if ones >= 31 {
		return false
	}

	// loop
	if ip4.Equal(base.Mask(mask)) {
		return true
	}

	broadcast := net.IPv4(
		base[0]|^mask[0],
		base[1]|^mask[1],
		base[2]|^mask[2],
		base[3]|^mask[3],
	)
	// broadcast
	return ip4.Equal(broadcast)
}
