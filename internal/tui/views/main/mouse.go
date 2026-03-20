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

func (m Model) HandleMouse(msg tea.Msg) UpdateResult {
	result := UpdateResult{Model: m}

	mouse, ok := mouseXY(msg)
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

	helpLineY := m.HelpLineY
	helpLayout := common.BuildHelpLineLayout(mainHelpItems, helpPrefixText, m.Viewport.Width())
	helpLineEndY := helpLineY + helpLayout.LineCount - 1

	// Handle hover for helpline items (update hover state for all mouse events)
	if helpLineY > 0 && mouse.Y >= helpLineY && mouse.Y <= helpLineEndY {
		result.Model.HoveredHelpItem = common.GetHelpItemAt(helpLayout, mouse.X, mouse.Y-helpLineY)
	} else {
		result.Model.HoveredHelpItem = -1
	}

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

	// Check if clicking on helpline item
	if helpLineY > 0 && mouse.Y >= helpLineY && mouse.Y <= helpLineEndY {
		itemIndex := common.GetHelpItemAt(helpLayout, mouse.X, mouse.Y-helpLineY)
		if itemIndex >= 0 {
			switch Action(helpLayout.Items[itemIndex].Action) {
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
	index := CardIndexAt(mouse.X, mouse.Y, m.Viewport.YOffset(), m.CardsPerRow, totalCards)
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

// mouseXY extracts Mouse data from any mouse message type.
func mouseXY(msg tea.Msg) (tea.Mouse, bool) {
	if mm, ok := msg.(tea.MouseMsg); ok {
		return mm.Mouse(), true
	}
	return tea.Mouse{}, false
}
