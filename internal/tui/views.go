package tui

import (
	scannerconfig "github.com/backendsystems/nibble/internal/scanner/config"
	historyview "github.com/backendsystems/nibble/internal/tui/views/history"
	mainview "github.com/backendsystems/nibble/internal/tui/views/main"
	portsview "github.com/backendsystems/nibble/internal/tui/views/ports"
	targetview "github.com/backendsystems/nibble/internal/tui/views/target"
	tea "github.com/charmbracelet/bubbletea"
	"net"
)

func (m model) handleViewScan(msg tea.Msg) (tea.Model, tea.Cmd) {
	result := m.scan.Update(msg)
	if !result.Handled {
		return m, nil
	}
	m.scan = result.Model
	if result.Quit {
		return m, tea.Quit
	}
	return m, result.Cmd
}

func (m model) handleViewPorts(msg tea.Msg) (tea.Model, tea.Cmd) {
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
}

func (m model) handleViewHistory(msg tea.Msg) (tea.Model, tea.Cmd) {
	result := m.history.Update(msg)
	m.history = result.Model

	if result.Quit {
		m.main.ErrorMsg = ""
		m.active = viewMain
		return m, nil
	}

	if result.ScanAllPorts {
		var scanCmd tea.Cmd
		m.history, scanCmd = historyview.StartDetailRescan(m.history, m.scan.NetworkScan, result.SelectedHostIP)
		return m, scanCmd
	}

	return m, result.Cmd
}

func (m model) handleViewTarget(msg tea.Msg) (tea.Model, tea.Cmd) {
	result, cmd := (&m.target).Update(msg)
	if result.Quit {
		m.main.ErrorMsg = ""
		m.active = viewMain
		return m, nil
	}
	if result.StartScan {
		m.main.ErrorMsg = ""
		scannerconfig.SetPorts(m.scan.NetworkScan, result.Ports)

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
}

func (m model) handleViewMain(msg tea.Msg) (tea.Model, tea.Cmd) {
	var result mainview.UpdateResult
	switch msg := msg.(type) {
	case tea.KeyMsg:
		result = m.main.Update(msg)
	case tea.MouseMsg:
		result = m.main.HandleMouse(msg)
	default:
		return m, nil
	}
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
		m.target = targetview.BuildFromMain(m.target, result.Model)
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
}
