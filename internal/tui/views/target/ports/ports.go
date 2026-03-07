package targetports

import (
	"errors"
	"fmt"
	"net"

	"github.com/backendsystems/nibble/internal/ports"
)

// ScanConfig is the resolved output of BuildScanConfig.
type ScanConfig struct {
	TargetAddr string
	TotalHosts int
	Ports      []int
}

// BuildScanConfig validates IP/CIDR/port inputs and returns a complete scan configuration.
func BuildScanConfig(ipInput, cidrInput, portPack, customPorts string) (ScanConfig, error) {
	if ipInput == "" {
		return ScanConfig{}, errors.New("IP address required")
	}

	cidrStr := cidrInput
	if cidrStr == "" {
		cidrStr = "32"
	}
	cidrVal := 0
	_, err := fmt.Sscanf(cidrStr, "%d", &cidrVal)
	if err != nil || cidrVal < 16 || cidrVal > 32 {
		return ScanConfig{}, errors.New("CIDR must be 16-32")
	}

	cidrNotation := ipInput + "/" + cidrStr
	_, ipnet, err := net.ParseCIDR(cidrNotation)
	if err != nil {
		return ScanConfig{}, errors.New("invalid IP address")
	}

	var resolvedPorts []int
	switch portPack {
	case "all":
		resolvedPorts, err = ports.ParseList("1-65535")
	case "custom":
		if customPorts == "" {
			resolvedPorts = []int{}
		} else {
			resolvedPorts, err = ports.ParseList(customPorts)
		}
	default:
		resolvedPorts = ports.DefaultPorts()
	}
	if err != nil {
		return ScanConfig{}, err
	}

	return ScanConfig{
		TargetAddr: cidrNotation,
		TotalHosts: HostCount(ipnet),
		Ports:      resolvedPorts,
	}, nil
}

// HostCount returns the number of scannable hosts in an IP network.
func HostCount(ipnet *net.IPNet) int {
	ones, bits := ipnet.Mask.Size()
	hostBits := bits - ones

	if hostBits <= 0 {
		return 1
	}

	total := 1 << uint(hostBits)

	// /31 keeps both addresses as usable hosts (RFC 3021)
	if hostBits == 1 {
		return total
	}

	return total - 2
}
