package targetview

import (
	"net"

	"github.com/backendsystems/nibble/internal/scanner/shared"
	"github.com/charmbracelet/huh"
)

type Model struct {
	Form         *huh.Form
	IPInput      string
	CIDRInput    string // e.g. "32", "24", "16"
	PortPack     string
	CustomPorts  string
	ErrorMsg     string
	NetworkScan  shared.Scanner
	InterfaceIPs []string        // Available interface IPs
	IPIndex      int             // Current index in InterfaceIPs
	Interfaces   []net.Interface // Store interfaces for later use
}
