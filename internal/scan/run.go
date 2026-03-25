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
	CIDR         string            `json:"cidr"`
	PortsScanned []int             `json:"ports_scanned"`
	Duration     float64           `json:"duration_seconds"`
	Hosts        []shared.HostResult `json:"hosts"`
}

// Run performs a headless scan of the given CIDR and writes JSON to stdout.
func Run(cidr string, customPorts []int, demoMode bool) error {
	s := scanner.New(demoMode)
	if customPorts != nil {
		config.SetPorts(s, customPorts)
	}

	scanPorts := customPorts
	if scanPorts == nil {
		scanPorts = ports.DefaultPorts()
	}

	progressChan := make(chan shared.ProgressUpdate, 256)

	start := time.Now()

	go s.ScanNetwork("", cidr, progressChan)

	var hosts []shared.HostResult
	seen := make(map[string]struct{})

	for update := range progressChan {
		var hostStr string
		switch u := update.(type) {
		case shared.NeighborProgress:
			hostStr = u.Host
		case shared.SweepProgress:
			hostStr = u.Host
		}
		if hostStr == "" {
			continue
		}
		h := shared.ParseHost(hostStr)
		if h.IP == "" {
			continue
		}
		if _, dup := seen[h.IP]; dup {
			continue
		}
		seen[h.IP] = struct{}{}
		hosts = append(hosts, h)
	}

	duration := time.Since(start).Seconds()

	out := result{
		CIDR:         cidr,
		PortsScanned: scanPorts,
		Duration:     duration,
		Hosts:        hosts,
	}
	if out.Hosts == nil {
		out.Hosts = []shared.HostResult{}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}
	return nil
}
