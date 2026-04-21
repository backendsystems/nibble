//go:build !windows

package docker

import (
	"context"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

const defaultSocketPath = "/var/run/docker.sock"

func socketPath() string {
	if h := os.Getenv("DOCKER_HOST"); h != "" {
		return strings.TrimPrefix(h, "unix://")
	}
	home, _ := os.UserHomeDir()
	if desktop := home + "/.docker/desktop/docker.sock"; fileExists(desktop) {
		return desktop
	}
	return defaultSocketPath
}

func client() *http.Client {
	sock := socketPath()
	return &http.Client{
		Timeout: 200 * time.Millisecond,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, "unix", sock)
			},
		},
	}
}
