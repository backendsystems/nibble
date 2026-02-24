package targetview

import (
	"fmt"
	"net"
	"regexp"

	"github.com/backendsystems/nibble/internal/tui/views/common"
	"github.com/charmbracelet/huh"
)

// initializeForm creates the form for the target view model
func (m *Model) initializeForm() {
	// Collect all available interface IPs
	m.InterfaceIPs = getInterfaceIPs(m.Interfaces)

	// Default to ethernet interface IP if no IP provided
	if m.IPInput == "" {
		m.IPInput = getDefaultIP(m.Interfaces)
	}

	// Default CIDR if not provided
	if m.CIDRInput == "" {
		m.CIDRInput = "32"
	}

	// Default port pack if not provided
	if m.PortPack == "" {
		m.PortPack = "default"
	}

	// Build form with IP, CIDR, and port selection
	m.Form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("ip").
				Title("IP address").
				Placeholder("192.168.1.0").
				Description(getIPsDescription(m)).
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
					return getHostCountDesc(m)
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
	).WithTheme(common.FormTheme())
}
