package targetview

import (
	"net"
	"strings"
)

// getInterfaceIPs collects all IPv4 addresses from active network interfaces
func getInterfaceIPs(ifaces []net.Interface) []string {
	var ips []string

	for _, iface := range ifaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil || len(addrs) == 0 {
			continue
		}

		// Get IPv4 addresses from this interface
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if ipv4 := ipnet.IP.To4(); ipv4 != nil {
					ips = append(ips, ipv4.String())
				}
			}
		}
	}

	return ips
}

// getInterfaceInfos collects interface names and IPv4 addresses from active network interfaces
func getInterfaceInfos(ifaces []net.Interface) []InterfaceInfo {
	var infos []InterfaceInfo

	for _, iface := range ifaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil || len(addrs) == 0 {
			continue
		}

		// Get IPv4 addresses from this interface
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if ipv4 := ipnet.IP.To4(); ipv4 != nil {
					infos = append(infos, InterfaceInfo{
						Name: iface.Name,
						IP:   ipv4.String(),
					})
				}
			}
		}
	}

	return infos
}

// getDefaultIP tries to find an ethernet interface IP, otherwise returns empty string
func getDefaultIP(ifaces []net.Interface) string {
	if len(ifaces) == 0 {
		return ""
	}

	// Try to find an ethernet interface (eth, en, etc.)
	for _, iface := range ifaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		// Check if it's likely an ethernet interface
		name := strings.ToLower(iface.Name)
		if strings.HasPrefix(name, "eth") || strings.HasPrefix(name, "en") || strings.HasPrefix(name, "wlan") {
			addrs, err := iface.Addrs()
			if err != nil || len(addrs) == 0 {
				continue
			}

			// Get first IPv4 address
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok {
					if ipv4 := ipnet.IP.To4(); ipv4 != nil {
						return ipv4.String()
					}
				}
			}
		}
	}

	// Fallback to first non-loopback interface with an address
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil || len(addrs) == 0 {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if ipv4 := ipnet.IP.To4(); ipv4 != nil {
					return ipv4.String()
				}
			}
		}
	}

	return ""
}
