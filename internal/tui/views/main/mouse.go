package mainview

import (
	tea "charm.land/bubbletea/v2"
	"github.com/backendsystems/nibble/internal/tui/views/common"
)

const (
	cardHeight      = 4 // top border + 2 content lines + bottom border
	cardTitleOffset = 1 // title line before cards
)

// CardIndexAt returns the card index at the given terminal (x, y) position
// accounting for the viewport scroll offset, or -1 if the position misses all cards.
func CardIndexAt(x, y, yOffset, cardsPerRow, totalCards, cardWidth int) int {
	if y < cardTitleOffset {
		return -1
	}
	// Convert screen Y to content Y by adding the viewport scroll offset.
	contentY := (y - cardTitleOffset) + yOffset
	row := contentY / cardHeight
	col := x / cardWidth
	if col >= cardsPerRow {
		return -1
	}
	index := row*cardsPerRow + col
	if index >= totalCards {
		return -1
	}
	return index
}

func (m Model) HandleMouse(msg tea.Msg) UpdateResult {
	result := UpdateResult{Model: m}

	mouse, ok := common.MouseXY(msg)
	if !ok {
		return result
	}

	// Help overlay should capture mouse input so underlying cards are not interactive.
	// Match keyboard behavior: any click closes help.
	if m.ShowHelp {
		if _, ok := msg.(tea.MouseReleaseMsg); ok && mouse.Button == tea.MouseLeft {
			result.Model.ShowHelp = false
		}
		return result
	}

	hlr := common.HandleHelpLineMouse(mouse, msg, mainHelpItems, helpPrefixText, m.HelpLineY, m.Viewport.Width())
	result.Model.HoveredHelpItem = hlr.NewHoveredItem

	switch msg.(type) {
	case tea.MouseWheelMsg:
		if mouse.Button == tea.MouseWheelUp {
			result.Model.Viewport.SetYOffset(max(0, result.Model.Viewport.YOffset()-cardHeight))
			return result
		}
		if mouse.Button == tea.MouseWheelDown {
			totalCards := len(m.Interfaces) + 2
			totalRows := (totalCards + m.CardsPerRow - 1) / m.CardsPerRow
			maxOffset := max(0, totalRows*cardHeight-m.Viewport.Height())
			result.Model.Viewport.SetYOffset(min(result.Model.Viewport.YOffset()+cardHeight, maxOffset))
			return result
		}
	}

	if _, ok := msg.(tea.MouseReleaseMsg); !ok || mouse.Button != tea.MouseLeft {
		return result
	}

	if hlr.Consumed {
		switch Action(hlr.ClickedAction) {
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

	totalCards := len(m.Interfaces) + 2
	index := CardIndexAt(mouse.X, mouse.Y, m.Viewport.YOffset(), m.CardsPerRow, totalCards, m.CardWidth)
	if index < 0 {
		return result
	}
	if index == m.Cursor {
		// Second click on already-selected card: activate it
		activateResult := result.Model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
		return activateResult
	}
	result.Model.Cursor = index
	return result
}
