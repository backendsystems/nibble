package docker

import (
	"encoding/json"
	"net"
	"os"
	"strings"
)

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Port is a published port binding from a Docker container.
type Port struct {
	PrivatePort int
	PublicPort  int
	Type        string
}

// Neighbor is a container visible on a Docker bridge network.
type Neighbor struct {
	IP    string
	MAC   string
	Name  string
	Image string
	Ports []Port
}

// DesktopNetwork is a Docker network returned by DesktopNetworks, carrying
// its subnet so a synthetic interface can be constructed for Docker Desktop.
type DesktopNetwork struct {
	Name   string
	Subnet *net.IPNet
}

// Networks queries the Docker daemon and returns a map of kernel bridge interface
// name (e.g. "br-09ff0748a767") to human-readable Docker network name
// (e.g. "myproject_default"). Returns nil if the socket is unavailable.
func Networks() map[string]string {
	resp, err := client().Get("http://localhost/networks")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var networks []struct {
		Name string `json:"Name"`
		ID   string `json:"Id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&networks); err != nil {
		return nil
	}

	result := make(map[string]string, len(networks))
	for _, n := range networks {
		if len(n.ID) >= 12 {
			result["br-"+n.ID[:12]] = n.Name
		}
	}
	return result
}

// DesktopNetworks queries the Docker daemon and returns bridge networks with
// their subnets. Used on Docker Desktop (Linux/Mac) where the bridge interfaces
// don't exist as kernel interfaces on the host. Returns nil if unavailable.
func DesktopNetworks() []DesktopNetwork {
	resp, err := client().Get("http://localhost/networks")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var raw []struct {
		Name   string `json:"Name"`
		Driver string `json:"Driver"`
		IPAM   struct {
			Config []struct {
				Subnet string `json:"Subnet"`
			} `json:"Config"`
		} `json:"IPAM"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil
	}

	var out []DesktopNetwork
	for _, n := range raw {
		if n.Driver != "bridge" || n.Name == "bridge" {
			continue
		}
		for _, cfg := range n.IPAM.Config {
			if cfg.Subnet == "" {
				continue
			}
			_, ipnet, err := net.ParseCIDR(cfg.Subnet)
			if err != nil || ipnet.IP.To4() == nil {
				continue
			}
			out = append(out, DesktopNetwork{Name: n.Name, Subnet: ipnet})
			break
		}
	}
	return out
}

// IsDesktop returns true when Docker is reachable via socket but no kernel
// bridge interfaces exist (i.e. Docker Desktop running in a VM).
func IsDesktop(networks map[string]string) bool {
	if len(networks) == 0 {
		return false
	}
	return isDesktop(networks)
}

// IsBridgeIface returns true if the kernel interface name belongs to a Docker
// network, using the map returned by Networks(). Falls back to name-prefix
// heuristic (docker0, br-*) when the socket was unavailable and networks is nil.
func IsBridgeIface(kernelName string, networks map[string]string) bool {
	if networks != nil {
		_, ok := networks[kernelName]
		return ok
	}
	return kernelName == "docker0" || strings.HasPrefix(kernelName, "br-")
}

// BridgeHasPeers returns true if the bridge has at least one veth peer attached,
// meaning at least one container is connected. Uses sysfs rather than ARP cache
// so idle-but-running containers are not incorrectly filtered out.
func BridgeHasPeers(name string) bool {
	entries, err := os.ReadDir("/sys/class/net/" + name + "/brif")
	return err == nil && len(entries) > 0
}

// ApplyNetworkNames renames any Docker bridge entries in ifaces and addrsByIface
// to their human-readable Docker network names. Returns the set of display names
// that correspond to Docker networks, for use in scan-time lookups.
func ApplyNetworkNames(ifaces []net.Interface, addrsByIface map[string][]net.Addr, networks map[string]string) map[string]struct{} {
	dockerDisplayNames := make(map[string]struct{}, len(networks))
	for i, iface := range ifaces {
		dockerName, ok := networks[iface.Name]
		if !ok {
			// No rename — track under the kernel name if it's a bridge.
			if strings.HasPrefix(iface.Name, "br-") || iface.Name == "docker0" {
				dockerDisplayNames[iface.Name] = struct{}{}
			}
			continue
		}
		addrs := addrsByIface[iface.Name]
		delete(addrsByIface, iface.Name)
		addrsByIface[dockerName] = addrs
		ifaces[i].Name = dockerName
		dockerDisplayNames[dockerName] = struct{}{}
	}
	return dockerDisplayNames
}

// ContainerNeighbors queries the Docker daemon and returns the IP, MAC, and
// published ports of every running container attached to networkName that falls
// within subnet. Returns nil if the socket is unavailable.
func ContainerNeighbors(networkName string, subnet *net.IPNet) []Neighbor {
	resp, err := client().Get("http://localhost/containers/json")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var containers []struct {
		Names []string `json:"Names"`
		Image string   `json:"Image"`
		Ports []struct {
			PrivatePort int    `json:"PrivatePort"`
			PublicPort  int    `json:"PublicPort"`
			Type        string `json:"Type"`
		} `json:"Ports"`
		NetworkSettings struct {
			Networks map[string]struct {
				IPAddress  string `json:"IPAddress"`
				MacAddress string `json:"MacAddress"`
			} `json:"Networks"`
		} `json:"NetworkSettings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		return nil
	}

	var out []Neighbor
	for _, c := range containers {
		for netName, ns := range c.NetworkSettings.Networks {
			if netName != networkName {
				continue
			}
			ip := net.ParseIP(ns.IPAddress)
			if ip == nil || !subnet.Contains(ip) {
				continue
			}
			cName := ""
			if len(c.Names) > 0 {
				cName = strings.TrimPrefix(c.Names[0], "/")
			}
			n := Neighbor{IP: ns.IPAddress, MAC: ns.MacAddress, Name: cName, Image: c.Image}
			seen := make(map[int]struct{})
			for _, p := range c.Ports {
				if p.Type != "tcp" || p.PublicPort == 0 {
					continue
				}
				if _, ok := seen[p.PublicPort]; ok {
					continue
				}
				seen[p.PublicPort] = struct{}{}
				n.Ports = append(n.Ports, Port{
					PrivatePort: p.PrivatePort,
					PublicPort:  p.PublicPort,
					Type:        p.Type,
				})
			}
			out = append(out, n)
		}
	}
	return out
}

// ClampToCIDR returns addrs with any IPv4 prefix longer than bits clamped to bits.
func ClampToCIDR(addrs []net.Addr, bits int) []net.Addr {
	out := make([]net.Addr, 0, len(addrs))
	for _, addr := range addrs {
		ipnet, ok := addr.(*net.IPNet)
		if !ok {
			out = append(out, addr)
			continue
		}
		ones, total := ipnet.Mask.Size()
		if total == 32 && ones > bits {
			out = append(out, &net.IPNet{
				IP:   ipnet.IP,
				Mask: net.CIDRMask(bits, 32),
			})
		} else {
			out = append(out, addr)
		}
	}
	return out
}
