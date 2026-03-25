package shared

import (
	"github.com/backendsystems/nibble/internal/ports/services"
)

// PortInfo holds a port number and its service banner
type PortInfo struct {
	Port   int    `json:"port"`
	Service string `json:"service,omitempty"`
	Source string `json:"source,omitempty"`
}

// HostResult holds all scan info for a single host
type HostResult struct {
	IP       string     `json:"ip"`
	Hardware string     `json:"hardware,omitempty"`
	Ports    []PortInfo `json:"ports,omitempty"`
}

// EnrichPorts fills in service names for ports that have no banner.
// This keeps the data layer consistent for both TUI and JSON output.
func EnrichPorts(h *HostResult) {
	for i := range h.Ports {
		if h.Ports[i].Service == "" {
			if info := services.Lookup(h.Ports[i].Port); info != nil {
				h.Ports[i].Service = info.Name + " - " + info.Description
				h.Ports[i].Source = "lookup"
			}
		}
	}
}
