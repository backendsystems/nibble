package common

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// CustomPortInput manages custom port text entry state.
type CustomPortInput struct {
	Input  textinput.Model
	Ready  bool
	Value  string
	Cursor int
}

// Prepare initializes or re-syncs the textinput, optionally focusing it.
func (c CustomPortInput) Prepare(focused bool) (CustomPortInput, tea.Cmd) {
	if !c.Ready {
		c.Input = NewCustomPortsInput()
		c.Ready = true
		if c.Cursor == 0 && len(c.Value) > 0 {
			c.Cursor = len(c.Value)
		}
	}
	c.Input.SetValue(c.Value)
	c.Cursor = ClampCursor(c.Cursor, len(c.Value))
	c.Input.SetCursor(c.Cursor)

	if focused {
		cmd := c.Input.Focus()
		return c.syncFromInput(), cmd
	}
	c.Input.Blur()
	return c.syncFromInput(), nil
}

// syncFromInput syncs textinput state back to Value and Cursor.
func (c CustomPortInput) syncFromInput() CustomPortInput {
	if !c.Ready {
		return c
	}
	c.Value = c.Input.Value()
	c.Cursor = c.Input.Position()
	return c
}

// UpdateNonKey forwards non-keyboard messages to the textinput.
func (c CustomPortInput) UpdateNonKey(msg tea.Msg) (CustomPortInput, tea.Cmd) {
	var cmd tea.Cmd
	c.Input, cmd = c.Input.Update(msg)
	return c.syncFromInput(), cmd
}

// HandleKey dispatches key actions (move, delete, insert) or falls through to textinput.
func (c CustomPortInput) HandleKey(action PortInputAction, keyMsg tea.KeyMsg) (CustomPortInput, tea.Cmd) {
	switch {
	case action.MoveLeft:
		if pos := c.Input.Position(); pos > 0 {
			c.Input.SetCursor(MoveCursorLeft(pos))
		}
		return c.syncFromInput(), nil
	case action.MoveRight:
		pos, vlen := c.Input.Position(), len(c.Input.Value())
		if pos < vlen {
			c.Input.SetCursor(MoveCursorRight(pos, vlen))
		}
		return c.syncFromInput(), nil
	case action.MoveHome:
		c.Input.CursorStart()
		return c.syncFromInput(), nil
	case action.MoveEnd:
		c.Input.CursorEnd()
		return c.syncFromInput(), nil
	case action.Backspace:
		val, cur := c.Input.Value(), c.Input.Position()
		if cur > 0 && len(val) > 0 {
			val, cur = Backspace(val, cur)
			c.Input.SetValue(val)
			c.Input.SetCursor(cur)
		}
		return c.syncFromInput(), nil
	case action.DeleteAll:
		c.Input.SetValue("")
		c.Input.SetCursor(0)
		return c.syncFromInput(), nil
	case action.InsertRunes:
		filtered := filterPrintable(keyMsg.Runes)
		if len(filtered) > 0 {
			val, cur := c.Input.Value(), c.Input.Position()
			val, cur = InsertRunes(val, cur, filtered)
			c.Input.SetValue(val)
			c.Input.SetCursor(cur)
		}
		return c.syncFromInput(), nil
	}
	var cmd tea.Cmd
	c.Input, cmd = c.Input.Update(keyMsg)
	return c.syncFromInput(), cmd
}

// SetValue replaces value and moves cursor to end.
func (c CustomPortInput) SetValue(v string) CustomPortInput {
	c.Value = v
	c.Cursor = len(v)
	if c.Ready {
		c.Input.SetValue(v)
		c.Input.SetCursor(len(v))
	}
	return c
}

// PortInputAction describes which editing operation a key maps to.
type PortInputAction struct {
	MoveLeft    bool
	MoveRight   bool
	MoveHome    bool
	MoveEnd     bool
	Backspace   bool
	DeleteAll   bool
	InsertRunes bool
}

// PortInputActionFromKey maps a key string to an action.
func PortInputActionFromKey(key string, isRunes bool) PortInputAction {
	switch key {
	case "left", "a", "h":
		return PortInputAction{MoveLeft: true}
	case "right", "d", "l":
		return PortInputAction{MoveRight: true}
	case "home", "ctrl+a":
		return PortInputAction{MoveHome: true}
	case "end", "ctrl+e":
		return PortInputAction{MoveEnd: true}
	case "backspace":
		return PortInputAction{Backspace: true}
	case "delete":
		return PortInputAction{DeleteAll: true}
	}
	if isRunes {
		return PortInputAction{InsertRunes: true}
	}
	return PortInputAction{}
}

func filterPrintable(runes []rune) []rune {
	out := make([]rune, 0, len(runes))
	for _, r := range runes {
		if r >= 32 {
			out = append(out, r)
		}
	}
	return out
}
