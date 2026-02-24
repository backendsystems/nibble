package targetview

import (
	"fmt"
	"net"
	"regexp"

	"github.com/backendsystems/nibble/internal/ports"
	"github.com/backendsystems/nibble/internal/tui/views/common"
	"github.com/charmbracelet/huh"
)

// initializeForm creates the form for the target view model
func (m *Model) initializeForm() {
	// Keep backward compatibility fields in sync from precomputed interface infos.
	if len(m.InterfaceIPs) == 0 && len(m.InterfaceInfos) > 0 {
		m.InterfaceIPs = buildInterfaceIPs(m.InterfaceInfos)
	}
	if len(m.InterfaceInfos) > 0 {
		if m.IPIndex < 0 || m.IPIndex >= len(m.InterfaceInfos) {
			m.IPIndex = 0
		}
		for i, info := range m.InterfaceInfos {
			if info.IP == m.IPInput {
				m.IPIndex = i
				break
			}
		}
		if m.IPInput == "" {
			m.IPInput = m.InterfaceInfos[m.IPIndex].IP
		}
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
				Description(common.CustomPortsDescription).
				Validate(func(s string) error {
					if m.PortPack == "custom" {
						if s == "" {
							return nil // Empty is valid for host-only scan
						}
						_, err := ports.NormalizeCustom(s)
						if err != nil {
							return err
						}
					}
					return nil
				}).
				Value(&m.CustomPorts),
		).WithHideFunc(func() bool {
			return m.PortPack != "custom"
		}),
	).WithTheme(common.FormTheme()).
		WithShowHelp(false) // Disable per-field help, use static help text instead
}
