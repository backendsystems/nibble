package wsl

import (
	"os"
)

// IsWSLVirtualIface returns true for virtual network interfaces created by WSL
// itself (eth0, the Hyper-V NAT adapter) that have no unique scannable network.
// Windows interfaces are exposed via win: names instead.
func IsWSLVirtualIface(name string) bool {
	if !IsWSL() {
		return false
	}
	// eth0 is the WSL Hyper-V NAT adapter — covered by win: interfaces
	return name == "eth0"
}

// IsWSL returns true when running inside Windows Subsystem for Linux.
// It checks for WSLInterop, which is a WSL-specific file that does not exist
// on native Linux (including Azure/Hyper-V VMs that also carry "microsoft" in
// /proc/version), macOS, or Windows.
func IsWSL() bool {
	_, err := os.Stat("/proc/sys/fs/binfmt_misc/WSLInterop")
	return err == nil
}
