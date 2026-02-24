package targetview

import (
	"github.com/backendsystems/nibble/internal/tui/views/common"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

// handleKeyPress processes keyboard input for the target form
// Returns a Result and optional tea.Cmd
func handleKeyPress(m *Model, keyMsg tea.KeyMsg) (Result, tea.Cmd) {
	result := Result{}

	// Help screen toggle
	if m.ShowHelp {
		m.ShowHelp = false
		return result, nil
	}

	switch keyMsg.String() {
	case "q", "esc":
		// If in custom_ports field, convert to shift+tab to navigate back
		if m.Form != nil {
			focused := m.Form.GetFocusedField()
			if focused != nil && focused.GetKey() == "custom_ports" {
				// Convert to shift+tab and let form handle it
				return result, nil
			}
		}
		result.Quit = true
		return result, nil

	case "?":
		m.ShowHelp = true
		return result, nil

	case "delete":
		return handleDelete(m)

	case "left", "right":
		return handleLeftRight(m, keyMsg.String() == "right")

	case "up", "k", "w":
		return handleUpNavigation(m, keyMsg)

	case "down", "j", "s":
		return handleDownNavigation(m, keyMsg)

	default:
		return handleCharacterInput(m, keyMsg)
	}
}

// handleDelete clears the focused field when delete key is pressed
func handleDelete(m *Model) (Result, tea.Cmd) {
	result := Result{}
	if m.Form == nil {
		return result, nil
	}
	focused := m.Form.GetFocusedField()
	if focused == nil {
		return result, nil
	}

	switch focused.GetKey() {
	case "ip":
		m.IPInput = ""
	case "cidr":
		m.CIDRInput = ""
	case "custom_ports":
		m.CustomPorts = ""
	default:
		return result, nil
	}

	// Rebuild form so the focused input reflects the cleared value
	m.initializeForm()
	return result, m.Form.Init()
}

// handleLeftRight cycles through interface IPs when left/right keys are pressed
func handleLeftRight(m *Model, forward bool) (Result, tea.Cmd) {
	result := Result{}
	if m.Form != nil {
		focused := m.Form.GetFocusedField()
		if focused != nil && focused.GetKey() == "ip" {
			m.CycleInterfaceIP(forward)
			// Recreate the form with the new IP value
			m.initializeForm()
			return result, m.Form.Init()
		}
	}
	// For other fields, return empty result to let form handle it normally
	return result, nil
}

// handleUpNavigation processes up/k/w keys for form navigation
func handleUpNavigation(m *Model, keyMsg tea.KeyMsg) (Result, tea.Cmd) {
	result := Result{}
	if m.Form == nil {
		return result, nil
	}

	focused := m.Form.GetFocusedField()
	// Block navigation in custom_ports field
	if focused != nil && focused.GetKey() == "custom_ports" {
		return result, nil
	}

	// Wrap from the first field back to ports selection instead of exiting the form
	if focused != nil && focused.GetKey() == "ip" {
		m.Form.NextField() // ip -> cidr
		m.Form.NextField() // cidr -> port_mode
		return result, nil
	}

	// For port_mode select: if at first option (index 0), navigate up to CIDR
	if focused != nil && focused.GetKey() == "port_mode" {
		// Type assert to Select to access Hovered method
		if selectField, ok := focused.(*huh.Select[string]); ok {
			hovered, _ := selectField.Hovered()
			// Check if we're at the first option ("default")
			if hovered == "default" {
				// At first option, navigate to previous field
				m.Form.PrevField()
				return result, nil
			}
		}
		// Not at first option, convert k/w to up and let select handle it
		if keyMsg.String() == "k" || keyMsg.String() == "w" {
			// Return empty result with no cmd - the caller will pass KeyUp to form
			return result, nil
		}
		return result, nil
	}

	m.Form.PrevField()
	return result, nil
}

// handleDownNavigation processes down/j/s keys for form navigation
func handleDownNavigation(m *Model, keyMsg tea.KeyMsg) (Result, tea.Cmd) {
	result := Result{}
	if m.Form == nil {
		return result, nil
	}

	focused := m.Form.GetFocusedField()
	// Block navigation in custom_ports field
	if focused != nil && focused.GetKey() == "custom_ports" {
		return result, nil
	}

	// For port_mode select: convert j/s to down and let it handle navigation
	if focused != nil && focused.GetKey() == "port_mode" {
		if keyMsg.String() == "j" || keyMsg.String() == "s" {
			// Return empty result - caller will convert to KeyDown
			return result, nil
		}
		return result, nil
	}

	m.Form.NextField()
	return result, nil
}

// handleCharacterInput validates and filters character input based on focused field
func handleCharacterInput(m *Model, keyMsg tea.KeyMsg) (Result, tea.Cmd) {
	result := Result{}
	if m.Form == nil {
		return result, nil
	}

	focused := m.Form.GetFocusedField()
	if focused == nil {
		return result, nil
	}

	key := focused.GetKey()
	if len(keyMsg.String()) == 1 {
		ch := keyMsg.String()[0]

		// Validate single character input
		if !validateInputChar(key, ch) {
			return result, nil
		}
	}

	// Custom ports field: use portinput validation for rune sequences
	if key == "custom_ports" && keyMsg.Type == tea.KeyRunes {
		// Get current value from the input field
		currentValue := m.Form.GetString("custom_ports")

		// Filter runes through portinput
		filtered := make([]rune, 0, len(keyMsg.Runes))
		for _, r := range keyMsg.Runes {
			if r >= 32 {
				filtered = append(filtered, r)
			}
		}

		if len(filtered) > 0 {
			// Insert runes at the end (huh doesn't expose cursor position)
			newValue, _ := common.InsertRunes(currentValue, len(currentValue), filtered)

			// If the value changed, it means valid characters were added
			// Otherwise, block the input
			if newValue == currentValue {
				return result, nil
			}
		}
	}

	return result, nil
}
