package targetview

import (
	"github.com/backendsystems/nibble/internal/scanner/shared"
	targetform "github.com/backendsystems/nibble/internal/tui/views/target/form"
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
}

// initializeForm sets field defaults and (re)builds the huh form bound to this model.
func (m *Model) initializeForm() {
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
	m.Form = targetform.Build(&m.IPInput, &m.CIDRInput, &m.PortPack, m.IPIndex, m.InterfaceInfos)
}

// NewModel creates a new target view model with a form bound to the model's fields
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

	m.initializeForm()
	return m
}
