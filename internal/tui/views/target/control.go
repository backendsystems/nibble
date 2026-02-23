package targetview

import (
	"errors"
	"fmt"
	"net"

	mainview "github.com/backendsystems/nibble/internal/tui/views/main"
	"github.com/backendsystems/nibble/internal/ports"
	"github.com/backendsystems/nibble/internal/scanner/shared"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	errInvalidCIDR      = errors.New("invalid CIDR (e.g. 10.0.0.0/24)")
	errSubnetTooLarge   = errors.New("subnet too large (min /16)")
)

type Action int

const (
	ActionNone Action = iota
	ActionQuit
	ActionOpenHelp
	ActionCloseHelp
	ActionTabNext
	ActionConfirm
	ActionMoveLeft
	ActionMoveRight
	ActionMoveHome
	ActionMoveEnd
	ActionBackspace
	ActionDeleteAll
)

type Result struct {
	Model      Model
	Quit       bool
	Done       bool
	StartScan  bool
	Selection  mainview.ScanSelection
	SavePorts  bool // Signal to save port config
}

const helpText = "Tab/Shift+Tab: switch field • ←/→ a/d: move cursor • type • backspace: remove • Enter: confirm • Esc: back • ?: help • ctrl+c: quit"

func HandleKey(showHelp bool, key string) Action {
	if showHelp {
		return ActionCloseHelp
	}

	switch key {
	case "ctrl+c":
		return ActionQuit
	case "?":
		return ActionOpenHelp
	case "escape":
		return ActionQuit
	case "tab":
		return ActionTabNext
	case "shift+tab":
		return ActionTabNext
	case "enter":
		return ActionConfirm
	case "left", "a":
		return ActionMoveLeft
	case "right", "d":
		return ActionMoveRight
	case "home":
		return ActionMoveHome
	case "end":
		return ActionMoveEnd
	case "backspace":
		return ActionBackspace
	case "delete":
		return ActionDeleteAll
	default:
		return ActionNone
	}
}

