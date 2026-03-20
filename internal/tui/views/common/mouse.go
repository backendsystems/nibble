package common

import tea "charm.land/bubbletea/v2"

// MouseXY extracts Mouse data from any mouse message type.
func MouseXY(msg tea.Msg) (tea.Mouse, bool) {
	if mm, ok := msg.(tea.MouseMsg); ok {
		return mm.Mouse(), true
	}
	return tea.Mouse{}, false
}

// IsMouseMsg reports whether msg is any type of mouse message.
func IsMouseMsg(msg tea.Msg) bool {
	_, ok := msg.(tea.MouseMsg)
	return ok
}
