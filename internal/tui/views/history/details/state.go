package historydetailview

import (
	"github.com/backendsystems/nibble/internal/history"
	deletepkg "github.com/backendsystems/nibble/internal/tui/views/history/delete"
	"github.com/charmbracelet/bubbles/viewport"
)

type Model struct {
	History      history.ScanHistory
	HistoryPath  string
	Cursor       int
	ShowHelp     bool
	ErrorMsg     string
	Viewport     viewport.Model
	DeleteDialog *deletepkg.HistoryDeleteDialog
	WindowW      int
	WindowH      int
	NodePath     string
	NodeName     string
	NodeItemType string
}
