package targetview

import (
	"strings"

	mainview "github.com/backendsystems/nibble/internal/tui/views/main"
)

func BuildFromMain(current Model, mainModel mainview.Model) Model {
	ipInput := current.IPInput
	cidrInput := current.CIDRInput
	if cidrInput == "" {
		cidrInput = "32"
	}

	if mainModel.Cursor < len(mainModel.Interfaces) {
		selection, err := mainview.ResolveScanSelection(mainModel.Interfaces, mainModel.Cursor, mainModel.InterfaceMap)
		if err == nil && selection.TargetAddr != "" {
			ip := selection.TargetAddr
			if idx := strings.Index(ip, "/"); idx != -1 {
				ip = ip[:idx]
			}
			ipInput = ip
		}
	}

	return Model{
		NetworkScan:    current.NetworkScan,
		IPInput:        ipInput,
		CIDRInput:      cidrInput,
		PortPack:       current.PortPack,
		CustomPorts:    current.CustomPorts,
		InterfaceInfos: BuildInterfaceInfos(mainModel.Interfaces, mainModel.InterfaceMap),
	}
}
