package scan

import (
	"github.com/backendsystems/nibble/internal/scanner"
	"github.com/backendsystems/nibble/internal/scanner/config"
	"github.com/backendsystems/nibble/internal/scanner/shared"
)

// Run scans the given targets and returns a flat list of discovered hosts.
func Run(targets []string, customPorts []int, demoMode bool) []shared.HostResult {
	s := scanner.New(demoMode)
	if customPorts != nil {
		config.SetPorts(s, customPorts)
	}

	var hosts []shared.HostResult
	seen := make(map[string]struct{})

	for _, cidr := range targets {
		for _, h := range scanTarget(s, cidr) {
			if _, dup := seen[h.IP]; dup {
				continue
			}
			seen[h.IP] = struct{}{}
			hosts = append(hosts, h)
		}
	}

	if hosts == nil {
		hosts = []shared.HostResult{}
	}
	return hosts
}

func scanTarget(s shared.Scanner, cidr string) []shared.HostResult {
	progressChan := make(chan shared.ProgressUpdate, 256)

	go s.ScanNetwork("", cidr, progressChan)

	var hosts []shared.HostResult
	seen := make(map[string]struct{})

	for update := range progressChan {
		var host *shared.HostResult
		switch u := update.(type) {
		case shared.NeighborProgress:
			host = u.Host
		case shared.SweepProgress:
			host = u.Host
		}
		if host == nil {
			continue
		}
		if _, dup := seen[host.IP]; dup {
			continue
		}
		seen[host.IP] = struct{}{}
		hosts = append(hosts, *host)
	}

	return hosts
}
