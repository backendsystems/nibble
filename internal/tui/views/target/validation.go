package targetview

import (
	"fmt"
	"net"

	"github.com/backendsystems/nibble/internal/scanner/shared"
)

// getIPsDescription returns a description showing the count of available interface IPs
func getIPsDescription(m *Model) string {
	count := len(m.InterfaceIPs)
	if count == 0 {
		return "No interfaces found"
	}
	if count == 1 {
		return "1 interface available (↑↓ to cycle)"
	}
	return fmt.Sprintf("%d interfaces available (↑↓ to cycle)", count)
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
