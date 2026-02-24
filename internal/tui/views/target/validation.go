package targetview

import (
	"fmt"
	"net"

	"github.com/backendsystems/nibble/internal/scanner/shared"
)

// getIPsDescription returns a description showing the count of available interface IPs
func getIPsDescription(m *Model) string {
	count := len(m.InterfaceInfos)
	if count == 0 {
		return "No interfaces found"
	}
	if m.IPIndex < 0 || m.IPIndex >= count {
		m.IPIndex = 0
	}
	current := m.IPIndex + 1
	info := m.InterfaceInfos[m.IPIndex]
	return fmt.Sprintf("interfaces %d/%d ←/→ [%s]", current, count, info.Name)
}

// getHostCountDesc returns a description showing the number of hosts for the current CIDR
func getHostCountDesc(m *Model) string {
	if m.CIDRInput == "" {
		return ""
	}

	cidrStr := m.CIDRInput
	cidr := 0
	_, err := fmt.Sscanf(cidrStr, "%d", &cidr)
	if err != nil || cidr < 16 || cidr > 32 {
		return ""
	}

	// Use a dummy IP to calculate host count - any valid IP works since we only care about the CIDR
	fullCIDR := "0.0.0.0/" + cidrStr
	_, ipnet, err := net.ParseCIDR(fullCIDR)
	if err != nil {
		return ""
	}

	hostCount := shared.TotalScanHosts(ipnet)
	return fmt.Sprintf("targets: %d", hostCount)
}
