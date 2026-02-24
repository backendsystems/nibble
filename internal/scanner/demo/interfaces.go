// Package demo provides fake network data for demo recordings.
package demo

import "net"

// GetInterfaces returns fake interfaces used by demo mode.
func (s *Scanner) GetInterfaces() ([]net.Interface, map[string][]net.Addr, error) {
	specs := []struct {
		name string
		cidr string
	}{
		{name: "eth0", cidr: "192.168.1.100/24"},
		{name: "wlan0", cidr: "10.0.0.50/24"},
		{name: "docker0", cidr: "172.17.0.1/16"},
		{name: "wg0", cidr: "10.8.0.2/24"},
	}

	backing := backingInterfaces()

	ifaces := make([]net.Interface, 0, len(specs))
	addrsByIface := make(map[string][]net.Addr, len(specs))
	for i, s := range specs {
		iface, addrs, err := newInterface(s.name, s.cidr, backing[i%len(backing)])
		if err != nil {
			return nil, nil, err
		}
		ifaces = append(ifaces, iface)
		addrsByIface[iface.Name] = addrs
	}

	return ifaces, addrsByIface, nil
}

func backingInterfaces() []net.Interface {
	sys, err := net.Interfaces()
	if err != nil || len(sys) == 0 {
		return []net.Interface{{Index: 1, Flags: net.FlagUp}}
	}

	out := make([]net.Interface, 0, len(sys))
	for _, iface := range sys {
		if iface.Index <= 0 {
			continue
		}
		addrs, addrErr := iface.Addrs()
		if addrErr != nil {
			continue
		}
		if hasIPv4(addrs) {
			out = append(out, iface)
		}
	}
	if len(out) == 0 {
		return []net.Interface{{Index: 1, Flags: net.FlagUp}}
	}
	return out
}

func hasIPv4(addrs []net.Addr) bool {
	for _, addr := range addrs {
		ipnet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		if ipnet.IP.To4() != nil {
			return true
		}
	}
	return false
}

func newInterface(name, cidr string, backing net.Interface) (net.Interface, []net.Addr, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return net.Interface{}, nil, err
	}
	ipnet.IP = ip

	return net.Interface{
		Name:  name,
		Index: backing.Index,
		Flags: net.FlagUp,
	}, []net.Addr{ipnet}, nil
}
