package history

import "time"

// ScanHistory represents a complete scan record
type ScanHistory struct {
	Version      string       `json:"version"`
	ScanMetadata ScanMetadata `json:"scan_metadata"`
	ScanResults  ScanResults  `json:"scan_results"`
}

// ScanSummary is a lightweight view of a scan for list/tree rendering.
type ScanSummary struct {
	ScanMetadata ScanMetadataSummary `json:"scan_metadata"`
	ScanResults  ScanResultsSummary  `json:"scan_results"`
}

// ScanMetadata holds information about when and how the scan was performed
type ScanMetadata struct {
	Created         time.Time `json:"created"`
	Updated         time.Time `json:"updated"`
	DurationSeconds float64   `json:"duration_seconds"`
	InterfaceName   string    `json:"interface_name"`
	TargetCIDR      string    `json:"target_cidr"`
	PortsScanned    []int     `json:"ports_scanned"`
}

// ScanMetadataSummary holds only the metadata needed for lightweight loading.
type ScanMetadataSummary struct {
	Created       time.Time `json:"created"`
	Updated       time.Time `json:"updated"`
	InterfaceName string    `json:"interface_name"`
	TargetCIDR    string    `json:"target_cidr"`
}

// ScanResults holds the results of the scan
type ScanResults struct {
	TotalHostsScanned int          `json:"total_hosts_scanned"`
	HostsFound        int          `json:"hosts_found"`
	PortsFound        int          `json:"ports_found"`
	Hosts             []HostResult `json:"hosts"`
}

// ScanResultsSummary holds only aggregate counts for lightweight loading.
type ScanResultsSummary struct {
	HostsFound int `json:"hosts_found"`
	PortsFound int `json:"ports_found"`
}

// HostResult holds all scan info for a single host
type HostResult struct {
	IP           string     `json:"ip"`
	Hardware     string     `json:"hardware"`
	MAC          string     `json:"mac"`
	Ports        []PortInfo `json:"ports"`
	LastScanned  time.Time  `json:"last_scanned"`
	PortsScanned []int      `json:"ports_scanned"`
}

// PortInfo holds a port number and its service banner
type PortInfo struct {
	Port   int    `json:"port"`
	Banner string `json:"banner"`
}
