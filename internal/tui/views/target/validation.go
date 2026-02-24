package targetview

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/backendsystems/nibble/internal/ports"
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

// buildScanConfig extracts form values and builds a complete scan configuration
// Returns: targetAddr (CIDR notation), totalHosts, resolvedPorts, error
func buildScanConfig(ipInput, cidrInput, portPack, customPorts string) (string, int, []int, error) {
	// Validate IP
	if ipInput == "" {
		return "", 0, nil, errors.New("IP address required")
	}

	// Validate CIDR
	cidrStr := cidrInput
	if cidrStr == "" {
		cidrStr = "32" // Default to single host
	}
	cidrVal := 0
	_, err := fmt.Sscanf(cidrStr, "%d", &cidrVal)
	if err != nil || cidrVal < 16 || cidrVal > 32 {
		return "", 0, nil, errors.New("CIDR must be 16-32")
	}

	// Build CIDR notation and validate
	cidrNotation := ipInput + "/" + cidrStr
	_, ipnet, err := net.ParseCIDR(cidrNotation)
	if err != nil {
		return "", 0, nil, errors.New("invalid IP address")
	}

	// Calculate total hosts
	totalHosts := hostCount(ipnet)

	// Resolve ports based on mode
	var resolvedPorts []int
	switch portPack {
	case "all":
		resolvedPorts, err = ports.ParseList("1-65535")
	case "custom":
		if customPorts == "" {
			resolvedPorts = []int{} // Empty list = host-only scan
		} else {
			resolvedPorts, err = ports.ParseList(customPorts)
		}
	default: // "default" mode
		resolvedPorts = ports.DefaultPorts()
	}

	if err != nil {
		return "", 0, nil, err
	}

	return cidrNotation, totalHosts, resolvedPorts, nil
}

// hostCount calculates the number of hosts in an IP network
func hostCount(ipnet *net.IPNet) int {
	ones, bits := ipnet.Mask.Size()
	hostBits := bits - ones

	// /32 has one host
	if hostBits <= 0 {
		return 1
	}

	totalHosts := 1 << uint(hostBits)

	// /31 keeps both addresses as usable hosts (RFC 3021)
	if hostBits == 1 {
		return totalHosts // 2 hosts
	}

	// Larger subnets skip network and broadcast addresses
	return totalHosts - 2
}

// validateInputChar returns true if the character is valid for the given field
func validateInputChar(fieldKey string, ch byte) bool {
	switch fieldKey {
	case "ip":
		// IP field: only digits and dots
		return (ch >= '0' && ch <= '9') || ch == '.'
	case "cidr":
		// CIDR field: only digits
		return ch >= '0' && ch <= '9'
	case "custom_ports":
		// Custom ports field: block navigation keys w/s/k/j
		return ch != 'w' && ch != 's' && ch != 'k' && ch != 'j'
	default:
		return true
	}
}

// normalizeCustomPorts validates and normalizes custom port input
func normalizeCustomPorts(customPorts string) (string, error) {
	normalized, err := ports.NormalizeCustom(strings.TrimSpace(customPorts))
	if err != nil {
		return "", err
	}
	return normalized, nil
}
