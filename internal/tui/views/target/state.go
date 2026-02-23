package targetview

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/backendsystems/nibble/internal/scanner/shared"
	"github.com/charmbracelet/huh"
)

type Model struct {
	Form         *huh.Form
	IPInput      string
	CIDRInput    string // e.g. "32", "24", "16"
	PortPack     string
	CustomPorts  string
	ErrorMsg     string
	NetworkScan  shared.Scanner
	InterfaceIPs []string        // Available interface IPs
	IPIndex      int             // Current index in InterfaceIPs
	Interfaces   []net.Interface // Store interfaces for later use
}

// NewModel creates a new target view model with a form bound to the model's fields
func NewModel(networkScan shared.Scanner, ipInput, cidrInput, portPack, customPorts string, ifaces []net.Interface) Model {
	// Collect all available interface IPs
	interfaceIPs := getInterfaceIPs(ifaces)

	// Default to ethernet interface IP if no IP provided
	if ipInput == "" {
		ipInput = getDefaultIP(ifaces)
	}

	m := Model{
		IPInput:      ipInput,
		CIDRInput:    cidrInput,
		PortPack:     portPack,
		CustomPorts:  customPorts,
		NetworkScan:  networkScan,
		InterfaceIPs: interfaceIPs,
		IPIndex:      0,
		Interfaces:   ifaces,
	}

	// Build form with IP, CIDR, and port selection
	m.Form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("ip").
				Title("IP address").
				Placeholder("192.168.1.0").
				Description(getIPsDescription(&m)).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("IP address required")
					}
					// Only allow digits and dots
					if !regexp.MustCompile(`^[0-9.]*$`).MatchString(s) {
						return fmt.Errorf("only numbers and dots allowed")
					}
					if net.ParseIP(s) == nil {
						return fmt.Errorf("invalid IP address")
					}
					return nil
				}).
				Value(&m.IPInput),

			huh.NewInput().
				Key("cidr").
				Title("CIDR (16-32)").
				Placeholder("32").
				DescriptionFunc(func() string {
					return getHostCountDesc(&m)
				}, &m.CIDRInput).
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
				Value(&m.CIDRInput),

			huh.NewSelect[string]().
				Key("port_mode").
				Title("Ports to scan").
				Options(
					huh.NewOption("Default", "default"),
					huh.NewOption("All ports (1-65535)", "all"),
					huh.NewOption("Custom ports", "custom"),
				).
				Value(&m.PortPack),
		),

		// Custom ports field - only shown when "custom" is selected
		huh.NewGroup(
			huh.NewInput().
				Key("custom_ports").
				Title("Custom ports").
				Description("Comma-separated, e.g. 22,80,443 or 8000-9000").
				Validate(func(s string) error {
					if m.PortPack == "custom" && s == "" {
						return fmt.Errorf("custom ports required when 'custom' mode selected")
					}
					return nil
				}).
				Value(&m.CustomPorts),
		).WithHideFunc(func() bool {
			return m.PortPack != "custom"
		}),
	).WithTheme(huh.ThemeCharm())

	return m
}

// CycleInterfaceIP cycles to the next or previous interface IP
// forward=true moves to next, forward=false moves to previous
func (m *Model) CycleInterfaceIP(forward bool) {
	if len(m.InterfaceIPs) == 0 {
		return
	}
	if forward {
		m.IPIndex = (m.IPIndex + 1) % len(m.InterfaceIPs)
	} else {
		m.IPIndex = (m.IPIndex - 1 + len(m.InterfaceIPs)) % len(m.InterfaceIPs)
	}
	m.IPInput = m.InterfaceIPs[m.IPIndex]
}

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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
	return fmt.Sprintf("hosts: %d", hostCount)
}

// getInterfaceIPs collects all IPv4 addresses from active network interfaces
func getInterfaceIPs(ifaces []net.Interface) []string {
	var ips []string

	for _, iface := range ifaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil || len(addrs) == 0 {
			continue
		}

		// Get IPv4 addresses from this interface
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if ipv4 := ipnet.IP.To4(); ipv4 != nil {
					ips = append(ips, ipv4.String())
				}
			}
		}
	}

	return ips
}

// getDefaultIP tries to find an ethernet interface IP, otherwise returns 192.168.0.1
func getDefaultIP(ifaces []net.Interface) string {
	if len(ifaces) == 0 {
		return "192.168.0.1"
	}

	// Try to find an ethernet interface (eth, en, etc.)
	for _, iface := range ifaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		// Check if it's likely an ethernet interface
		name := strings.ToLower(iface.Name)
		if strings.HasPrefix(name, "eth") || strings.HasPrefix(name, "en") || strings.HasPrefix(name, "wlan") {
			addrs, err := iface.Addrs()
			if err != nil || len(addrs) == 0 {
				continue
			}

			// Get first IPv4 address
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok {
					if ipv4 := ipnet.IP.To4(); ipv4 != nil {
						return ipv4.String()
					}
				}
			}
		}
	}

	// Fallback to first non-loopback interface with an address
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil || len(addrs) == 0 {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if ipv4 := ipnet.IP.To4(); ipv4 != nil {
					return ipv4.String()
				}
			}
		}
	}

	return "192.168.0.1"
}
