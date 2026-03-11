package mainview

import (
	"fmt"
	"net"
	"strings"

	"github.com/backendsystems/nibble/internal/tui/views/common"
	"github.com/charmbracelet/lipgloss"
)

func Render(m Model, maxWidth int) string {
	cardsPerRow := m.CardsPerRow
	if cardsPerRow == 0 {
		cardsPerRow = 1
	}

	icons := make(map[string]string, len(m.Interfaces))
	for _, iface := range m.Interfaces {
		icons[iface.Name] = interfaceIcon(iface.Name)
	}

	var rows []string
	var currentRow []string

	for i, iface := range m.Interfaces {
		card := renderInterfaceCard(m, icons, i, iface)
		currentRow = append(currentRow, card)
		if len(currentRow) == cardsPerRow {
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, currentRow...))
			currentRow = nil
		}
	}

	targetCardIndex := len(m.Interfaces)
	targetCard := renderTargetCard(m, targetCardIndex)
	currentRow = append(currentRow, targetCard)
	if len(currentRow) == cardsPerRow {
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, currentRow...))
		currentRow = nil
	}

	historyCardIndex := len(m.Interfaces) + 1
	historyCard := renderHistoryCard(m, historyCardIndex)
	currentRow = append(currentRow, historyCard)
	if len(currentRow) > 0 {
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, currentRow...))
	}

	cardContent := lipgloss.JoinVertical(lipgloss.Left, rows...)

	// UpdateViewport must have already been called by the controller so YOffset
	// is current. Here we just refresh the content and dimensions before rendering.
	m.Viewport.SetContent(cardContent)

	var b strings.Builder
	b.WriteString(common.TitleStyle.Render("Nibble Network Scanner") + "\n")
	b.WriteString(m.Viewport.View())

	if m.ErrorMsg != "" {
		b.WriteString("\n\n" + common.ErrorStyle.Render("Error: "+m.ErrorMsg))
	}
	b.WriteString("\n" + RenderHelpLine(maxWidth, m.HoveredHelpItem))

	view := b.String()
	if m.ShowHelp {
		return renderHelpOverlay(view, maxWidth)
	}
	return view
}

// UpdateViewport refreshes viewport dimensions. Call after every model change
// so Viewport.Width/Height are current for mouse hit-testing.
func (m Model) UpdateViewport(maxWidth int) Model {
	reserved := 4
	vpHeight := m.WindowH - reserved
	if vpHeight < 1 {
		vpHeight = 1
	}
	m.Viewport.Width = maxWidth
	m.Viewport.Height = vpHeight
	return m
}

// ScrollToSelected adjusts the viewport offset so the selected card row is
// visible. Call this only when the cursor has moved, not on every update.
func (m Model) ScrollToSelected() Model {
	vpHeight := m.Viewport.Height
	if vpHeight < 1 {
		vpHeight = 1
	}
	totalCards := len(m.Interfaces) + 2
	totalRows := (totalCards + m.CardsPerRow - 1) / m.CardsPerRow
	maxOffset := totalRows*cardHeight - vpHeight
	if maxOffset < 0 {
		maxOffset = 0
	}

	selectedRow := cursorCardRow(m.Cursor, m.CardsPerRow)
	rowTop := selectedRow * cardHeight
	rowBottom := rowTop + cardHeight - 1

	offset := m.Viewport.YOffset
	if rowTop < offset {
		offset = rowTop
	} else if rowBottom >= offset+vpHeight {
		offset = rowBottom - vpHeight + 1
	}
	if offset > maxOffset {
		offset = maxOffset
	}
	if offset < 0 {
		offset = 0
	}
	m.Viewport.YOffset = offset
	return m
}

func cursorCardRow(cursor, cardsPerRow int) int {
	if cardsPerRow < 1 {
		cardsPerRow = 1
	}
	return cursor / cardsPerRow
}

func renderInterfaceCard(m Model, icons map[string]string, index int, iface net.Interface) string {
	isSelected := index == m.Cursor
	style := cardStyle
	if isSelected {
		style = selectedCardStyle
	}

	var cardContent strings.Builder
	name := iface.Name
	icon := icons[name]
	if icon == "" {
		icon = "🔌"
	}

	nameStyle := lipgloss.NewStyle().Bold(true)
	if isSelected {
		nameStyle = nameStyle.Foreground(common.Color.Selection)
	}
	cardContent.WriteString(nameStyle.Render(icon+" "+name) + "\n")

	addrs := ipv4Labels(m.InterfaceMap, name)
	addrStyle := common.HelpTextStyle
	addr := ""
	if len(addrs) > 0 {
		addr = addrs[0]
	}
	cardContent.WriteString(addrStyle.Render(addr))

	return style.Render(cardContent.String())
}

func renderTargetCard(m Model, index int) string {
	isSelected := index == m.Cursor
	style := cardStyle
	if isSelected {
		style = selectedCardStyle
	}

	var cardContent strings.Builder
	icon := "🎯"

	nameStyle := lipgloss.NewStyle().Bold(true)
	if isSelected {
		nameStyle = nameStyle.Foreground(common.Color.Selection)
	}
	cardContent.WriteString(nameStyle.Render(icon+" Custom Target") + "\n")

	subtitleStyle := common.HelpTextStyle
	cardContent.WriteString(subtitleStyle.Render("enter IP/CIDR"))

	return style.Render(cardContent.String())
}

func renderHistoryCard(m Model, index int) string {
	isSelected := index == m.Cursor
	style := cardStyle
	if isSelected {
		style = selectedCardStyle
	}

	var cardContent strings.Builder
	icon := "📜"

	nameStyle := lipgloss.NewStyle().Bold(true)
	if isSelected {
		nameStyle = nameStyle.Foreground(common.Color.Selection)
	}
	cardContent.WriteString(nameStyle.Render(icon+" History") + "\n")

	subtitleStyle := common.HelpTextStyle
	cardContent.WriteString(subtitleStyle.Render("view past scans"))

	return style.Render(cardContent.String())
}

// ipv4Labels returns IPv4 labels for an interface
func ipv4Labels(addrsByIface map[string][]net.Addr, name string) []string {
	labels := make([]string, 0)
	for _, addr := range addrsByIface[name] {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
			ones, _ := ipnet.Mask.Size()
			labels = append(labels, fmt.Sprintf("%s/%d", ipnet.IP.String(), ones))
		}
	}
	return labels
}
