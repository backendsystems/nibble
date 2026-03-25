package scanview

import (
	"net"

	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/stopwatch"
	"charm.land/bubbles/v2/viewport"
	"github.com/backendsystems/nibble/internal/scanner/shared"
)

type Model struct {
	NetworkScan       shared.Scanner
	SelectedIface     net.Interface
	SelectedAddrs     []net.Addr
	Scanning          bool
	ScanComplete      bool
	ShouldPrintFinal  bool
	FoundHosts        []shared.HostResult
	FinalHosts        []shared.HostResult
	TargetCIDR        string // The CIDR being scanned
	PortsScanned      []int  // Ports that were scanned
	IsRescan          bool   // True if rescanning from history
	RescanHistoryPath string // Path to history file to update
	ScannedCount      int
	TotalHosts        int
	NeighborSeen      int
	NeighborTotal     int
	ProgressChan      chan shared.ProgressUpdate
	Progress          progress.Model
	Results           viewport.Model
	Stopwatch         stopwatch.Model
	HelpLineY         int // Y row where the helpline starts, set during render
	HoveredHelpItem   int // -1 means no hover, otherwise index of helpline item
}
