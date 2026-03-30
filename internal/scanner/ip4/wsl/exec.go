package wsl

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// winCmdPath resolves a Windows executable name to its full path inside WSL.
// It first checks $WINDIR (set by WSL from the Windows environment), then falls
// back to the conventional mount point /mnt/c/Windows/System32.
func winCmdPath(name string) string {
	if windir := os.Getenv("WINDIR"); windir != "" {
		// WINDIR is a Windows path like "C:\Windows"; translate the drive letter.
		// e.g. "C:\Windows" -> "/mnt/c/Windows"
		if len(windir) >= 2 && windir[1] == ':' {
			drive := strings.ToLower(string(windir[0]))
			rest := filepath.ToSlash(windir[2:])
			candidate := "/mnt/" + drive + rest + "/System32/" + name
			if _, err := os.Stat(candidate); err == nil {
				return candidate
			}
		}
	}
	// Conventional fallback
	fallback := "/mnt/c/Windows/System32/" + name
	if _, err := os.Stat(fallback); err == nil {
		return fallback
	}
	// Last resort: let the shell find it (works when Windows paths are in $PATH)
	return name
}

func runWinCmd(name string, args ...string) (string, error) {
	cmd := exec.Command(winCmdPath(name), args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return buf.String(), nil
}
