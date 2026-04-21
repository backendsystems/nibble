//go:build linux

package docker

import "os"

// isDesktop returns true when none of the known bridge kernel interfaces exist
// in sysfs — meaning Docker is VM-backed (Docker Desktop).
func isDesktop(networks map[string]string) bool {
	for kernelName := range networks {
		if _, err := os.Stat("/sys/class/net/" + kernelName); err == nil {
			return false
		}
	}
	return true
}
