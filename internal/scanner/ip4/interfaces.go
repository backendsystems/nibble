package ip4

import (
	"net"
	"net/netip"

	"github.com/backendsystems/nibble/internal/scanner/ip4/wsl"
)

// GetInterfaces returns active non-loopback interfaces with at least one IPv4 address.
// When running inside WSL, Windows host interfaces are also included via interop.
func (s *Scanner) GetInterfaces() ([]net.Interface, map[string][]net.Addr, error) {
	sysIfaces, err := net.Interfaces()
	if err != nil {
		return nil, nil, err
	}

	ifaces := make([]net.Interface, 0, len(sysIfaces))
	addrsByIface := make(map[string][]net.Addr, len(sysIfaces))
	for _, iface := range sysIfaces {
		// Skip loopback interfaces.
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		// Skip interfaces that are down.
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		if hasIp4(addrs) && !wsl.IsWSLVirtualIface(iface.Name) {
			ifaces = append(ifaces, iface)
			addrsByIface[iface.Name] = addrs
		}
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

	return ifaces, addrsByIface, nil
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
	// addresses may be CIDR "192.168.1.10/24"
	prefix, err := netip.ParsePrefix(s)
	if err == nil {
		return prefix.Addr(), true
	}

	// or plain "192.168.1.10"
	ip, err := netip.ParseAddr(s)
	if err == nil {
		return ip, true
	}

	return netip.Addr{}, false
}
