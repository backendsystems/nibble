package docker

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

const socketPath = "/var/run/docker.sock"

// Neighbor is a container visible on a Docker bridge network.
type Neighbor struct {
	IP  string
	MAC string
}

func client() *http.Client {
	return &http.Client{
		Timeout: 2 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, "unix", socketPath)
			},
		},
	}
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

// ContainerNeighbors queries the Docker daemon and returns the IP and MAC of
// every running container attached to networkName that falls within subnet.
// Returns nil if the socket is unavailable.
func ContainerNeighbors(networkName string, subnet *net.IPNet) []Neighbor {
	resp, err := client().Get("http://localhost/containers/json")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var containers []struct {
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
		for name, ns := range c.NetworkSettings.Networks {
			if name != networkName {
				continue
			}
			ip := net.ParseIP(ns.IPAddress)
			if ip == nil || !subnet.Contains(ip) {
				continue
			}
			out = append(out, Neighbor{IP: ns.IPAddress, MAC: ns.MacAddress})
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
