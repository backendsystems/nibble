package targetview

import (
	"github.com/backendsystems/nibble/internal/scanner/shared"
	"github.com/backendsystems/nibble/internal/tui/views/common"
	"github.com/charmbracelet/bubbles/textinput"
)

// InterfaceInfo holds a network interface name and its IPv4 address.
type InterfaceInfo struct {
	Name string
	IP   string
}

// portModeOptions holds the ordered list of port mode values.
var portModeOptions = []struct {
	Label string
	Value string
}{
	{"Default", "default"},
	{"All ports (1-65535)", "all"},
	{"Custom ports", "custom"},
}

const (
	fieldIP       = 0
	fieldCIDR     = 1
	fieldPortMode = 2
	fieldCount    = 3
)

type Model struct {
	ShowHelp          bool
	FocusedField      int            // 0=IP, 1=CIDR, 2=PortMode
	IPTextInput       textinput.Model
	CIDRTextInput     textinput.Model
	IPInput           string
	CIDRInput         string // e.g. "32", "24", "16"
	PortPack          string
	PortModeIndex     int                    // index into portModeOptions
	CustomPorts       string                 // seed/persistence; synced from PortInput on submit
	PortInput         common.CustomPortInput // replaces huh custom_ports field
	InCustomPortInput bool                   // true when showing port textinput stage
	ErrorMsg          string
	NetworkScan       shared.Scanner
	InterfaceIPs      []string        // Available interface IPs (deprecated, use InterfaceInfos)
	InterfaceInfos    []InterfaceInfo // Available interfaces with names and IPs
	IPIndex           int             // Current index in InterfaceInfos
	HoveredHelpItem   int             // -1 means no hover, otherwise index of helpline item
	HelpLineY         int             // Y row where the helpline starts, set during render
	FieldY            [fieldCount]int // Y row where each field starts, set during render
}
