//go:build !linux

package docker

// On Windows and macOS Docker always runs in a VM — there are no kernel bridge
// interfaces on the host regardless of whether it's Docker Desktop or Engine.
func isDesktop(_ map[string]string) bool {
	return true
}
