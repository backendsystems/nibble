package wsl

import (
	"os"
)

// IsWSL returns true when running inside Windows Subsystem for Linux.
// It checks for WSLInterop, which is a WSL-specific file that does not exist
// on native Linux (including Azure/Hyper-V VMs that also carry "microsoft" in
// /proc/version), macOS, or Windows.
func IsWSL() bool {
	_, err := os.Stat("/proc/sys/fs/binfmt_misc/WSLInterop")
	return err == nil
}
