package targetview

import (
	"github.com/backendsystems/nibble/internal/ports"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

type Result struct {
	Cmd        tea.Cmd
	Quit       bool
	StartScan  bool   // Signal to start scan with saved config
	TargetAddr string // Target address in CIDR notation
	TotalHosts int    // Number of hosts to scan
	Ports      []int  // Resolved port list to scan
}

// Init returns the initialization command for the form
func (m *Model) Init() tea.Cmd {
	if m.Form == nil {
		m.initializeForm()
	}
	return m.Form.Init()
}

// Update handles tea.Msg and delegates to the form
// The model is updated in place to preserve form bindings
func (m *Model) Update(msg tea.Msg) (Result, tea.Cmd) {
	result := Result{}

	// Handle keyboard input
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		keyResult, keyCmd := handleKeyPress(m, keyMsg)

		// If the key handler returned early (quit, help, delete, navigation)
		if keyResult.Quit || keyCmd != nil {
			return keyResult, keyCmd
		}

		// Special case: convert k/w to up, j/s to down for port_mode select
		if m.Form != nil {
			focused := m.Form.GetFocusedField()
			if focused != nil && focused.GetKey() == "port_mode" {
				switch keyMsg.String() {
				case "k", "w":
					msg = tea.KeyMsg{Type: tea.KeyUp}
				case "j", "s":
					msg = tea.KeyMsg{Type: tea.KeyDown}
				}
			}
		}
	}

	if m.Form == nil {
		return result, nil
	}

	// Delegate update to the form
	formModel, cmd := m.Form.Update(msg)
	if f, ok := formModel.(*huh.Form); ok {
		m.Form = f
	}
	result.Cmd = cmd

	// Check if form is completed
	if m.Form.State == huh.StateCompleted {
		// Extract values directly from form fields
		ipInput := m.Form.GetString("ip")
		cidrInput := m.Form.GetString("cidr")
		portPack := m.Form.GetString("port_mode")
		customPorts := m.Form.GetString("custom_ports")
		savedCustomPorts := customPorts

		if portPack == "custom" {
			normalized, err := normalizeCustomPorts(customPorts)
			if err != nil {
				m.ErrorMsg = err.Error()
				return result, nil
			}
			savedCustomPorts = normalized
			customPorts = normalized
		}

		// Build scan configuration
		targetAddr, totalHosts, resolvedPorts, err := buildScanConfig(ipInput, cidrInput, portPack, customPorts)
		if err != nil {
			m.ErrorMsg = err.Error()
			return result, nil
		}

		// Save port configuration
		if err := ports.SaveConfig("target", ports.Config{Mode: portPack, Custom: savedCustomPorts}); err != nil {
			m.ErrorMsg = err.Error()
			return result, nil
		}

		// Keep target model state in sync
		m.IPInput = ipInput
		m.CIDRInput = cidrInput
		m.PortPack = portPack
		m.CustomPorts = savedCustomPorts

		m.ErrorMsg = ""
		result.StartScan = true
		result.TargetAddr = targetAddr
		result.TotalHosts = totalHosts
		result.Ports = resolvedPorts
	}

	// Check if form is aborted
	if m.Form.State == huh.StateAborted {
		result.Quit = true
	}

	return result, cmd
}
