package parameters

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/backendsystems/nibble/internal/ports"
)

type Mode int

const (
	ModeTUI     Mode = iota
	ModeVersion Mode = iota
	ModeHeadless    Mode = iota
)

type Params struct {
	Mode     Mode
	Version  string
	DemoMode bool
	CIDR     string
	Ports    []int
}

func Parse(version string) Params {
	var (
		demoMode    bool
		showVersion bool
		cidr        string
		portsRaw    string
	)

	flag.BoolVar(&demoMode, "demo", false, "use demo interfaces")
	flag.BoolVar(&showVersion, "v", false, "")
	flag.BoolVar(&showVersion, "version", false, "print version and exit")
	flag.StringVar(&cidr, "i", "", "scan target IP/CIDR and output JSON (e.g. 192.168.0.0/24)")
	flag.StringVar(&portsRaw, "p", "", "custom ports to scan, used with -i (e.g. 22,80,8000-8100)")
	flag.Parse()

	if showVersion {
		return Params{Mode: ModeVersion, Version: version}
	}

	if cidr != "" {
		normalized, err := validateAndNormalizeCIDR(cidr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid -i value: %v\n", err)
			os.Exit(1)
		}

		p := Params{
			Mode:     ModeHeadless,
			DemoMode: demoMode,
			CIDR:     normalized,
		}

		if portsRaw != "" {
			parsed, err := ports.ParseList(portsRaw)
			if err != nil {
				fmt.Fprintf(os.Stderr, "invalid -p value: %v\n", err)
				os.Exit(1)
			}
			p.Ports = parsed
		}

		return p
	}

	if portsRaw != "" {
		fmt.Fprintln(os.Stderr, "-p can only be used with -i")
		os.Exit(1)
	}

	return Params{Mode: ModeTUI, DemoMode: demoMode}
}

// validateAndNormalizeCIDR validates the input is a valid IPv4 address or CIDR
// and returns it in CIDR notation (plain IPs become /32).
func validateAndNormalizeCIDR(s string) (string, error) {
	if strings.Contains(s, "/") {
		ip, ipnet, err := net.ParseCIDR(s)
		if err != nil {
			return "", fmt.Errorf("%q is not a valid CIDR", s)
		}
		if ip.To4() == nil || ipnet.IP.To4() == nil {
			return "", fmt.Errorf("%q is not an IPv4 CIDR", s)
		}
		return s, nil
	}

	// Plain IP — default to /32
	parsed := net.ParseIP(s)
	if parsed == nil || parsed.To4() == nil {
		return "", fmt.Errorf("%q is not a valid IPv4 address", s)
	}
	return s + "/32", nil
}
