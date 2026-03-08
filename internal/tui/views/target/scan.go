package targetview

import (
	"github.com/backendsystems/nibble/internal/ports"
	targetports "github.com/backendsystems/nibble/internal/tui/views/target/ports"

	tea "github.com/charmbracelet/bubbletea"
)

// finalizeScan validates IP/CIDR/ports and emits a StartScan result
func (m *Model) finalizeScan(result Result) (Result, tea.Cmd) {
	customPorts := ""
	if m.PortPack == "custom" {
		customPorts = m.CustomPorts // already normalized
	}

	cfg, err := targetports.BuildScanConfig(m.IPInput, m.CIDRInput, m.PortPack, customPorts)
	if err != nil {
		m.ErrorMsg = err.Error()
		m.InCustomPortInput = false
		m.initializeForm()
		return result, m.Form.Init()
	}
	if err := ports.SaveConfig("target", ports.Config{Mode: m.PortPack, Custom: m.CustomPorts}); err != nil {
		m.ErrorMsg = err.Error()
		return result, nil
	}

	m.ErrorMsg = ""
	result.StartScan = true
	result.TargetAddr = cfg.TargetAddr
	result.TotalHosts = cfg.TotalHosts
	result.Ports = cfg.Ports
	return result, nil
}
