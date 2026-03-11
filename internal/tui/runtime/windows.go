//go:build windows

package runtime

import "github.com/atotto/clipboard"

var savedClipboard string

func PrepareRuntime() {
	// Save current clipboard content
	savedClipboard, _ = clipboard.ReadAll()

	// Clear clipboard to initialize Windows clipboard functionality
	_ = clipboard.WriteAll("")
}

func RestoreRuntime() {
	// Restore original clipboard content
	if savedClipboard != "" {
		_ = clipboard.WriteAll(savedClipboard)
	}
}
