package targetview

import (
	"github.com/backendsystems/nibble/internal/scanner/shared"
	"github.com/backendsystems/nibble/internal/tui/views/common"
	"github.com/charmbracelet/huh"
)

type InterfaceInfo struct {
	Name string
	IP   string
}

type Model struct {
	ShowHelp          bool
	Form              *huh.Form
	IPInput           string
	CIDRInput         string // e.g. "32", "24", "16"
	PortPack          string
	CustomPorts       string                 // seed/persistence; synced from PortInput on submit
	PortInput         common.CustomPortInput // replaces huh custom_ports field
	InCustomPortInput bool                   // true when showing port textinput stage
	ErrorMsg          string
	NetworkScan       shared.Scanner
	InterfaceIPs      []string        // Available interface IPs (deprecated, use InterfaceInfos)
	InterfaceInfos    []InterfaceInfo // Available interfaces with names and IPs
	IPIndex           int             // Current index in InterfaceInfos
}
