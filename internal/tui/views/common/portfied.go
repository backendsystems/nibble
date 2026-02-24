package common

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

const CustomPortsDescription = "e.g. 22,80,443,8000-9000 (valid: 1-65535)"

func NewCustomPortsInput() textinput.Model {
	input := textinput.New()
	input.Prompt = "> "
	input.PromptStyle = lipgloss.NewStyle().Foreground(Color.Selection)
	input.Cursor.Style = lipgloss.NewStyle().Foreground(Color.Selection)
	return input
}

// ClampCursor ensures cursor position is within valid bounds [0, valueLen].
func ClampCursor(cursor, valueLen int) int {
	if cursor < 0 {
		return 0
	}
	if cursor > valueLen {
		return valueLen
	}
	return cursor
}

// MoveCursorLeft moves cursor one position left, clamped to 0.
func MoveCursorLeft(cursor int) int {
	if cursor > 0 {
		cursor--
	}
	return cursor
}

// MoveCursorRight moves cursor one position right, clamped to valueLen.
func MoveCursorRight(cursor, valueLen int) int {
	if cursor < valueLen {
		cursor++
	}
	return cursor
}

// Backspace removes the character before the cursor.
func Backspace(value string, cursor int) (string, int) {
	if cursor <= 0 || len(value) == 0 {
		return value, ClampCursor(cursor, len(value))
	}
	i := cursor - 1
	return value[:i] + value[cursor:], i
}

// InsertRunes inserts runes at cursor position, filtering to valid port characters.
// Valid characters: digits (0-9), comma (,), dash (-).
// Enforces validation rules per token (max 5 digits, single dash).
func InsertRunes(value string, cursor int, runes []rune) (string, int) {
	cursor = ClampCursor(cursor, len(value))
	for _, r := range runes {
		if (r >= '0' && r <= '9') || r == '-' {
			if !canInsertPortChar(value, cursor, r) {
				continue
			}
			s := string(r)
			value = value[:cursor] + s + value[cursor:]
			cursor++
			continue
		}
		if r == ',' {
			s := string(r)
			value = value[:cursor] + s + value[cursor:]
			cursor++
		}
	}
	return value, cursor
}

// canInsertPortChar checks if a port character can be inserted at cursor position.
// Validates: max one dash per token, max 5 digits per token part.
func canInsertPortChar(s string, cursor int, ch rune) bool {
	start, end := currentTokenBounds(s, cursor)
	pos := cursor - start
	token := s[start:end]
	next := token[:pos] + string(ch) + token[pos:]

	// Don't allow multiple dashes in a token.
	if strings.Count(next, "-") > 1 {
		return false
	}

	// If no dash, just check token length (max 5 digits for a port number).
	if strings.Count(next, "-") == 0 {
		return len(next) <= 5
	}

	// If there's a dash, check both parts are <= 5 digits.
	parts := strings.SplitN(next, "-", 2)
	return len(parts[0]) <= 5 && len(parts[1]) <= 5
}

// currentTokenBounds finds the start and end indices of the current comma-separated token.
func currentTokenBounds(s string, cursor int) (int, int) {
	cursor = ClampCursor(cursor, len(s))

	// Find start by scanning backward for comma.
	start := -1
	for i := cursor - 1; i >= 0; i-- {
		if s[i] == ',' {
			start = i
			break
		}
	}
	if start == -1 {
		start = 0
	} else {
		start++ // Move past the comma.
	}

	// Find end by scanning forward for comma.
	end := len(s)
	for i := cursor; i < len(s); i++ {
		if s[i] == ',' {
			end = i
			break
		}
	}

	return start, end
}
