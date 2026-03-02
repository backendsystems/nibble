package tui

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/backendsystems/nibble/internal/ports"
	"github.com/backendsystems/nibble/internal/scanner/demo"
	"github.com/backendsystems/nibble/internal/scanner/ip4"
	"github.com/backendsystems/nibble/internal/scanner/shared"
	historyview "github.com/backendsystems/nibble/internal/tui/views/history"
	mainview "github.com/backendsystems/nibble/internal/tui/views/main"
	portsview "github.com/backendsystems/nibble/internal/tui/views/ports"
	scanview "github.com/backendsystems/nibble/internal/tui/views/scan"
	targetview "github.com/backendsystems/nibble/internal/tui/views/target"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/term"
)

type activeView int

const (
	viewMain activeView = iota
	viewPorts
	viewScan
	viewTarget
	viewHistory
)

type model struct {
	active  activeView
	windowW int
	windowH int
	main    mainview.Model
	ports   portsview.Model
	scan    scanview.Model
	target  targetview.Model
	history historyview.Model
}

func Run(networkScanner shared.Scanner, ifaces []net.Interface, addrsByIface map[string][]net.Addr) error {
	cfg, _ := ports.LoadConfig("ports")
	if resolvedPorts, err := resolvePortsConfig(cfg); err == nil {
		switch typed := networkScanner.(type) {
		case *ip4.Scanner:
			typed.Ports = resolvedPorts
		case *demo.Scanner:
			typed.Ports = resolvedPorts
		}
	}

	// Load separate target ports config
	targetCfg, _ := ports.LoadConfig("target")
	targetPack := targetCfg.Mode

	initialWindowW, initialWindowH, initialCardsPerRow := initialLayoutMetrics()
	portsModel, _ := portsview.Prepare(portsview.Model{
		PortPack:    cfg.Mode,
		CustomPorts: cfg.Custom,
		NetworkScan: networkScanner,
	})

	initialModel := model{
		active:  viewMain,
		windowW: initialWindowW,
		windowH: initialWindowH,
		main: mainview.Model{
			Interfaces:   ifaces,
			InterfaceMap: addrsByIface,
			CardsPerRow:  initialCardsPerRow,
		},
		ports: portsModel,
		scan: scanview.Model{
			NetworkScan: networkScanner,
			Progress: progress.New(
				progress.WithScaledGradient("#FFD700", "#B8B000"),
			),
		},
		target: targetview.Model{
			NetworkScan:    networkScanner,
			PortPack:       targetPack,
			CustomPorts:    targetCfg.Custom,
			InterfaceInfos: targetview.BuildInterfaceInfos(ifaces, addrsByIface),
		},
	}
	initialModel.scan = initialModel.scan.SetViewportSize(scanViewWidth(initialModel.windowW), initialModel.windowH)

	prog := tea.NewProgram(initialModel)
	finalModel, err := prog.Run()
	if err != nil {
		return err
	}

	finalState, ok := finalModel.(model)
	if !ok {
		return nil
	}
	if !finalState.scan.ShouldPrintFinal {
		return nil
	}

	fmt.Printf("%s\n", scanview.FinalOutput(finalState.scan))
	return nil
}

