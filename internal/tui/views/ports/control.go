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

func Prepare(m Model) (Model, tea.Cmd) {
	if !m.InputReady {
		m.CustomInput = common.NewCustomPortsInput()
		m.InputReady = true
		if len(m.CustomPorts) > 0 {
			m.CustomCursor = len(m.CustomPorts)
		}
	}

	m.CustomInput.SetValue(m.CustomPorts)
	m.CustomCursor = common.ClampCursor(m.CustomCursor, len(m.CustomPorts))
	m.CustomInput.SetCursor(m.CustomCursor)

	if m.PortPack == "custom" {
		cmd := m.CustomInput.Focus()
		return syncStateFromInput(m), cmd
	}

	m.CustomInput.Blur()
	return syncStateFromInput(m), nil
}

func syncStateFromInput(m Model) Model {
	if !m.InputReady {
		return m
	}
	m.CustomPorts = m.CustomInput.Value()
	m.CustomCursor = m.CustomInput.Position()
	return m
}

func applyConfig(m Model) (Model, bool) {
	customPorts := ""
	if m.PortPack == "custom" {
		m = syncStateFromInput(m)
		normalized, err := ports.NormalizeCustom(strings.TrimSpace(m.CustomPorts))
		if err != nil {
			m.ErrorMsg = err.Error()
			return m, false
		}
		customPorts = normalized
		m.CustomPorts = normalized
		m.CustomCursor = len(normalized)
		if m.InputReady {
			m.CustomInput.SetValue(normalized)
			m.CustomInput.SetCursor(len(normalized))
		}
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
		if result.Model.PortPack == "custom" && result.Model.InputReady {
			var cmd tea.Cmd
			result.Model.CustomInput, cmd = result.Model.CustomInput.Update(msg)
			result.Model = syncStateFromInput(result.Model)
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

	if !result.Model.InputReady {
		var cmd tea.Cmd
		result.Model, cmd = Prepare(result.Model)
		result.Cmd = cmd
	}

	if action.MoveLeft {
		if pos := result.Model.CustomInput.Position(); pos > 0 {
			result.Model.CustomInput.SetCursor(common.MoveCursorLeft(pos))
			result.Model = syncStateFromInput(result.Model)
		}
		return result
	}
	if action.MoveRight {
		pos := result.Model.CustomInput.Position()
		valueLen := len(result.Model.CustomInput.Value())
		if pos < valueLen {
			result.Model.CustomInput.SetCursor(common.MoveCursorRight(pos, valueLen))
			result.Model = syncStateFromInput(result.Model)
		}
		return result
	}
	if action.MoveHome {
		result.Model.CustomInput.CursorStart()
		result.Model = syncStateFromInput(result.Model)
		return result
	}
	if action.MoveEnd {
		result.Model.CustomInput.CursorEnd()
		result.Model = syncStateFromInput(result.Model)
		return result
	}
	if action.Backspace {
		value := result.Model.CustomInput.Value()
		cursor := result.Model.CustomInput.Position()
		if cursor > 0 && len(value) > 0 {
			value, cursor = common.Backspace(value, cursor)
			result.Model.CustomInput.SetValue(value)
			result.Model.CustomInput.SetCursor(cursor)
			result.Model = syncStateFromInput(result.Model)
		}
		return result
	}
	if action.DeleteAll {
		result.Model.CustomInput.SetValue("")
		result.Model.CustomInput.SetCursor(0)
		result.Model = syncStateFromInput(result.Model)
		return result
	}

	if keyMsg.Type == tea.KeyRunes {
		filtered := make([]rune, 0, len(keyMsg.Runes))
		for _, r := range keyMsg.Runes {
			if r >= 32 {
				filtered = append(filtered, r)
			}
		}
		if len(filtered) > 0 {
			value := result.Model.CustomInput.Value()
			cursor := result.Model.CustomInput.Position()
			value, cursor = common.InsertRunes(value, cursor, filtered)
			result.Model.CustomInput.SetValue(value)
			result.Model.CustomInput.SetCursor(cursor)
			result.Model = syncStateFromInput(result.Model)
		}
		return result
	}

	// Keep textinput internal cursor behavior in sync for remaining key messages.
	var cmd tea.Cmd
	result.Model.CustomInput, cmd = result.Model.CustomInput.Update(keyMsg)
	result.Model = syncStateFromInput(result.Model)
	if result.Cmd == nil {
		result.Cmd = cmd
	}
	return result
}
