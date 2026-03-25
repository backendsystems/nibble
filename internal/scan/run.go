package scan

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/backendsystems/nibble/internal/ports"
	"github.com/backendsystems/nibble/internal/scanner"
	"github.com/backendsystems/nibble/internal/scanner/config"
	"github.com/backendsystems/nibble/internal/scanner/shared"
)

type result struct {
	Targets      []targetResult `json:"targets"`
	PortsScanned []int          `json:"ports_scanned"`
	Duration     float64        `json:"duration_seconds"`
}

type targetResult struct {
	CIDR  string              `json:"cidr"`
	Hosts []shared.HostResult `json:"hosts"`
}

// Run performs a headless scan of the given targets and writes JSON to stdout.
func Run(targets []string, customPorts []int, demoMode bool) error {
	s := scanner.New(demoMode)
	if customPorts != nil {
		config.SetPorts(s, customPorts)
	}

	scanPorts := customPorts
	if scanPorts == nil {
		scanPorts = ports.DefaultPorts()
	}

	start := time.Now()

	var results []targetResult
	for _, cidr := range targets {
		hosts := scanTarget(s, cidr)
		results = append(results, targetResult{
			CIDR:  cidr,
			Hosts: hosts,
		})
	}

	duration := time.Since(start).Seconds()

	out := result{
		Targets:      results,
		PortsScanned: scanPorts,
		Duration:     duration,
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}
	return nil
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
