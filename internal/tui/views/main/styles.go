package mainview

import (
	"fmt"
	"net"

	"charm.land/lipgloss/v2"
	"github.com/backendsystems/nibble/internal/tui/views/common"
)

const (
	minCardWidth = 20
	cardPaddingX = 1
	// border (2) + padding (2*cardPaddingX) consumed by lipgloss Width
	cardChrome = 2 + 2*cardPaddingX
)

// ComputeCardWidth returns the card width needed to fit all interface labels
// without truncation. The returned value is a lipgloss Width (includes
// border + padding).
func ComputeCardWidth(ifaces []net.Interface, addrsByIface map[string][]net.Addr) int {
	longest := 0
	for _, iface := range ifaces {
		for _, addr := range addrsByIface[iface.Name] {
			if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
				ones, _ := ipnet.Mask.Size()
				label := fmt.Sprintf("%s/%d", ipnet.IP.String(), ones)
				if len(label) > longest {
					longest = len(label)
				}
			}
		}
	}
	width := longest + cardChrome
	if width < minCardWidth {
		width = minCardWidth
	}
	return width
}

func CardsPerRow(windowWidth, cardWidth int) int {
	cardsPerRow := windowWidth / cardWidth
	if cardsPerRow < 1 {
		return 1
	}
	return cardsPerRow
}

func cardStyle(width int) lipgloss.Style {
	return common.CardStyle.Padding(0, cardPaddingX).Width(width)
}

func selectedCardStyle(width int) lipgloss.Style {
	return common.SelectedCardStyle.Padding(0, cardPaddingX).Width(width)
}
