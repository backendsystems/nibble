package targetview

import (
	"github.com/backendsystems/nibble/internal/scanner/shared"
	"github.com/charmbracelet/huh"
)

type InterfaceInfo struct {
	Name string
	IP   string
}

type Model struct {
	ShowHelp       bool
	Form           *huh.Form
	IPInput        string
	CIDRInput      string // e.g. "32", "24", "16"
	PortPack       string
	CustomPorts    string
	ErrorMsg       string
	NetworkScan    shared.Scanner
	InterfaceIPs   []string        // Available interface IPs (deprecated, use InterfaceInfos)
	InterfaceInfos []InterfaceInfo // Available interfaces with names and IPs
	IPIndex        int             // Current index in InterfaceInfos
}
