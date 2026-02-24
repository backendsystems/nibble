package targetview

import (
	"net"
)

// BuildInterfaceInfos converts scanner/main interface data into target IP choices.
func BuildInterfaceInfos(ifaces []net.Interface, addrsByIface map[string][]net.Addr) []InterfaceInfo {
	infos := make([]InterfaceInfo, 0)

	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs := addrsByIface[iface.Name]
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if ipv4 := ipnet.IP.To4(); ipv4 != nil {
					infos = append(infos, InterfaceInfo{Name: iface.Name, IP: ipv4.String()})
				}
			}
		}
	}

	return infos
}

func buildInterfaceIPs(infos []InterfaceInfo) []string {
	ips := make([]string, 0, len(infos))
	for _, info := range infos {
		ips = append(ips, info.IP)
	}
	return ips
}