func (m Model) Update(msg tea.KeyMsg) Result {
	result := Result{Model: m}

	action := HandleKey(m.ShowHelp, msg.String())

	switch action {
	case ActionQuit:
		result.Done = true
	case ActionOpenHelp:
		result.Model.ShowHelp = true
	case ActionCloseHelp:
		result.Model.ShowHelp = false
	case ActionTabNext:
		result.Model.FocusField = (result.Model.FocusField + 1) % 3
	case ActionConfirm:
		selection, err := validateAndBuild(result.Model)
		if err != nil {
			result.Model.ErrorMsg = err.Error()
			return result
		}
		result.Model.ErrorMsg = ""
		result.StartScan = true
		result.Selection = selection
		result.SavePorts = true // Signal to save target ports config
	case ActionMoveLeft:
		if result.Model.FocusField == 0 {
			if result.Model.IPCursor > 0 {
				result.Model.IPCursor--
			}
		} else if result.Model.FocusField == 1 {
			if result.Model.CIDRCursor > 0 {
				result.Model.CIDRCursor--
			}
		} else {
			if result.Model.PortCursor > 0 {
				result.Model.PortCursor--
			}
		}
	case ActionMoveRight:
		if result.Model.FocusField == 0 {
			if result.Model.IPCursor < len(result.Model.IPInput) {
				result.Model.IPCursor++
			}
		} else if result.Model.FocusField == 1 {
			if result.Model.CIDRCursor < len(result.Model.CIDRInput) {
				result.Model.CIDRCursor++
			}
		} else {
			if result.Model.PortCursor < len(result.Model.CustomPorts) {
				result.Model.PortCursor++
			}
		}
	case ActionMoveHome:
		if result.Model.FocusField == 0 {
			result.Model.IPCursor = 0
		} else if result.Model.FocusField == 1 {
			result.Model.CIDRCursor = 0
		} else {
			result.Model.PortCursor = 0
		}
	case ActionMoveEnd:
		if result.Model.FocusField == 0 {
			result.Model.IPCursor = len(result.Model.IPInput)
		} else if result.Model.FocusField == 1 {
			result.Model.CIDRCursor = len(result.Model.CIDRInput)
		} else {
			result.Model.PortCursor = len(result.Model.CustomPorts)
		}
	case ActionBackspace:
		if result.Model.FocusField == 0 {
			if result.Model.IPCursor > 0 {
				result.Model.IPInput = result.Model.IPInput[:result.Model.IPCursor-1] + result.Model.IPInput[result.Model.IPCursor:]
				result.Model.IPCursor--
			}
		} else if result.Model.FocusField == 1 {
			if result.Model.CIDRCursor > 0 {
				result.Model.CIDRInput = result.Model.CIDRInput[:result.Model.CIDRCursor-1] + result.Model.CIDRInput[result.Model.CIDRCursor:]
				result.Model.CIDRCursor--
			}
		} else {
			if result.Model.PortCursor > 0 {
				result.Model.CustomPorts = result.Model.CustomPorts[:result.Model.PortCursor-1] + result.Model.CustomPorts[result.Model.PortCursor:]
				result.Model.PortCursor--
			}
		}
	case ActionDeleteAll:
		if result.Model.FocusField == 2 {
			result.Model.CustomPorts = ""
			result.Model.PortCursor = 0
		}
	default:
		// Handle character input
		if len(msg.String()) == 1 && msg.Type == tea.KeyRunes {
			char := msg.Runes[0]
			if result.Model.FocusField == 0 {
				// IP field - allow digits and dots
				if isIPChar(char) {
					result.Model.IPInput = insertChar(result.Model.IPInput, result.Model.IPCursor, char)
					result.Model.IPCursor++
				}
			} else if result.Model.FocusField == 1 {
				// CIDR field - allow digits only (0-32)
				if isCIDRChar(char) {
					result.Model.CIDRInput = insertChar(result.Model.CIDRInput, result.Model.CIDRCursor, char)
					result.Model.CIDRCursor++
				}
			} else {
				// Ports field - allow digits, hyphens, commas
				if isPortChar(char) {
					result.Model.CustomPorts = insertChar(result.Model.CustomPorts, result.Model.PortCursor, char)
					result.Model.PortCursor++
				}
			}
		}
	}

	return result
}

func isIPChar(r rune) bool {
	return (r >= '0' && r <= '9') || r == '.'
}

func isCIDRChar(r rune) bool {
	return r >= '0' && r <= '9'
}

func isPortChar(r rune) bool {
	return (r >= '0' && r <= '9') || r == '-' || r == ','
}

func insertChar(s string, pos int, ch rune) string {
	if pos < 0 {
		pos = 0
	}
	if pos > len(s) {
		pos = len(s)
	}
	return s[:pos] + string(ch) + s[pos:]
}

func validateAndBuild(m Model) (mainview.ScanSelection, error) {
	// Validate IP
	ipInput := m.IPInput
	if ipInput == "" {
		return mainview.ScanSelection{}, errors.New("IP address required")
	}

	// Validate CIDR
	cidrStr := m.CIDRInput
	if cidrStr == "" {
		cidrStr = "32" // Default to single host
	}
	cidrVal := 0
	_, err := fmt.Sscanf(cidrStr, "%d", &cidrVal)
	if err != nil || cidrVal < 16 || cidrVal > 32 {
		return mainview.ScanSelection{}, errors.New("CIDR must be 16-32")
	}

	// Build CIDR notation
	cidrNotation := ipInput + "/" + cidrStr

	// Parse CIDR
	_, ipnet, err := net.ParseCIDR(cidrNotation)
	if err != nil {
		return mainview.ScanSelection{}, errors.New("invalid IP address")
	}

	// Compute total hosts
	totalHosts := shared.TotalScanHosts(ipnet)

	// Build selection
	selection := mainview.ScanSelection{
		TargetAddr: cidrNotation,
		TotalHosts: totalHosts,
	}

	return selection, nil
}

// SaveTargetPortsConfig saves the target ports configuration to disk
func SaveTargetPortsConfig(portPack, customPorts string) error {
	cfg := ports.Config{
		Mode:   portPack,
		Custom: customPorts,
	}
	return ports.SaveConfig("target", cfg)
}
