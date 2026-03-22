package tui

import (
	"net"

	tea "charm.land/bubbletea/v2"
	scannerconfig "github.com/backendsystems/nibble/internal/scanner/config"
	"github.com/backendsystems/nibble/internal/tui/views/common"
	historyview "github.com/backendsystems/nibble/internal/tui/views/history"
	mainview "github.com/backendsystems/nibble/internal/tui/views/main"
	portsview "github.com/backendsystems/nibble/internal/tui/views/ports"
	targetview "github.com/backendsystems/nibble/internal/tui/views/target"
)

func (m *model) handleViewScan(msg tea.Msg) (tea.Model, tea.Cmd) {
	if common.IsMouseMsg(msg) {
		result := m.scan.HandleMouse(msg, scanViewWidth(m.windowW))
		m.scan = result.Model
		if result.Quit {
			return m, tea.Quit
		}
		return m, result.Cmd
	}
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

func (m *model) handleViewPorts(msg tea.Msg) (tea.Model, tea.Cmd) {
	if common.IsMouseMsg(msg) {
		result := m.ports.HandleMouse(msg, scanViewWidth(m.windowW))
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

func (m *model) handleViewHistory(msg tea.Msg) (tea.Model, tea.Cmd) {
	if common.IsMouseMsg(msg) && m.history.Mode == historyview.ViewList {
		result := m.history.HandleMouse(msg, scanViewWidth(m.windowW))
		m.history = result.Model
		if result.Quit {
			m.main.ErrorMsg = ""
			m.active = viewMain
			return m, nil
		}
		return m, result.Cmd
	}
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

func (m *model) handleViewTarget(msg tea.Msg) (tea.Model, tea.Cmd) {
	if common.IsMouseMsg(msg) {
		result, cmd := (&m.target).HandleMouse(msg, scanViewWidth(m.windowW))
		if result.Quit {
			m.main.ErrorMsg = ""
			m.active = viewMain
			return m, nil
		}
		if result.StartScan {
			m.main.ErrorMsg = ""
			scannerconfig.SetPorts(m.scan.NetworkScan, result.Ports)
			nextScan, scanCmd := m.scan.Start(net.Interface{}, nil, result.TotalHosts, result.TargetAddr)
			nextScan = nextScan.SetViewportSize(scanViewWidth(m.windowW), m.windowH)
			m.scan = nextScan
			m.active = viewScan
			m.altScreen = false
			return m, scanCmd
		}
		return m, cmd
	}
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
		m.altScreen = false
		return m, scanCmd
	}
	return m, cmd
}

func (m *model) handleViewMain(msg tea.Msg) (tea.Model, tea.Cmd) {
	var result mainview.UpdateResult
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		result = m.main.Update(msg)
	case tea.MouseMsg:
		result = m.main.HandleMouse(msg)
	default:
		return m, nil
	}
	next := result.Model.UpdateViewport(scanViewWidth(m.windowW))
	if result.Model.Cursor != m.main.Cursor {
		next = next.ScrollToSelected()
	}
	m.main = next
	if result.Quit {
		return m, tea.Quit
	}
	if result.OpenPorts {
		m.ports.ShowHelp = false
		var cmd tea.Cmd
		m.ports, cmd = portsview.Init(m.ports)
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
			WindowW:         m.windowW,
			WindowH:         m.windowH,
			HoveredHelpItem: -1,
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
		m.altScreen = false
		return m, cmd
	}

	return m, nil
}
