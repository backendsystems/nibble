package historyview

import (
	"fmt"
	"strings"

	"github.com/backendsystems/nibble/internal/history"
	"github.com/backendsystems/nibble/internal/tui/views/common"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

type NodeType int

const (
	NodeInterface NodeType = iota
	NodeNetwork
	NodeScan
)

type TreeNode struct {
	Type      NodeType
	Name      string
	Path      string
	Expanded  bool
	Children  []*TreeNode
	ScanData  *history.ScanHistory
	Level     int
}

type ViewMode int

const (
	ViewList ViewMode = iota
	ViewDetail
)

type Model struct {
	Mode         ViewMode
	Tree         []*TreeNode
	FlatList     []*TreeNode // Flattened view of expanded tree
	Cursor       int
	ShowHelp     bool
	DeleteDialog *DeleteDialog // nil when not shown
	ErrorMsg     string
	Viewport     viewport.Model
	WindowW      int
	WindowH      int

	// Detail view state
	DetailHistory  *history.ScanHistory
	DetailPath     string
	DetailCursor   int
	DetailViewport viewport.Model
}

type UpdateResult struct {
	Model           Model
	Quit            bool
	ScanAllPorts    bool
	SelectedHostIP  string
	ScanHistoryPath string
}

// DeleteDialog is a delete confirmation dialog for the history view
type DeleteDialog struct {
	Target      *TreeNode
	CursorOnYes bool // true = Delete selected, false = Cancel selected
}

// Render displays the delete confirmation dialog
func (d DeleteDialog) Render(view string, viewWidth, viewHeight int) string {
	if d.Target == nil {
		return view
	}

	var itemType string
	var itemName string

	switch d.Target.Type {
	case NodeScan:
		itemType = "scan"
		itemName = d.Target.Name
	case NodeNetwork:
		itemType = "all scans in network"
		itemName = d.Target.Name
	case NodeInterface:
		itemType = "all scans on interface"
		itemName = d.Target.Name
	}

	// Button styles
	buttonStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Margin(0, 1)

	selectedButtonStyle := buttonStyle.
		Foreground(lipgloss.Color("0")).
		Background(common.Color.Selection).
		Bold(true)

	unselectedButtonStyle := buttonStyle.
		Foreground(common.Color.Info).
		Background(lipgloss.Color("236"))

	// Build content
	warning := common.ErrorStyle.Render(fmt.Sprintf("Delete %s: %s?", itemType, itemName))
	note := common.HelpTextStyle.Render("This action cannot be undone.")

	// Buttons (Cancel on left, Delete on right)
	var cancelBtn, deleteBtn string
	if d.CursorOnYes {
		cancelBtn = unselectedButtonStyle.Render("Cancel")
		deleteBtn = selectedButtonStyle.Render("Delete")
	} else {
		cancelBtn = selectedButtonStyle.Render("Cancel")
		deleteBtn = unselectedButtonStyle.Render("Delete")
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, cancelBtn, deleteBtn)
	help := common.HelpTextStyle.Render("←/→: navigate • Enter: select • q: back")

	// Combine content
	content := strings.Join([]string{
		warning,
		note,
		"",
		buttons,
		"",
		help,
	}, "\n")

	// Calculate box width
	width := int(float64(viewWidth) * 0.6)
	if width < 46 {
		width = 46
	}
	if width > 60 {
		width = 60
	}

	// Create overlay with common style
	overlay := common.HelpBoxStyle.Width(width).Render(content)

	// Place overlay over the view (top-centered like help dialog)
	return lipgloss.Place(
		viewWidth,
		viewHeight,
		lipgloss.Center,
		lipgloss.Top,
		overlay,
		lipgloss.WithWhitespaceChars(" "),
	)
}
