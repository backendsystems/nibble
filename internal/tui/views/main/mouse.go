package mainview

import (
	tea "github.com/charmbracelet/bubbletea"
)

const (
	cardHeight      = 4 // top border + 2 content lines + bottom border
	cardTitleOffset = 1 // title line before cards
)

// CardIndexAt returns the card index at the given terminal (x, y) position
// accounting for the viewport scroll offset, or -1 if the position misses all cards.
func CardIndexAt(x, y, yOffset, cardsPerRow, totalCards int) int {
	if y < cardTitleOffset {
		return -1
	}
	// Convert screen Y to content Y by adding the viewport scroll offset.
	contentY := (y - cardTitleOffset) + yOffset
	row := contentY / cardHeight
	col := x / cardTotalWidth
	if col >= cardsPerRow {
		return -1
	}
	index := row*cardsPerRow + col
	if index >= totalCards {
		return -1
	}
	return index
}

func (m Model) HandleMouse(msg tea.MouseMsg) UpdateResult {
	result := UpdateResult{Model: m}

	// Help overlay should capture mouse input so underlying cards are not interactive.
	// Match keyboard behavior: any click closes help.
	if m.ShowHelp {
		if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionRelease {
			result.Model.ShowHelp = false
		}
		return result
	}

	// Calculate helpline Y position:
	// Y=0: title
	// Y=1 to Y=vpHeight: viewport (vpHeight lines)
	// Y=(1+vpHeight): helpline (no error) OR blank line (with error)
	// Y=(1+vpHeight+1): error line (with error)
	// Y=(1+vpHeight+2): helpline (with error)
	helpLineY := 1 + m.Viewport.Height
	if m.ErrorMsg != "" {
		helpLineY += 2 // +2 for "\n\n" + error, then helpline is next line
	}
	helpLayout := BuildHelpLineLayout(m.Viewport.Width)
	helpLineEndY := helpLineY + helpLayout.LineCount - 1

	// Handle hover for helpline items (update hover state for all mouse events)
	if msg.Y >= helpLineY && msg.Y <= helpLineEndY {
		result.Model.HoveredHelpItem = GetHelpItemAt(msg.X, msg.Y-helpLineY, m.Viewport.Width)
	} else {
		result.Model.HoveredHelpItem = -1
	}

	switch msg.Button {
	case tea.MouseButtonWheelUp:
		result.Model.Viewport.YOffset = max(0, result.Model.Viewport.YOffset-cardHeight)
		return result
	case tea.MouseButtonWheelDown:
		totalCards := len(m.Interfaces) + 2
		totalRows := (totalCards + m.CardsPerRow - 1) / m.CardsPerRow
		maxOffset := max(0, totalRows*cardHeight-m.Viewport.Height)
		result.Model.Viewport.YOffset = min(result.Model.Viewport.YOffset+cardHeight, maxOffset)
		return result
	}

	if msg.Button != tea.MouseButtonLeft || msg.Action != tea.MouseActionRelease {
		return result
	}

	// Check if clicking on helpline item
	if msg.Y >= helpLineY && msg.Y <= helpLineEndY {
		itemIndex := GetHelpItemAt(msg.X, msg.Y-helpLineY, m.Viewport.Width)
		if itemIndex >= 0 {
			action := GetHelpItemAction(itemIndex, m.Viewport.Width)
			switch action {
			case ActionOpenPorts:
				result.OpenPorts = true
			case ActionOpenHistory:
				result.OpenHistory = true
			case ActionOpenTarget:
				result.OpenTarget = true
			case ActionOpenHelp:
				result.Model.ShowHelp = true
			case ActionQuit:
				result.Quit = true
			}
			return result
		}
	}

	totalCards := len(m.Interfaces) + 2
	index := CardIndexAt(msg.X, msg.Y, m.Viewport.YOffset, m.CardsPerRow, totalCards)
	if index < 0 {
		return result
	}
	if index == m.Cursor {
		// Second click on already-selected card: activate it
		activateResult := result.Model.Update(tea.KeyMsg{Type: tea.KeyEnter})
		return activateResult
	}
	result.Model.Cursor = index
	return result
}
