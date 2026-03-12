package historydetailview

import (
	"github.com/backendsystems/nibble/internal/history"
	"github.com/backendsystems/nibble/internal/scanner/shared"
	deletepkg "github.com/backendsystems/nibble/internal/tui/views/history/delete"
	"github.com/charmbracelet/bubbles/stopwatch"
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
	NodePath        string
	NodeName        string
	NodeItemType    string
	HoveredHelpItem int // -1 means no hover, otherwise index of helpline item
	HelpLineY       int // Y row where the helpline starts, set during render

	// Rescan state
	Scanning         bool
	ScanningHostIdx  int // Index of host being scanned (-1 if none)
	ProgressChan     chan shared.ProgressUpdate
	Stopwatch        stopwatch.Model
	NewPortsByHost   map[string]map[int]bool // Track newly found ports per host IP
	ScannedCount     int
	TotalHosts       int
	ScannedHostStr   string // Last host string from scanner for the scanned host
	ScanPortsScanned []int  // Ports that were scanned in the current rescan
}
