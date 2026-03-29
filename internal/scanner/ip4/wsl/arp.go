package wsl

import (
	"net"
	"strings"

	"github.com/backendsystems/nibble/internal/scanner/shared"
)

// Neighbor is an ARP entry from the Windows ARP table.
type Neighbor struct {
	IP  string
	MAC string
}

// Neighbors returns ARP entries from the Windows ARP table via arp.exe.
func Neighbors(winIfaceName string) []Neighbor {
	alias := strings.TrimPrefix(winIfaceName, "win:")

	out, err := runWinCmd("arp.exe", "-a")
	if err != nil {
		return nil
	}

	var rows []Neighbor
	inSection := alias == "" // if no alias filter, include all
	for line := range strings.SplitSeq(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Section headers look like: "Interface: 192.168.1.5 --- 0x4"
		if strings.HasPrefix(line, "Interface:") {
			inSection = alias == ""
			continue
		}
		if !inSection {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		ip := fields[0]
		if net.ParseIP(ip) == nil {
			continue
		}
		mac := shared.NormalizeMAC(fields[1])
		if mac == "" || mac == "00:00:00:00:00:00" || strings.EqualFold(mac, "ff:ff:ff:ff:ff:ff") {
			continue
		}
		rows = append(rows, Neighbor{IP: ip, MAC: mac})
	}
	return rows
}

// LookupMAC looks up a MAC address from the Windows ARP table.
func LookupMAC(ip string) string {
	for _, n := range Neighbors("") {
		if n.IP == ip {
			return n.MAC
		}
	}
	return ""
}
