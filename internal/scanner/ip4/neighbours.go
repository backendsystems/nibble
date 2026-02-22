package ip4

import (
	"fmt"
	"net"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/backendsystems/nibble/internal/scanner/shared"
)

const portDialTimeout = 70 * time.Millisecond
const macosGlobalDialConcurrencyCap = 2 * 1024
const windowsGlobalDialConcurrencyCap = 6 * 1024
const unixGlobalDialConcurrencyCap = 12 * 1024

var dialLimiter = newDialLimiter()

type portResult struct {
	port   int
	banner string
}

func scanHost(ifaceName, ip string, ports []int) string {
	return scanHostMac(ifaceName, ip, "", ports)
}

func scanHostMac(ifaceName, ip, knownMAC string, ports []int) string {
	if len(ports) == 0 {
		// Host-only mode: ARP to check liveness (requires CAP_NET_RAW).
		// For neighbors knownMAC is already set so no ARP request is made.
		hardware := resolveHardware(net.ParseIP(ip), knownMAC)
		if knownMAC == "" && hardware == "" {
			return ""
		}
		return shared.FormatHost(shared.HostResult{IP: ip, Hardware: hardware})
	}

	results := scanOpenPorts(ip, ports)
	if len(results) == 0 {
		return ""
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].port < results[j].port
	})

	host := shared.HostResult{
		IP:       ip,
		Hardware: resolveHardware(net.ParseIP(ip), knownMAC),
		Ports:    make([]shared.PortInfo, 0, len(results)),
	}

	for _, result := range results {
		host.Ports = append(host.Ports, shared.PortInfo{Port: result.port, Banner: result.banner})
	}

	return shared.FormatHost(host)
}

func scanOpenPorts(ip string, ports []int) []portResult {
	var wg sync.WaitGroup
	var resultMu sync.Mutex
	results := make([]portResult, 0, len(ports))

	for _, port := range ports {
		wg.Add(1)
		go func(port int) {
			defer wg.Done()
			if dialLimiter != nil {
				// Acquire one slot in the global dial work pool
				dialLimiter <- struct{}{}
				defer func() {
					<-dialLimiter // release
				}()
			}

			conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), portDialTimeout)
			if err != nil {
				return
			}
			defer conn.Close()

			resultMu.Lock()
			results = append(results, portResult{port: port, banner: getServiceBanner(conn)})
			resultMu.Unlock()
		}(port)
	}

	wg.Wait()
	return results
}

// newDialLimiter returns a process wide semaphore that caps concurrent TCP dials.
// Windows uses a lower cap due to stricter socket/buffer limits.
func newDialLimiter() chan struct{} {
	switch runtime.GOOS {
	case "windows":
		return make(chan struct{}, windowsGlobalDialConcurrencyCap)
	case "darwin":
		return make(chan struct{}, macosGlobalDialConcurrencyCap)
	case "linux":
		return make(chan struct{}, unixGlobalDialConcurrencyCap)
	default:
		return nil
	}
}

func resolveHardware(targetIP net.IP, knownMAC string) string {
	if knownMAC != "" {
		return shared.VendorFromMac(knownMAC)
	}
	if targetIP == nil {
		return ""
	}

	mac := lookupMacFromCache(targetIP.String())
	if mac == "" {
		return ""
	}
	return shared.VendorFromMac(mac)
}
