package targetview

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	"github.com/backendsystems/nibble/internal/scanner/shared"
	"github.com/backendsystems/nibble/internal/tui/views/common"
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
	fieldInterface = 0
	fieldIP        = 1
	fieldCIDR      = 2
	fieldPortMode  = 3
	fieldCount     = 4
)

type Model struct {
	ShowHelp          bool
	FocusedField      int // 0=Interface, 1=IP, 2=CIDR, 3=PortMode
	IPTextInput       textinput.Model
	CIDRTextInput     textinput.Model
	IPInput           string
	CIDRInput         string // e.g. "32", "24", "16"
	IPIsCustom        bool   // true when IP was manually typed and doesn't match any interface
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

// ipMatchesInterface returns true if ip shares the first three octets with any known interface IP.
func (m *Model) ipMatchesInterface(ip string) bool {
	prefix := ipPrefix(ip)
	if prefix == "" {
		return false
	}
	for _, info := range m.InterfaceInfos {
		if ipPrefix(info.IP) == prefix {
			return true
		}
	}
	return false
}

// ipPrefix returns the first three octets of an IPv4 address as a string, or "" if not parseable.
func ipPrefix(ip string) string {
	first := strings.Index(ip, ".")
	if first < 0 {
		return ""
	}
	second := strings.Index(ip[first+1:], ".")
	if second < 0 {
		return ""
	}
	third := strings.Index(ip[first+1+second+1:], ".")
	if third < 0 {
		return ""
	}
	return ip[:first+1+second+1+third]
}
