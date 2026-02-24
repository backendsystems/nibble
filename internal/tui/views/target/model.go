package targetview

import (
	"net"

	"github.com/backendsystems/nibble/internal/scanner/shared"
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

// NewModel creates a new target view model with a form bound to the model's fields
// Deprecated: Use struct literal initialization and call Init() instead
func NewModel(networkScan shared.Scanner, ipInput, cidrInput, portPack, customPorts string, ifaces []net.Interface) Model {
	m := Model{
		IPInput:     ipInput,
		CIDRInput:   cidrInput,
		PortPack:    portPack,
		CustomPorts: customPorts,
		NetworkScan: networkScan,
		Interfaces:  ifaces,
	}

	m.initializeForm()
	return m
}
