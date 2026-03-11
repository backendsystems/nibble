//go:build windows

package tui

import "github.com/atotto/clipboard"

func prepareRuntime() {
	_ = clipboard.WriteAll("")
}
