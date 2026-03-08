package targetform

import (
	"fmt"
	"net"
	"regexp"

	"github.com/backendsystems/nibble/internal/scanner/shared"
	"github.com/backendsystems/nibble/internal/tui/views/common"
	"github.com/charmbracelet/huh"
)

// InterfaceInfo holds a network interface name and its IPv4 address.
type InterfaceInfo struct {
	Name string
	IP   string
}

// Build constructs the target form from the given field values.
// ipInput, cidrInput, and portPack are bound by pointer so the form updates them on submit.
func Build(ipInput, cidrInput, portPack *string, ipIndex int, ifaces []InterfaceInfo) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("ip").
				Title("IP address").
				Placeholder("192.168.1.0").
				CharLimit(15).
				Description(ipsDescription(ipIndex, ifaces)).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("IP address required")
					}
					if !regexp.MustCompile(`^[0-9.]*$`).MatchString(s) {
						return fmt.Errorf("only numbers and dots allowed")
					}
					if net.ParseIP(s) == nil {
						return fmt.Errorf("invalid IP address")
					}
					return nil
				}).
				Value(ipInput),

			huh.NewInput().
				Key("cidr").
				Title("CIDR (16-32)").
				Placeholder("32").
				CharLimit(2).
				DescriptionFunc(func() string {
					return hostCountDesc(*cidrInput)
				}, cidrInput).
				Validate(func(s string) error {
					if s == "" {
						return nil // optional, defaults to 32
					}
					cidr := 0
					_, err := fmt.Sscanf(s, "%d", &cidr)
					if err != nil || cidr < 16 || cidr > 32 {
						return fmt.Errorf("CIDR must be 16-32")
					}
					return nil
				}).
				Value(cidrInput),

			huh.NewSelect[string]().
				Key("port_mode").
				Title("Ports to scan").
				Options(
					huh.NewOption("Default", "default"),
					huh.NewOption("All ports (1-65535)", "all"),
					huh.NewOption("Custom ports", "custom"),
				).
				Value(portPack),
		),
	).WithTheme(common.FormTheme()).
		WithShowHelp(false)
}

func ipsDescription(ipIndex int, ifaces []InterfaceInfo) string {
	count := len(ifaces)
	if count == 0 {
		return "No interfaces found"
	}
	if ipIndex < 0 || ipIndex >= count {
		ipIndex = 0
	}
	return fmt.Sprintf("interfaces %d/%d ←/→ [%s]", ipIndex+1, count, ifaces[ipIndex].Name)
}

func hostCountDesc(cidrStr string) string {
	isDefault := false
	if cidrStr == "" {
		cidrStr = "32"
		isDefault = true
	}

	cidr := 0
	_, err := fmt.Sscanf(cidrStr, "%d", &cidr)
	if err != nil || cidr < 16 || cidr > 32 {
		return "targets: -"
	}

	_, ipnet, err := net.ParseCIDR("0.0.0.0/" + cidrStr)
	if err != nil {
		return "targets: -"
	}

	hosts := shared.TotalScanHosts(ipnet)
	if isDefault {
		return fmt.Sprintf("targets: %d (default /32)", hosts)
	}
	return fmt.Sprintf("targets: %d", hosts)
}
