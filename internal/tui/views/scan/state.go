package scanview

import (
	"net"

	"github.com/backendsystems/nibble/internal/scanner/shared"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"
)

type Model struct {
	NetworkScan      shared.Scanner
	SelectedIface    net.Interface
	SelectedAddrs    []net.Addr
	Scanning         bool
	ScanComplete     bool
	ShouldPrintFinal bool
	FoundHosts       []string
	FinalHosts       []string
	ScannedCount     int
	TotalHosts       int
	NeighborSeen     int
	NeighborTotal    int
	ProgressChan     chan shared.ProgressUpdate
	Progress         progress.Model
	Results          viewport.Model
}
