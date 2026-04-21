package ip4

import (
	"net"
	"net/netip"

	"github.com/backendsystems/nibble/internal/scanner/ip4/docker"
	"github.com/backendsystems/nibble/internal/scanner/ip4/wsl"
)

// GetInterfaces returns active non-loopback interfaces with at least one IPv4 address.
// When running inside WSL, Windows host interfaces are also included via interop.
func (s *Scanner) GetInterfaces() ([]net.Interface, map[string][]net.Addr, error) {
	sysIfaces, err := net.Interfaces()
	if err != nil {
		return nil, nil, err
	}

	// Fetch Docker network info once — used for detection, clamping, and renaming.
	dockerNetworks := docker.Networks()

	ifaces := make([]net.Interface, 0, len(sysIfaces))
	addrsByIface := make(map[string][]net.Addr, len(sysIfaces))
	for _, iface := range sysIfaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		if !hasIp4(addrs) || wsl.IsWSLVirtualIface(iface.Name) {
			continue
		}

		if docker.IsBridgeIface(iface.Name, dockerNetworks) {
			if !docker.BridgeHasPeers(iface.Name) {
				continue
			}
			addrs = docker.ClampToCIDR(addrs, 24)
		}

		ifaces = append(ifaces, iface)
		addrsByIface[iface.Name] = addrs
	}

	// In WSL, also include Windows host interfaces via interop.
	if wsl.IsWSL() {
		for _, wiface := range wsl.Interfaces() {
			ifaces = append(ifaces, net.Interface{
				Index:        wiface.Index,
				Name:         wiface.Name,
				HardwareAddr: wiface.HWAddr,
				Flags:        net.FlagUp | net.FlagBroadcast | net.FlagMulticast,
			})
			addrsByIface[wiface.Name] = wiface.Addrs
		}
	}

	ifaces, addrsByIface = deduplicateSubnets(ifaces, addrsByIface)
	s.dockerIfaces = docker.ApplyNetworkNames(ifaces, addrsByIface, dockerNetworks)

	// Docker Desktop runs containers in a VM — no kernel bridge interfaces exist
	// on the host. Synthesize interface entries from the Docker API so the TUI
	// can show and scan each Docker network.
	if docker.IsDesktop(dockerNetworks) {
		for i, dn := range docker.DesktopNetworks() {
			displayName := dn.Name
			synth := net.Interface{
				Index: 10000 + i,
				Name:  displayName,
				Flags: net.FlagUp | net.FlagBroadcast | net.FlagMulticast,
			}
			addr := &net.IPNet{
				IP:   dn.Subnet.IP,
				Mask: dn.Subnet.Mask,
			}
			ifaces = append(ifaces, synth)
			addrsByIface[displayName] = []net.Addr{addr}
			s.dockerIfaces[displayName] = struct{}{}
		}
	}

	return ifaces, addrsByIface, nil
}

// deduplicateSubnets removes interfaces whose IPv4 address is already contained
// within another interface's subnet. This prevents showing redundant cards when
// multiple interfaces share the same network (e.g. eth0/24 and services1/32 on
// the same 192.168.65.0/24).
func deduplicateSubnets(ifaces []net.Interface, addrsByIface map[string][]net.Addr) ([]net.Interface, map[string][]net.Addr) {
	// Collect all subnets with their prefix length so we prefer the broader one.
	type ifaceSubnet struct {
		ones int
		net  *net.IPNet
	}
	var subnets []ifaceSubnet
	for _, iface := range ifaces {
		for _, addr := range addrsByIface[iface.Name] {
			ipnet, ok := addr.(*net.IPNet)
			if ok && ipnet.IP.To4() != nil {
				ones, _ := ipnet.Mask.Size()
				subnets = append(subnets, ifaceSubnet{ones: ones, net: ipnet})
			}
		}
	}

	out := make([]net.Interface, 0, len(ifaces))
	outAddrs := make(map[string][]net.Addr, len(ifaces))
	for _, iface := range ifaces {
		covered := false
		for _, addr := range addrsByIface[iface.Name] {
			ipnet, ok := addr.(*net.IPNet)
			if !ok || ipnet.IP.To4() == nil {
				continue
			}
			ones, _ := ipnet.Mask.Size()
			for _, s := range subnets {
				if s.net == ipnet {
					continue // skip self
				}
				if s.ones < ones && s.net.Contains(ipnet.IP) {
					covered = true
					break
				}
			}
			if covered {
				break
			}
		}
		if !covered {
			out = append(out, iface)
			outAddrs[iface.Name] = addrsByIface[iface.Name]
		}
	}
	return out, outAddrs
}

func hasIp4(addrs []net.Addr) bool {
	for _, addr := range addrs {
		ip, ok := parseAddr(addr.String())
		if ok && ip.Is4() {
			return true
		}
	}
	return false
}

func parseAddr(s string) (netip.Addr, bool) {
	prefix, err := netip.ParsePrefix(s)
	if err == nil {
		return prefix.Addr(), true
	}
	ip, err := netip.ParseAddr(s)
	if err == nil {
		return ip, true
	}
	return netip.Addr{}, false
}
