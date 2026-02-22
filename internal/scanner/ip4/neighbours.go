package ip4

import (
	"fmt"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/backendsystems/nibble/internal/scanner/shared"
)

const portDialTimeout = 70 * time.Millisecond

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
	results := make([]portResult, 0, 16)

	portCh := make(chan int, len(ports))
	for _, p := range ports {
		portCh <- p
	}
	close(portCh)

	n := min(maxWorkers, len(ports))

	wg.Add(n)
	for range n {
		go func() {
			defer wg.Done()
			for port := range portCh {
				dialAndRecord(ip, port, &resultMu, &results)
			}
		}()
	}

	wg.Wait()
	return results
}

func dialAndRecord(ip string, port int, mu *sync.Mutex, results *[]portResult) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), portDialTimeout)
	if err != nil {
		return
	}
	defer conn.Close()
	banner := getServiceBanner(conn)
	mu.Lock()
	*results = append(*results, portResult{port: port, banner: banner})
	mu.Unlock()
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
