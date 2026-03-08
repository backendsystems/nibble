package scanview

import (
	"net"

	"github.com/backendsystems/nibble/internal/scanner/shared"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/stopwatch"
	"github.com/charmbracelet/bubbles/viewport"
)

type Model struct {
	NetworkScan       shared.Scanner
	SelectedIface     net.Interface
	SelectedAddrs     []net.Addr
	Scanning          bool
	ScanComplete      bool
	ShouldPrintFinal  bool
	FoundHosts        []string
	FinalHosts        []string
	FoundHostsData    []shared.HostResult // Structured data for history
	FinalHostsData    []shared.HostResult // Structured data for history
	TargetCIDR        string              // The CIDR being scanned
	PortsScanned      []int               // Ports that were scanned
	IsRescan          bool                // True if rescanning from history
	RescanHistoryPath string              // Path to history file to update
	ScannedCount      int
	TotalHosts        int
	NeighborSeen      int
	NeighborTotal     int
	ProgressChan      chan shared.ProgressUpdate
	Progress          progress.Model
	Results           viewport.Model
	Stopwatch         stopwatch.Model
}
