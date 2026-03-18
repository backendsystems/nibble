package targetview

import (
	"github.com/backendsystems/nibble/internal/scanner/shared"
	"github.com/backendsystems/nibble/internal/tui/views/common"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// CycleInterfaceIP cycles to the next or previous interface IP
// forward=true moves to next, forward=false moves to previous
func (m *Model) CycleInterfaceIP(forward bool) {
	if len(m.InterfaceInfos) == 0 {
		return
	}
	if forward {
		m.IPIndex = (m.IPIndex + 1) % len(m.InterfaceInfos)
	} else {
		m.IPIndex = (m.IPIndex - 1 + len(m.InterfaceInfos)) % len(m.InterfaceInfos)
	}
	m.IPInput = m.InterfaceInfos[m.IPIndex].IP
	m.IPTextInput.SetValue(m.IPInput)
	m.IPTextInput.CursorEnd()
}

// initializeInputs sets up textinput fields from model state.
func (m *Model) initializeInputs() {
	if len(m.InterfaceIPs) == 0 && len(m.InterfaceInfos) > 0 {
		m.InterfaceIPs = buildInterfaceIPs(m.InterfaceInfos)
	}
	if len(m.InterfaceInfos) > 0 {
		if m.IPIndex < 0 || m.IPIndex >= len(m.InterfaceInfos) {
			m.IPIndex = 0
		}
		for i, info := range m.InterfaceInfos {
			if info.IP == m.IPInput {
				m.IPIndex = i
				break
			}
		}
		if m.IPInput == "" {
			m.IPInput = m.InterfaceInfos[m.IPIndex].IP
		}
	}
	if m.CIDRInput == "" {
		m.CIDRInput = "32"
	}
	if m.PortPack == "" {
		m.PortPack = "default"
	}

	// Find port mode index
	m.PortModeIndex = 0
	for i, opt := range portModeOptions {
		if opt.Value == m.PortPack {
			m.PortModeIndex = i
			break
		}
	}

	// Build IP textinput
	m.IPTextInput = newFieldInput(15)
	m.IPTextInput.Placeholder = "192.168.1.0"
	m.IPTextInput.SetValue(m.IPInput)
	m.IPTextInput.CursorEnd()

	// Build CIDR textinput
	m.CIDRTextInput = newFieldInput(2)
	m.CIDRTextInput.Placeholder = "32"
	m.CIDRTextInput.SetValue(m.CIDRInput)
	m.CIDRTextInput.CursorEnd()
}

func newFieldInput(charLimit int) textinput.Model {
	ti := textinput.New()
	ti.CharLimit = charLimit
	ti.Prompt = "> "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(common.Color.Selection)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(common.Color.Selection)
	ti.TextStyle = lipgloss.NewStyle().Foreground(common.Color.Info)
	return ti
}

// NewModel creates a new target view model with inputs bound to the model's fields
// Deprecated: Use struct literal initialization and call Init() instead
func NewModel(networkScan shared.Scanner, ipInput, cidrInput, portPack, customPorts string, interfaceInfos []InterfaceInfo) Model {
	m := Model{
		IPInput:        ipInput,
		CIDRInput:      cidrInput,
		PortPack:       portPack,
		CustomPorts:    customPorts,
		NetworkScan:    networkScan,
		InterfaceInfos: interfaceInfos,
	}

	m.initializeInputs()
	return m
}

// focusField focuses the given field index and blurs others.
func (m *Model) focusField(field int) tea.Cmd {
	m.FocusedField = field
	var cmd tea.Cmd
	switch field {
	case fieldIP:
		cmd = m.IPTextInput.Focus()
		m.CIDRTextInput.Blur()
	case fieldCIDR:
		m.IPTextInput.Blur()
		cmd = m.CIDRTextInput.Focus()
	case fieldPortMode:
		m.IPTextInput.Blur()
		m.CIDRTextInput.Blur()
	}
	return cmd
}

