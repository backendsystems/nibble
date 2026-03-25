package scan

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/backendsystems/nibble/internal/ports"
	"github.com/backendsystems/nibble/internal/scanner"
	"github.com/backendsystems/nibble/internal/scanner/config"
	"github.com/backendsystems/nibble/internal/scanner/shared"
)

type output struct {
	Meta    meta           `json:"meta"`
	Targets []targetResult `json:"targets"`
}

type meta struct {
	Ports      string    `json:"ports"`
	PortCount  int       `json:"port_count"`
	StartedAt  time.Time `json:"started_at"`
	DurationMs int64     `json:"duration_ms"`
}

type targetResult struct {
	CIDR  string              `json:"cidr"`
	Hosts []shared.HostResult `json:"hosts"`
}

// ErrNoHosts is returned when a scan completes successfully but finds no hosts.
var ErrNoHosts = fmt.Errorf("no hosts found")

// Run performs a headless scan of the given targets and writes JSON output.
// If outputPath is empty, writes to stdout. Otherwise writes to the specified file.
// Returns nil on success, ErrNoHosts if no hosts were found, or another error on failure.
func Run(targets []string, customPorts []int, portsRaw string, demoMode bool, outputPath string) error {
	s := scanner.New(demoMode)
	if customPorts != nil {
		config.SetPorts(s, customPorts)
	}

	scanPorts := customPorts
	if scanPorts == nil {
		scanPorts = ports.DefaultPorts()
	}

	if portsRaw == "" {
		portsRaw = ports.FormatDefault()
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

	duration := time.Since(start)

	out := output{
		Meta: meta{
			Ports:      portsRaw,
			PortCount:  len(scanPorts),
			StartedAt:  start.UTC(),
			DurationMs: duration.Milliseconds(),
		},
		Targets: results,
	}

	var w io.Writer = os.Stdout
	if outputPath != "" {
		f, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("creating output file: %w", err)
		}
		defer f.Close()
		w = f
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}

	totalHosts := 0
	for _, t := range results {
		totalHosts += len(t.Hosts)
	}
	if totalHosts == 0 {
		return ErrNoHosts
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