func (m model) Init() tea.Cmd {
	if m.ports.PortPack == "" {
		m.ports.PortPack = "default"
	}
	if m.ports.CustomCursor < 0 || m.ports.CustomCursor > len(m.ports.CustomPorts) {
		m.ports.CustomCursor = len(m.ports.CustomPorts)
	}
	return enterAltScreenCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle ctrl+c globally - always quit the entire program
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	if resize, ok := msg.(tea.WindowSizeMsg); ok {
		m.windowW = resize.Width
		m.windowH = resize.Height
		m.main.CardsPerRow = mainview.CardsPerRow(resize.Width)
		m.scan = m.scan.SetViewportSize(scanViewWidth(m.windowW), m.windowH)
		return m, nil
	}

	switch m.active {
	case viewScan:
		result := m.scan.Update(msg)
		if !result.Handled {
			return m, nil
		}
		m.scan = result.Model
		if result.Quit {
			return m, tea.Quit
		}
		return m, result.Cmd
	case viewPorts:
		result := m.ports.Update(msg)
		m.ports = result.Model
		if result.Quit {
			return m, tea.Quit
		}
		if result.Back {
			m.main.ErrorMsg = ""
			m.active = viewMain
			return m, nil
		}
		if result.Done {
			m.main.ErrorMsg = ""
			m.active = viewMain
		}
		return m, result.Cmd
	case viewHistory:
		result := m.history.Update(msg)
		m.history = result.Model

		if result.Quit {
			m.main.ErrorMsg = ""
			m.active = viewMain
			return m, nil
		}

		if result.ScanAllPorts {
			// Launch all-port scan on selected host
			hostIP := result.SelectedHostIP
			targetCIDR := hostIP + "/32"

			// Set scanner to scan all ports
			allPorts := make([]int, 65535)
			for i := 0; i < 65535; i++ {
				allPorts[i] = i + 1
			}

			switch s := m.scan.NetworkScan.(type) {
			case *ip4.Scanner:
				s.Ports = allPorts
			case *demo.Scanner:
				s.Ports = allPorts
			}

			// Start the scan (empty interface for target scans)
			nextScan, scanCmd := m.scan.Start(
				net.Interface{},
				nil,
				1,
				targetCIDR,
			)

			// Mark as rescan to update history file
			nextScan.IsRescan = true
			nextScan.RescanHistoryPath = result.ScanHistoryPath

			nextScan = nextScan.SetViewportSize(scanViewWidth(m.windowW), m.windowH)
			m.scan = nextScan
			m.active = viewScan
			return m, tea.Sequence(exitAltScreenCmd(), scanCmd)
		}

		return m, nil
	case viewTarget:
		result, cmd := (&m.target).Update(msg)
		// Note: m.target is updated in place to preserve form bindings
		if result.Quit {
			m.main.ErrorMsg = ""
			m.active = viewMain
			return m, nil
		}
		if result.StartScan {
			m.main.ErrorMsg = ""

			// Set the resolved ports on the scanner
			switch typed := m.scan.NetworkScan.(type) {
			case *ip4.Scanner:
				typed.Ports = result.Ports
			case *demo.Scanner:
				typed.Ports = result.Ports
			}

			// Start scan with the target configuration
			nextScan, scanCmd := m.scan.Start(
				net.Interface{},
				nil,
				result.TotalHosts,
				result.TargetAddr,
			)
			nextScan = nextScan.SetViewportSize(scanViewWidth(m.windowW), m.windowH)
			m.scan = nextScan
			m.active = viewScan
			return m, tea.Sequence(exitAltScreenCmd(), scanCmd)
		}
		return m, cmd
	case viewMain:
		key, ok := msg.(tea.KeyMsg)
		if !ok {
			return m, nil
		}
		result := m.main.Update(key)
		m.main = result.Model
		if result.Quit {
			return m, tea.Quit
		}
		if result.OpenPorts {
			m.ports.ShowHelp = false
			var cmd tea.Cmd
			m.ports, cmd = portsview.Prepare(m.ports)
			m.active = viewPorts
			return m, cmd
		}
		if result.OpenTarget {
			ipInput := m.target.IPInput
			cidrInput := m.target.CIDRInput
			if cidrInput == "" {
				cidrInput = "32" // Default to single host
			}

			// Pre-fill IP from selected interface if not on target card
			if result.Model.Cursor < len(result.Model.Interfaces) {
				selection, err := mainview.ResolveScanSelection(result.Model.Interfaces, result.Model.Cursor, result.Model.InterfaceMap)
				if err == nil && selection.TargetAddr != "" {
					// Extract IP from CIDR notation (e.g., "192.168.1.0/24" -> "192.168.1.0")
					ip := selection.TargetAddr
					if idx := strings.Index(ip, "/"); idx != -1 {
						ip = ip[:idx]
					}
					ipInput = ip
				}
			}

			m.target = targetview.Model{
				NetworkScan:    m.target.NetworkScan,
				IPInput:        ipInput,
				CIDRInput:      cidrInput,
				PortPack:       m.target.PortPack,
				CustomPorts:    m.target.CustomPorts,
				InterfaceInfos: targetview.BuildInterfaceInfos(result.Model.Interfaces, result.Model.InterfaceMap),
			}
			m.active = viewTarget
			return m, (&m.target).Init()
		}
		if result.OpenHistory {
			m.history = historyview.Model{
				WindowW: m.windowW,
				WindowH: m.windowH,
			}
			m.active = viewHistory
			return m, m.history.Init()
		}
		if result.StartScan {
			m.main.ErrorMsg = ""
			nextScan, cmd := m.scan.Start(
				result.Selection.Iface,
				result.Selection.Addrs,
				result.Selection.TotalHosts,
				result.Selection.TargetAddr,
			)
			nextScan = nextScan.SetViewportSize(scanViewWidth(m.windowW), m.windowH)
			m.scan = nextScan
			m.active = viewScan
			return m, tea.Sequence(exitAltScreenCmd(), cmd)
		}
		return m, nil
	default:
		return m, nil
	}
}

func (m model) View() string {
	maxWidth := scanViewWidth(m.windowW)
	switch m.active {
	case viewScan:
		return scanview.Render(m.scan, maxWidth)
	case viewPorts:
		return portsview.Render(m.ports, maxWidth)
	case viewTarget:
		return targetview.Render(m.target, maxWidth)
	case viewHistory:
		return historyview.Render(m.history, maxWidth)
	default:
		return mainview.Render(m.main, maxWidth)
	}
}

func initialLayoutMetrics() (windowW int, windowH int, cardsPerRow int) {
	cardsPerRow = 1
	fd := os.Stdout.Fd()
	if !term.IsTerminal(fd) {
		return 0, 0, cardsPerRow
	}

	width, height, err := term.GetSize(fd)
	if err != nil || width <= 0 || height <= 0 {
		return 0, 0, cardsPerRow
	}

	return width, height, mainview.CardsPerRow(width)
}

func scanViewWidth(windowW int) int {
	maxWidth := 72
	if windowW > 8 {
		maxWidth = windowW - 4
	}
	return maxWidth
}

func enterAltScreenCmd() tea.Cmd {
	return func() tea.Msg { return tea.EnterAltScreen() }
}

func exitAltScreenCmd() tea.Cmd {
	return func() tea.Msg { return tea.ExitAltScreen() }
}

func resolvePortsConfig(cfg ports.Config) ([]int, error) {
	switch cfg.Mode {
	case "all":
		return ports.ParseList("1-65535")
	case "custom":
		resolved, err := ports.ParseList(cfg.Custom)
		if err != nil {
			return nil, err
		}
		if resolved == nil {
			return []int{}, nil
		}
		return resolved, nil
	default:
		// nil means "use scanner defaults"
		return nil, nil
	}
}
