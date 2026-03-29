package ip4

import (
	"net"
	"runtime"
	"strings"

	"github.com/backendsystems/nibble/internal/scanner/ip4/docker"
	"github.com/backendsystems/nibble/internal/scanner/ip4/linux"
	"github.com/backendsystems/nibble/internal/scanner/ip4/macos"
	"github.com/backendsystems/nibble/internal/scanner/ip4/windows"
	"github.com/backendsystems/nibble/internal/scanner/ip4/wsl"
)

// lookupMacFromCache reads the OS ARP cache to find a MAC without needing root
// Linux reads /proc/net/arp, macOS reads routing table entries, Windows reads IP helper table entries
func lookupMacFromCache(ip string) string {
	if runtime.GOOS == "windows" {
		return windows.LookupMAC(ip)
	}
	if runtime.GOOS == "darwin" {
		return macos.LookupMAC(ip)
	}
	if wsl.IsWSL() {
		if mac := wsl.LookupMAC(ip); mac != "" {
			return mac
		}
	}
	return linux.LookupMAC(ip)
}

// NeighborEntry is a visible L2/L3 neighbor from the host ARP/neighbor table
type NeighborEntry struct {
	IP  string
	MAC string
}

// visibleNeighbors returns neighbors currently visible in the OS ARP
// table for the selected interface and subnet
// visibleNeighbors returns neighbors from the OS ARP table or Docker daemon.
// The second return value is true when the result is exhaustive — i.e. came
// from the Docker socket — meaning no subnet sweep is needed.
func (s *Scanner) visibleNeighbors(ifaceName string, subnet *net.IPNet) ([]NeighborEntry, bool) {
	// For Docker bridge interfaces, the daemon knows exactly which containers
	// are running — skip the ARP table entirely.
	if _, ok := s.dockerIfaces[ifaceName]; ok {
		if dockerNeighbors := docker.ContainerNeighbors(ifaceName, subnet); len(dockerNeighbors) > 0 {
			out := make([]NeighborEntry, len(dockerNeighbors))
			for i, n := range dockerNeighbors {
				out[i] = NeighborEntry{IP: n.IP, MAC: n.MAC}
			}
			return out, true
		}
	}

	var rows []NeighborEntry
	switch runtime.GOOS {
	case "windows":
		for _, row := range windows.Neighbors() {
			rows = append(rows, NeighborEntry{IP: row.IP, MAC: row.MAC})
		}
	case "darwin":
		for _, row := range macos.Neighbors(ifaceName) {
			rows = append(rows, NeighborEntry{IP: row.IP, MAC: row.MAC})
		}
	default:
		if strings.HasPrefix(ifaceName, "win:") {
			for _, row := range wsl.Neighbors(ifaceName) {
				rows = append(rows, NeighborEntry{IP: row.IP, MAC: row.MAC})
			}
		} else {
			for _, row := range linux.Neighbors(ifaceName) {
				rows = append(rows, NeighborEntry{IP: row.IP, MAC: row.MAC})
			}
		}
	}

	seen := make(map[string]struct{})
	var out []NeighborEntry
	for _, row := range rows {
		ip := net.ParseIP(row.IP)
		if ip == nil || ip.To4() == nil || !subnet.Contains(ip) {
			continue
		}
		if row.MAC == "" || row.MAC == "00:00:00:00:00:00" {
			continue
		}
		if strings.EqualFold(row.MAC, "ff:ff:ff:ff:ff:ff") {
			continue
		}
		if isSubnetBroadcastIpv4(ip, subnet) {
			continue
		}
		if _, ok := seen[row.IP]; ok {
			continue
		}
		seen[row.IP] = struct{}{}
		out = append(out, row)
	}
	return out, false
}

func isSubnetBroadcastIpv4(ip net.IP, subnet *net.IPNet) bool {
	ip4 := ip.To4()
	base := subnet.IP.To4()
	mask := subnet.Mask
	if ip4 == nil || base == nil || len(mask) != net.IPv4len {
		return false
	}
	broadcast := net.IPv4(
		base[0]|^mask[0],
		base[1]|^mask[1],
		base[2]|^mask[2],
		base[3]|^mask[3],
	)
	return ip4.Equal(broadcast)
}
