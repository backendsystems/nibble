package targetview

import "github.com/backendsystems/nibble/internal/scanner/shared"

type Model struct {
	IPInput     string
	IPCursor    int
	CIDRInput   string // e.g. "32", "24", "16"
	CIDRCursor  int
	PortPack    string
	CustomPorts string
	PortCursor  int
	FocusField  int // 0 = IP field, 1 = CIDR field, 2 = ports field
	ErrorMsg    string
	ShowHelp    bool
	NetworkScan shared.Scanner
}
