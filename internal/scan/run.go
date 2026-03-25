package scan

import (
	"github.com/backendsystems/nibble/internal/scanner"
	"github.com/backendsystems/nibble/internal/scanner/config"
	"github.com/backendsystems/nibble/internal/scanner/shared"
)

// TargetResult holds scan results for a single CIDR target.
type TargetResult struct {
	CIDR  string              `json:"cidr"`
	Hosts []shared.HostResult `json:"hosts"`
}

// Run scans the given targets and returns results per target.
func Run(targets []string, customPorts []int, demoMode bool) []TargetResult {
	s := scanner.New(demoMode)
	if customPorts != nil {
		config.SetPorts(s, customPorts)
	}

	var results []TargetResult
	for _, cidr := range targets {
		hosts := scanTarget(s, cidr)
		results = append(results, TargetResult{
			CIDR:  cidr,
			Hosts: hosts,
		})
	}
	return results
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

	if hosts == nil {
		hosts = []shared.HostResult{}
	}
	return hosts
}
