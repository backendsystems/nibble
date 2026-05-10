package scanview

import (
	"net"
	"sort"
	"time"

	"charm.land/bubbles/v2/stopwatch"
	tea "charm.land/bubbletea/v2"
	"github.com/backendsystems/nibble/internal/ports"
	"github.com/backendsystems/nibble/internal/scanner/demo"
	"github.com/backendsystems/nibble/internal/scanner/ip4"
	"github.com/backendsystems/nibble/internal/scanner/shared"
)

type ProgressMsg struct {
	Update shared.ProgressUpdate
}

type CompleteMsg struct{}
type QuitMsg struct{}

func sortHosts(hosts []shared.HostResult) {
	sort.Slice(hosts, func(i, j int) bool {
		a := net.ParseIP(hosts[i].IP).To4()
		b := net.ParseIP(hosts[j].IP).To4()
		if a == nil || b == nil {
			return hosts[i].IP < hosts[j].IP
		}
		for k := range a {
			if a[k] != b[k] {
				return a[k] < b[k]
			}
		}
		return false
	})
}

// appendIfNew appends host to hosts only if no existing entry has the same IP,
// keeping the slice sorted by IP.
func appendIfNew(hosts []shared.HostResult, host *shared.HostResult) []shared.HostResult {
	if host == nil {
		return hosts
	}
	for _, h := range hosts {
		if h.IP == host.IP {
			return hosts
		}
	}
	hosts = append(hosts, *host)
	sortHosts(hosts)
	return hosts
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
	m.ScannedCount = 0
	m.NeighborSeen = 0
	m.NeighborTotal = 0
	m.ProgressChan = make(chan shared.ProgressUpdate, 256)
	m.Stopwatch = stopwatch.New(stopwatch.WithInterval(10 * time.Millisecond))
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
		m.FinalHosts = make([]shared.HostResult, len(m.FoundHosts))
		copy(m.FinalHosts, m.FoundHosts)
		sortHosts(m.FinalHosts)
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
