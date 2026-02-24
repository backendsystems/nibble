package portsview

import (
	"strings"

	"github.com/backendsystems/nibble/internal/ports"
	"github.com/backendsystems/nibble/internal/scanner/demo"
	"github.com/backendsystems/nibble/internal/scanner/ip4"
	"github.com/backendsystems/nibble/internal/tui/views/common/portinput"

	tea "github.com/charmbracelet/bubbletea"
)

const portsGuideText = "  • enter ports e.g. 22,80,443,8000-9000 or empty. valid values are 1-65535"
const portsHelpText = "tab • backspace • delete: clear all • enter • ?: help • q: cancel"

type Action struct {
	Handled    bool
	Quit       bool
	Back       bool
	CloseHelp  bool
	OpenHelp   bool
	ToggleMode bool
	Apply      bool
	MoveLeft   bool
	MoveRight  bool
	MoveHome   bool
	MoveEnd    bool
	Backspace  bool
	DeleteAll  bool
}

type Result struct {
	Model Model
	Quit  bool
	Back  bool
	Done  bool
}

func HandleKey(showHelp bool, key string) Action {
	if showHelp {
		return Action{Handled: true, CloseHelp: true}
	}

	switch key {
	case "ctrl+c":
		return Action{Handled: true, Quit: true}
	case "q":
		return Action{Handled: true, Back: true}
	case "?":
		return Action{Handled: true, OpenHelp: true}
	case "tab", "up", "down", "w", "s", "k", "j":
		return Action{Handled: true, ToggleMode: true}
	case "enter":
		return Action{Handled: true, Apply: true}
	case "left", "a", "h":
		return Action{Handled: true, MoveLeft: true}
	case "right", "d", "l":
		return Action{Handled: true, MoveRight: true}
	case "home", "ctrl+a":
		return Action{Handled: true, MoveHome: true}
	case "end", "ctrl+e":
		return Action{Handled: true, MoveEnd: true}
	case "backspace":
		return Action{Handled: true, Backspace: true}
	case "delete":
		return Action{Handled: true, DeleteAll: true}
	default:
		return Action{}
	}
}

func ToggleMode(portPack string) string {
	if portPack == "default" {
		return "custom"
	}
	return "default"
}


func applyConfig(m Model) (Model, bool) {
	customPorts := ""
	if m.PortPack == "custom" {
		normalized, err := ports.NormalizeCustom(strings.TrimSpace(m.CustomPorts))
		if err != nil {
			m.ErrorMsg = err.Error()
			return m, false
		}
		customPorts = normalized
		m.CustomPorts = normalized
		m.CustomCursor = len(normalized)
	}

	// Build ports string based on mode
	portStr := customPorts
	if m.PortPack == "all" {
		portStr = "1-65535"
	}

	resolvedPorts, err := ports.ParseList(portStr)
	if err != nil {
		m.ErrorMsg = err.Error()
		return m, false
	}
	if err := ports.SaveConfig("ports", ports.Config{Mode: m.PortPack, Custom: m.CustomPorts}); err != nil {
		m.ErrorMsg = err.Error()
		return m, false
	}

	// In custom mode with empty ports, use empty slice for host-only scan
	if m.PortPack == "custom" && resolvedPorts == nil {
		resolvedPorts = []int{}
	}

	switch typed := m.NetworkScan.(type) {
	case *ip4.Scanner:
		typed.Ports = resolvedPorts
	case *demo.Scanner:
		typed.Ports = resolvedPorts
	}
	m.ErrorMsg = ""
	return m, true
}

func (m Model) Update(msg tea.KeyMsg) Result {
	result := Result{Model: m}
	action := HandleKey(m.ShowHelp, msg.String())
	if action.Quit {
		result.Quit = true
		return result
	}
	if action.Back {
		result.Back = true
		return result
	}
	if action.CloseHelp {
		result.Model.ShowHelp = false
		return result
	}
	if action.OpenHelp {
		result.Model.ShowHelp = true
		return result
	}
	if action.ToggleMode {
		result.Model.PortPack = ToggleMode(result.Model.PortPack)
		result.Model.CustomCursor = portinput.ClampCursor(result.Model.CustomCursor, len(result.Model.CustomPorts))
		return result
	}
	if action.Apply {
		next, ok := applyConfig(result.Model)
		result.Model = next
		result.Done = ok
		return result
	}
	if action.MoveLeft {
		if result.Model.PortPack == "custom" && result.Model.CustomCursor > 0 {
			result.Model.CustomCursor = portinput.MoveCursorLeft(result.Model.CustomCursor)
		}
		return result
	}
	if action.MoveRight {
		if result.Model.PortPack == "custom" && result.Model.CustomCursor < len(result.Model.CustomPorts) {
			result.Model.CustomCursor = portinput.MoveCursorRight(result.Model.CustomCursor, len(result.Model.CustomPorts))
		}
		return result
	}
	if action.MoveHome {
		if result.Model.PortPack == "custom" {
			result.Model.CustomCursor = 0
		}
		return result
	}
	if action.MoveEnd {
		if result.Model.PortPack == "custom" {
			result.Model.CustomCursor = len(result.Model.CustomPorts)
		}
		return result
	}
	if action.Backspace {
		if result.Model.PortPack == "custom" && result.Model.CustomCursor > 0 && len(result.Model.CustomPorts) > 0 {
			result.Model.CustomPorts, result.Model.CustomCursor = portinput.Backspace(result.Model.CustomPorts, result.Model.CustomCursor)
		}
		return result
	}
	if action.DeleteAll {
		if result.Model.PortPack == "custom" {
			result.Model.CustomPorts = ""
			result.Model.CustomCursor = 0
		}
		return result
	}
	if result.Model.PortPack == "custom" && msg.Type == tea.KeyRunes {
		// Filter out control characters
		filtered := make([]rune, 0, len(msg.Runes))
		for _, r := range msg.Runes {
			if r >= 32 {
				filtered = append(filtered, r)
			}
		}
		if len(filtered) > 0 {
			result.Model.CustomPorts, result.Model.CustomCursor = portinput.InsertRunes(result.Model.CustomPorts, result.Model.CustomCursor, filtered)
		}
	}
	return result
}
