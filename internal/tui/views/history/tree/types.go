package tree

import (
	"context"
	"time"

	"github.com/backendsystems/nibble/internal/history"
)

type NodeType int

const (
	NodeInterface NodeType = iota
	NodeNetwork
	NodeScan
)

type ScanCounts struct {
	Hosts int
	Ports int
}

type Node struct {
	Type       NodeType
	Name       string
	Path       string
	Expanded   bool
	Children   []*Node
	ScanData   *history.ScanHistory
	Created    time.Time
	Counts     *ScanCounts
	CancelLoad context.CancelFunc
	Level      int
}
