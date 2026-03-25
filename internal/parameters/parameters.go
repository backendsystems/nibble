package parameters

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/backendsystems/nibble/internal/ports"
)

type Mode int

const (
	ModeTUI      Mode = iota
	ModeVersion  Mode = iota
	ModeHeadless Mode = iota
)

type Params struct {
	Mode     Mode
	Version  string
	DemoMode bool
	Targets  []string
	Ports    []int
	PortsRaw string
	Output   string
}

func Parse(version string) Params {
	var (
		demoMode    bool
		showVersion bool
		input    string
		portsRaw string
		output   string
	)

	flag.BoolVar(&demoMode, "demo", false, "use demo interfaces")
	flag.BoolVar(&showVersion, "v", false, "")
	flag.BoolVar(&showVersion, "version", false, "print version and exit")
	flag.StringVar(&input, "i", "", "scan targets: IP/CIDR comma-separated or file (e.g. 192.168.0.0/24,10.0.0.0/24 or targets.txt)")
	flag.StringVar(&portsRaw, "p", "", "custom ports to scan, used with -i (e.g. 22,80,8000-8100 or - for all)")
	flag.StringVar(&output, "o", "", "write JSON output to file (default: stdout)")
	flag.Parse()

	if flag.NArg() > 0 {
		fmt.Fprintf(os.Stderr, "unexpected argument: %s\n", flag.Arg(0))
		os.Exit(2)
	}

	if showVersion {
		return Params{Mode: ModeVersion, Version: version}
	}

	if input != "" {
		var targets []string
		for _, part := range strings.Split(input, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			resolved, err := resolveTargets(part)
			if err != nil {
				fmt.Fprintf(os.Stderr, "invalid -i value: %v\n", err)
				os.Exit(2)
			}
			targets = append(targets, resolved...)
		}
		if len(targets) == 0 {
			fmt.Fprintln(os.Stderr, "-i requires at least one target")
			os.Exit(2)
		}

		p := Params{
			Mode:     ModeHeadless,
			Version:  version,
			DemoMode: demoMode,
			Targets:  targets,
			Output:   output,
		}

		if portsRaw != "" {
			p.PortsRaw = strings.TrimSpace(portsRaw)
			if p.PortsRaw == "-" {
				p.Ports = ports.AllPorts()
				p.PortsRaw = "1-65535"
			} else {
				parsed, err := ports.ParseList(portsRaw)
				if err != nil {
					fmt.Fprintf(os.Stderr, "invalid -p value: %v\n", err)
					os.Exit(2)
				}
				p.Ports = parsed
			}
		}

		return p
	}

	if portsRaw != "" {
		fmt.Fprintln(os.Stderr, "-p can only be used with -i")
		os.Exit(2)
	}

	if output != "" {
		fmt.Fprintln(os.Stderr, "-o can only be used with -i")
		os.Exit(2)
	}

	return Params{Mode: ModeTUI, DemoMode: demoMode}
}

// resolveTargets returns a list of validated CIDR strings.
// If input is a file path, reads targets from it (one per line).
// Otherwise treats input as a single IP/CIDR.
func resolveTargets(input string) ([]string, error) {
	// Check if it's a file
	info, err := os.Stat(input)
	if err == nil && !info.IsDir() {
		return readTargetsFile(input)
	}

	// Treat as direct IP/CIDR
	normalized, err := validateAndNormalizeCIDR(input)
	if err != nil {
		return nil, err
	}
	return []string{normalized}, nil
}

func readTargetsFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer f.Close()

	var targets []string
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		normalized, err := validateAndNormalizeCIDR(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}
		targets = append(targets, normalized)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("no targets found in %s", path)
	}
	return targets, nil
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
