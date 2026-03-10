package mainview

import (
	tea "github.com/charmbracelet/bubbletea"
)

const (
	cardHeight      = 4 // top border + 2 content lines + bottom border
	cardTitleOffset = 1 // title line before cards
)

// CardIndexAt returns the card index at the given terminal (x, y) position,
// or -1 if the position doesn't land on a card.
func CardIndexAt(x, y, cardsPerRow, totalCards int) int {
	if y < cardTitleOffset {
		return -1
	}
	row := (y - cardTitleOffset) / cardHeight
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
	if msg.Button != tea.MouseButtonLeft || msg.Action != tea.MouseActionRelease {
		return result
	}
	totalCards := len(m.Interfaces) + 2
	index := CardIndexAt(msg.X, msg.Y, m.CardsPerRow, totalCards)
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
