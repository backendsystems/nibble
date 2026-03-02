package historyview

import (
	"github.com/backendsystems/nibble/internal/history"
	"github.com/charmbracelet/bubbles/viewport"
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
