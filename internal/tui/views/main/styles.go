package mainview

import (
	"github.com/backendsystems/nibble/internal/tui/views/common"
)

const (
	cardWidth      = 20
	cardPaddingX   = 1
	cardTotalWidth = cardWidth // lipgloss v2 Width includes padding+border
)

func CardsPerRow(windowWidth int) int {
	cardsPerRow := windowWidth / cardTotalWidth
	if cardsPerRow < 1 {
		return 1
	}
	return cardsPerRow
}

var (
	cardStyle         = common.CardStyle.Padding(0, cardPaddingX).Width(cardWidth)
	selectedCardStyle = common.SelectedCardStyle.Padding(0, cardPaddingX).Width(cardWidth)
)
