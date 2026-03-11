package tui

import (
	"fmt"
	"net"
	"os"

	scannerconfig "github.com/backendsystems/nibble/internal/scanner/config"
	"github.com/backendsystems/nibble/internal/tui/views/common"
	mainview "github.com/backendsystems/nibble/internal/tui/views/main"

	portsview "github.com/backendsystems/nibble/internal/tui/views/ports"
	scanview "github.com/backendsystems/nibble/internal/tui/views/scan"
	targetview "github.com/backendsystems/nibble/internal/tui/views/target"

	"github.com/backendsystems/nibble/internal/ports"
	"github.com/backendsystems/nibble/internal/scanner/shared"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/term"
)

func Run(networkScanner shared.Scanner, ifaces []net.Interface, addrsByIface map[string][]net.Addr) error {
	cfg, _ := ports.LoadConfig("ports")
	if resolvedPorts, err := resolvePortsConfig(cfg); err == nil {
		scannerconfig.SetPorts(networkScanner, resolvedPorts)
	}

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
			Interfaces:      ifaces,
			InterfaceMap:    addrsByIface,
			CardsPerRow:     initialCardsPerRow,
			WindowH:         initialWindowH,
			HoveredHelpItem: -1,
		},
		ports: portsModel,
		scan: scanview.Model{
			NetworkScan: networkScanner,
			Progress:    progress.New(progress.WithSolidFill(string(common.Color.Selection))),
		},
		target: targetview.Model{
			NetworkScan:    networkScanner,
			PortPack:       targetPack,
			CustomPorts:    targetCfg.Custom,
			InterfaceInfos: targetview.BuildInterfaceInfos(ifaces, addrsByIface),
		},
	}
	initialModel.scan = initialModel.scan.SetViewportSize(scanViewWidth(initialModel.windowW), initialModel.windowH)

	prog := tea.NewProgram(initialModel, tea.WithMouseAllMotion())
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
		return nil, nil
	}
}
