// Package scanner selects a scanner implementation based on runtime mode
package scanner

import (
	"github.com/backendsystems/nibble/internal/scanner/demo"
	"github.com/backendsystems/nibble/internal/scanner/ip4"
	"github.com/backendsystems/nibble/internal/scanner/shared"
)

// New returns a Scanner for the given mode
func New(demoMode bool) shared.Scanner {
	if demoMode {
		return &demo.Scanner{}
	}
	return &ip4.Scanner{}
}
