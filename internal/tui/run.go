package tui

import (
	"fmt"
	"net"
	"os"

	"charm.land/lipgloss/v2"
	scannerconfig "github.com/backendsystems/nibble/internal/scanner/config"
	mainview "github.com/backendsystems/nibble/internal/tui/views/main"

	portsview "github.com/backendsystems/nibble/internal/tui/views/ports"
	scanview "github.com/backendsystems/nibble/internal/tui/views/scan"
	targetview "github.com/backendsystems/nibble/internal/tui/views/target"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"github.com/backendsystems/nibble/internal/ports"
	"github.com/backendsystems/nibble/internal/scanner/shared"
	"github.com/charmbracelet/x/term"
)

func Run(networkScanner shared.Scanner, ifaces []net.Interface, addrsByIface map[string][]net.Addr, dockerIfaces map[string]struct{}) error {
	cfg, _ := ports.LoadConfig("ports")
	if resolvedPorts, err := resolvePortsConfig(cfg); err == nil {
		scannerconfig.SetPorts(networkScanner, resolvedPorts)
	}

	targetCfg, _ := ports.LoadConfig("target")
	targetPack := targetCfg.Mode

	cardWidth := mainview.ComputeCardWidth(ifaces, addrsByIface)
	initialWindowW, initialWindowH, initialCardsPerRow := initialLayoutMetrics(cardWidth)
	portsModel, _ := portsview.Init(portsview.Model{
		PortPack:        cfg.Mode,
		CustomPorts:     cfg.Custom,
		NetworkScan:     networkScanner,
		HoveredHelpItem: -1,
	})

	initialModel := model{
		active:  viewMain,
		windowW: initialWindowW,
		windowH: initialWindowH,
		main: mainview.Model{
			Interfaces:      ifaces,
			InterfaceMap:    addrsByIface,
			DockerIfaces:    dockerIfaces,
			CardsPerRow:     initialCardsPerRow,
			CardWidth:       cardWidth,
			WindowH:         initialWindowH,
			HoveredHelpItem: -1,
		},
		ports: portsModel,
		scan: scanview.Model{
			NetworkScan:     networkScanner,
			Progress:        progress.New(progress.WithColors(lipgloss.Yellow), progress.WithFillCharacters(progress.DefaultFullCharFullBlock, ' ')),
			HoveredHelpItem: -1,
		},
		target: targetview.Model{
			NetworkScan:     networkScanner,
			PortPack:        targetPack,
			CustomPorts:     targetCfg.Custom,
			IPInput:         targetCfg.IP,
			CIDRInput:       targetCfg.CIDR,
			InterfaceInfos:  targetview.BuildInterfaceInfos(ifaces, addrsByIface),
			HoveredHelpItem: -1,
			WindowH:         initialWindowH,
		},
	}
	initialModel.scan = initialModel.scan.SetViewportSize(scanViewWidth(initialModel.windowW), initialModel.windowH)

	prog := tea.NewProgram(&initialModel)
	finalModel, err := prog.Run()
	if err != nil {
		return err
	}

	finalState, ok := finalModel.(*model)
	if !ok {
		return nil
	}
	if !finalState.scan.ShouldPrintFinal {
		return nil
	}

	fmt.Printf("%s\n", scanview.FinalOutput(finalState.scan))
	return nil
}

func initialLayoutMetrics(cardWidth int) (windowW int, windowH int, cardsPerRow int) {
	cardsPerRow = 1
	fd := os.Stdout.Fd()
	if !term.IsTerminal(fd) {
		return 0, 0, cardsPerRow
	}

	width, height, err := term.GetSize(fd)
	if err != nil || width <= 0 || height <= 0 {
		return 0, 0, cardsPerRow
	}

	return width, height, mainview.CardsPerRow(scanViewWidth(width), cardWidth)
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
