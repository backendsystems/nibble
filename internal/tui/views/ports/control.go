package portsview

import (
	"strings"

	"github.com/backendsystems/nibble/internal/ports"
	"github.com/backendsystems/nibble/internal/scanner/demo"
	"github.com/backendsystems/nibble/internal/scanner/ip4"
	"github.com/backendsystems/nibble/internal/tui/views/common"

	tea "github.com/charmbracelet/bubbletea"
)

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
	Cmd   tea.Cmd
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
	case "q", "esc":
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

func Prepare(m Model) (Model, tea.Cmd) {
	// Seed PortInput from the model's persistent fields on first call.
	if !m.PortInput.Ready {
		m.PortInput.Value = m.CustomPorts
		m.PortInput.Cursor = m.CustomCursor
	}
	focused := m.PortPack == "custom"
	var cmd tea.Cmd
	m.PortInput, cmd = m.PortInput.Prepare(focused)
	// Keep top-level fields in sync for persistence layer.
	m.CustomPorts = m.PortInput.Value
	m.CustomCursor = m.PortInput.Cursor
	return m, cmd
}

func applyConfig(m Model) (Model, bool) {
	customPorts := ""
	if m.PortPack == "custom" {
		normalized, err := ports.NormalizeCustom(strings.TrimSpace(m.PortInput.Value))
		if err != nil {
			m.ErrorMsg = err.Error()
			return m, false
		}
		customPorts = normalized
		m.PortInput = m.PortInput.SetValue(normalized)
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

func (m Model) Update(msg tea.Msg) Result {
	result := Result{Model: m}

	// Let the textinput cursor blink/update on non-key messages.
	if _, ok := msg.(tea.KeyMsg); !ok {
		if result.Model.PortPack == "custom" && result.Model.PortInput.Ready {
			var cmd tea.Cmd
			result.Model.PortInput, cmd = result.Model.PortInput.UpdateNonKey(msg)
			result.Model.CustomPorts = result.Model.PortInput.Value
			result.Model.CustomCursor = result.Model.PortInput.Cursor
			result.Cmd = cmd
		}
		return result
	}

	keyMsg := msg.(tea.KeyMsg)
	action := HandleKey(result.Model.ShowHelp, keyMsg.String())
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
		var cmd tea.Cmd
		result.Model, cmd = Prepare(result.Model)
		result.Cmd = cmd
		return result
	}
	if action.Apply {
		next, ok := applyConfig(result.Model)
		result.Model = next
		result.Done = ok
		return result
	}

	if result.Model.PortPack != "custom" {
		return result
	}

	if !result.Model.PortInput.Ready {
		var cmd tea.Cmd
		result.Model, cmd = Prepare(result.Model)
		result.Cmd = cmd
	}

	// Delegate all editing operations to PortInput.
	portAction := common.PortInputActionFromKey(keyMsg.String(), keyMsg.Type == tea.KeyRunes)
	var cmd tea.Cmd
	result.Model.PortInput, cmd = result.Model.PortInput.HandleKey(portAction, keyMsg)
	result.Model.CustomPorts = result.Model.PortInput.Value
	result.Model.CustomCursor = result.Model.PortInput.Cursor
	result.Cmd = cmd
	return result
}
