package shared

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/backendsystems/nibble/internal/ports/services"
	"github.com/backendsystems/nibble/internal/tui/views/common"
)

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// PortInfo holds a port number and its service banner
type PortInfo struct {
	Port   int    `json:"port"`
	Banner string `json:"banner,omitempty"`
}

// HostResult holds all scan info for a single host
type HostResult struct {
	IP       string     `json:"ip"`
	Hardware string     `json:"hardware,omitempty"`
	Ports    []PortInfo `json:"ports,omitempty"`
}

// FormatHost renders a HostResult into the display string
func FormatHost(h HostResult) string {
	var lines []string
	if h.Hardware != "" {
		lines = append(lines, fmt.Sprintf("%s - %s", h.IP, h.Hardware))
	} else {
		lines = append(lines, h.IP)
	}
	for _, p := range h.Ports {
		if p.Banner != "" {
			lines = append(lines, fmt.Sprintf("port %d: %s", p.Port, p.Banner))
		} else {
			// No banner - try to identify the service
			if info := services.Lookup(p.Port); info != nil {
				// Style with muted style - show both name and description
				serviceText := fmt.Sprintf("%s - %s", info.Name, info.Description)
				lines = append(lines, fmt.Sprintf("port %d: %s", p.Port, common.MutedStyle.Render(serviceText)))
			} else {
				lines = append(lines, fmt.Sprintf("port %d", p.Port))
			}
		}
	}
	return strings.Join(lines, "\n")
}

// ParseHost converts a formatted host string (from FormatHost) back to structured data.
func ParseHost(hostStr string) HostResult {
	hostStr = ansiRe.ReplaceAllString(hostStr, "")
	lines := strings.Split(hostStr, "\n")
	if len(lines) == 0 {
		return HostResult{}
	}

	ip, hardware, _ := strings.Cut(lines[0], " - ")
	ip = strings.TrimSpace(ip)
	hardware = strings.TrimSpace(hardware)

	var ports []PortInfo
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "port ") {
			continue
		}
		line = strings.TrimPrefix(line, "port ")
		portStr, banner, _ := strings.Cut(line, ":")
		portNum, err := strconv.Atoi(strings.TrimSpace(portStr))
		if err != nil {
			continue
		}
		ports = append(ports, PortInfo{
			Port:   portNum,
			Banner: strings.TrimSpace(banner),
		})
	}

	return HostResult{
		IP:       ip,
		Hardware: hardware,
		Ports:    ports,
	}
}
