package parameters

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/backendsystems/nibble/internal/ports"
	"github.com/backendsystems/nibble/internal/scan"
	"github.com/backendsystems/nibble/internal/scanner/shared"
)

type headlessOutput struct {
	Meta  headlessMeta        `json:"meta"`
	Hosts []shared.HostResult `json:"hosts"`
}

type headlessMeta struct {
	Scanner    string   `json:"scanner"`
	Version    string   `json:"version"`
	Targets    []string `json:"targets"`
	Ports      string   `json:"ports"`
	PortCount  int      `json:"port_count"`
	StartedAt  string   `json:"started_at"`
	DurationMs int64    `json:"duration_ms"`
}

// RunHeadless executes a headless scan and writes JSON output.
func RunHeadless(p Params) {
	portsRaw := p.PortsRaw
	if portsRaw == "" {
		portsRaw = ports.FormatDefault()
	}

	portCount := len(p.Ports)
	if p.Ports == nil {
		portCount = len(ports.DefaultPorts())
	}

	start := time.Now()
	hosts := scan.Run(p.Targets, p.Ports, p.DemoMode)
	duration := time.Since(start)

	out := headlessOutput{
		Meta: headlessMeta{
			Scanner:    "nibble",
			Version:    p.Version,
			Targets:    p.Targets,
			Ports:      portsRaw,
			PortCount:  portCount,
			StartedAt:  start.UTC().Format(time.RFC3339),
			DurationMs: duration.Milliseconds(),
		},
		Hosts: hosts,
	}

	var w io.Writer = os.Stdout
	if p.Output != "" {
		f, err := os.Create(p.Output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "creating output file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		w = f
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "encoding JSON: %v\n", err)
		os.Exit(1)
	}

	if len(hosts) == 0 {
		os.Exit(2)
	}
}
