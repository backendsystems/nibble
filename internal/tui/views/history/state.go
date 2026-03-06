package historyview

import (
	"github.com/backendsystems/nibble/internal/history"
	"github.com/backendsystems/nibble/internal/tui/views/history/delete"
	detailsview "github.com/backendsystems/nibble/internal/tui/views/history/details"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type NodeType int

const (
	NodeInterface NodeType = iota
	NodeNetwork
	NodeScan
)

type ViewMode int

const (
	ViewList ViewMode = iota
	ViewDetail
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

type Model struct {
	Mode         ViewMode
	Tree         []*TreeNode
	FlatList     []*TreeNode // Flattened view of expanded tree
	Cursor       int
	ShowHelp     bool
	DeleteDialog *delete.HistoryDeleteDialog // nil when not shown
	ErrorMsg     string
	Viewport     viewport.Model
	WindowW      int
	WindowH      int
	Details      detailsview.Model
}

type UpdateResult struct {
	Model           Model
	Quit            bool
	ScanAllPorts    bool
	SelectedHostIP  string
	ScanHistoryPath string
	Cmd             tea.Cmd
}
