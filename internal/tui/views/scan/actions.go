package scanview

import (
	"net"
	"strings"
	"time"

	"github.com/backendsystems/nibble/internal/ports"
	"github.com/backendsystems/nibble/internal/scanner/demo"
	"github.com/backendsystems/nibble/internal/scanner/ip4"
	"github.com/backendsystems/nibble/internal/scanner/shared"
	"github.com/charmbracelet/bubbles/stopwatch"
	tea "github.com/charmbracelet/bubbletea"
)

type ProgressMsg struct {
	Update shared.ProgressUpdate
}

type CompleteMsg struct{}
type QuitMsg struct{}

// appendIfNew appends host to hosts only if no existing entry has the same IP.
func appendIfNew(hosts []string, host string) []string {
	newIP := hostIP(host)
	for _, h := range hosts {
		if hostIP(h) == newIP {
			return hosts
		}
	}
	return append(hosts, host)
}

// hostIP extracts the IP address from the first line of a host string.
// Host format is "IP - vendor" or just "IP".
func hostIP(host string) string {
	line, _, _ := strings.Cut(host, "\n")
	ip, _, _ := strings.Cut(line, " - ")
	return strings.TrimSpace(ip)
}

func ListenForProgress(progressChan <-chan shared.ProgressUpdate) tea.Cmd {
	return func() tea.Msg {
		progress, ok := <-progressChan
		if !ok {
			return CompleteMsg{}
		}
		return ProgressMsg{Update: progress}
	}
}

func continueScanLoop(m Model) tea.Cmd {
	return ListenForProgress(m.ProgressChan)
}

func PerformScan(networkScanner shared.Scanner, ifaceName, targetAddr string, progressChan chan shared.ProgressUpdate) tea.Cmd {
	return func() tea.Msg {
		go networkScanner.ScanNetwork(ifaceName, targetAddr, progressChan)
		return ListenForProgress(progressChan)()
	}
}

func (m Model) Start(iface net.Interface, addrs []net.Addr, totalHosts int, targetAddr string) (Model, tea.Cmd) {
	m.SelectedIface = iface
	m.SelectedAddrs = addrs
	m.TotalHosts = totalHosts
	m.TargetCIDR = targetAddr
	m.Scanning = true
	m.ScanComplete = false
	m.ShouldPrintFinal = false
	m.FoundHosts = nil
	m.FinalHosts = nil
	m.FoundHostsData = nil
	m.FinalHostsData = nil
	m.ScannedCount = 0
	m.NeighborSeen = 0
	m.NeighborTotal = 0
	m.ProgressChan = make(chan shared.ProgressUpdate, 256)
	m.Stopwatch = stopwatch.NewWithInterval(10 * time.Millisecond)
	m.PortsScanned = getScannedPorts(m.NetworkScan)
	m = m.RefreshResults(false)

	return m, tea.Batch(m.Stopwatch.Init(), PerformScan(m.NetworkScan, iface.Name, targetAddr, m.ProgressChan))
}

func sendQuitMsg() tea.Cmd {
	return func() tea.Msg { return QuitMsg{} }
}

// prepareForExit preserves discovered hosts for optional final output, then clears
// the live viewport data to avoid duplicate terminal content on exit.
func prepareForExit(m Model, shouldPrint bool) Model {
	m.ShouldPrintFinal = shouldPrint
	if len(m.FinalHosts) == 0 && len(m.FoundHosts) > 0 {
		m.FinalHosts = append([]string(nil), m.FoundHosts...)
	}
	m.FoundHosts = nil
	m.Results.SetContent("")
	return m
}

// getScannedPorts extracts the ports being scanned from the scanner
func getScannedPorts(scanner shared.Scanner) []int {
	switch s := scanner.(type) {
	case *ip4.Scanner:
		if s.Ports == nil {
			return ports.DefaultPorts()
		}
		return s.Ports
	case *demo.Scanner:
		if s.Ports == nil {
			return ports.DefaultPorts()
		}
		return s.Ports
	default:
		return nil
	}
}
