package mainview

import (
	"fmt"
	"net"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/backendsystems/nibble/internal/tui/views/common"
)

func Render(m *Model, maxWidth int) string {
	cardsPerRow := m.CardsPerRow
	if cardsPerRow == 0 {
		cardsPerRow = 1
	}

	icons := make(map[string]string, len(m.Interfaces))
	for _, iface := range m.Interfaces {
		_, isDocker := m.DockerIfaces[iface.Name]
		icons[iface.Name] = interfaceIcon(iface.Name, isDocker)
	}

	var rows []string
	var currentRow []string
	var rowHeights []int

	flushRow := func() {
		joined := lipgloss.JoinHorizontal(lipgloss.Top, currentRow...)
		rows = append(rows, joined)
		rowHeights = append(rowHeights, strings.Count(joined, "\n")+1)
		currentRow = nil
	}

	for i, iface := range m.Interfaces {
		card := renderInterfaceCard(*m, icons, i, iface)
		currentRow = append(currentRow, card)
		if len(currentRow) == cardsPerRow {
			flushRow()
		}
	}

	targetCard := renderTargetCard(*m, len(m.Interfaces))
	currentRow = append(currentRow, targetCard)
	if len(currentRow) == cardsPerRow {
		flushRow()
	}

	historyCard := renderHistoryCard(*m, len(m.Interfaces)+1)
	currentRow = append(currentRow, historyCard)
	if len(currentRow) > 0 {
		flushRow()
	}

	m.RowHeights = rowHeights

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
	m.HelpLineY = strings.Count(b.String(), "\n") + 1
	layout := common.BuildHelpLineLayout(mainHelpItems, helpPrefixText, maxWidth)
	b.WriteString("\n" + common.RenderHelpLine(layout, helpPrefixText, maxWidth, m.HoveredHelpItem))

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
	vpHeight := max(m.WindowH-reserved, 1)
	m.Viewport.SetWidth(maxWidth)
	m.Viewport.SetHeight(vpHeight)
	return m
}

// ScrollToSelected adjusts the viewport offset so the selected card row is
// visible. Call this only when the cursor has moved, not on every update.
func (m Model) ScrollToSelected() Model {
	vpHeight := max(m.Viewport.Height(), 1)

	selectedRow := cursorCardRow(m.Cursor, m.CardsPerRow)

	rowTop, rowBottom, totalH := 0, cardHeight-1, 0
	if len(m.RowHeights) > 0 {
		acc := 0
		for i, h := range m.RowHeights {
			totalH += h
			if i < selectedRow {
				acc += h
			} else if i == selectedRow {
				rowTop = acc
				rowBottom = acc + h - 1
			}
		}
	} else {
		rowTop = selectedRow * cardHeight
		rowBottom = rowTop + cardHeight - 1
		totalH = ((len(m.Interfaces) + 2 + m.CardsPerRow - 1) / m.CardsPerRow) * cardHeight
	}
	maxOffset := max(totalH-vpHeight, 0)

	offset := m.Viewport.YOffset()
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
	m.Viewport.SetYOffset(offset)
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
	style := cardStyle(m.CardWidth)
	if isSelected {
		style = selectedCardStyle(m.CardWidth)
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
	style := cardStyle(m.CardWidth)
	if isSelected {
		style = selectedCardStyle(m.CardWidth)
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
	style := cardStyle(m.CardWidth)
	if isSelected {
		style = selectedCardStyle(m.CardWidth)
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
