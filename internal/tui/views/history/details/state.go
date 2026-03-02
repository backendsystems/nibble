package historydetailview

import (
	"github.com/backendsystems/nibble/internal/history"
	"github.com/backendsystems/nibble/internal/scanner/shared"
	"github.com/charmbracelet/bubbles/viewport"
)

type Model struct {
	History       history.ScanHistory
	HistoryPath   string
	Cursor        int
	ShowHelp      bool
	ErrorMsg      string
	Viewport      viewport.Model
	NetworkScan   shared.Scanner
}
