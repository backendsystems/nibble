package tui

import (
	historyview "github.com/backendsystems/nibble/internal/tui/views/history"
	mainview "github.com/backendsystems/nibble/internal/tui/views/main"
	portsview "github.com/backendsystems/nibble/internal/tui/views/ports"
	scanview "github.com/backendsystems/nibble/internal/tui/views/scan"
	targetview "github.com/backendsystems/nibble/internal/tui/views/target"
	tea "github.com/charmbracelet/bubbletea"
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

func (m *model) Init() tea.Cmd {
	if m.ports.PortPack == "" {
		m.ports.PortPack = "default"
	}
	if m.ports.CustomCursor < 0 || m.ports.CustomCursor > len(m.ports.CustomPorts) {
		m.ports.CustomCursor = len(m.ports.CustomPorts)
	}
	return enterAltScreenCmd()
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	if resize, ok := msg.(tea.WindowSizeMsg); ok {
		m.windowW = resize.Width
		m.windowH = resize.Height
		m.main.CardsPerRow = mainview.CardsPerRow(resize.Width)
		m.main.WindowH = resize.Height
		m.main = m.main.UpdateViewport(scanViewWidth(resize.Width))
		m.scan = m.scan.SetViewportSize(scanViewWidth(m.windowW), m.windowH)
		m.history.WindowW = resize.Width
		m.history.WindowH = resize.Height
		result := m.history.Update(msg)
		m.history = result.Model
		return m, nil
	}

	switch m.active {
	case viewScan:
		return m.handleViewScan(msg)
	case viewPorts:
		return m.handleViewPorts(msg)
	case viewHistory:
		return m.handleViewHistory(msg)
	case viewTarget:
		return m.handleViewTarget(msg)
	case viewMain:
		return m.handleViewMain(msg)
	default:
		return m, nil
	}
}

func (m *model) View() string {
	maxWidth := scanViewWidth(m.windowW)
	switch m.active {
	case viewScan:
		return scanview.Render(m.scan, maxWidth)
	case viewPorts:
		return portsview.Render(&m.ports, maxWidth)
	case viewTarget:
		return targetview.Render(&m.target, maxWidth)
	case viewHistory:
		return historyview.Render(&m.history, maxWidth)
	default:
		return mainview.Render(&m.main, maxWidth)
	}
}
