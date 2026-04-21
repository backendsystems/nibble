//go:build windows

package docker

import (
	"context"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Microsoft/go-winio"
)

// Docker on Windows uses a named pipe. Docker Desktop also exposes a unix
// socket under %USERPROFILE%\.docker\desktop\docker.sock when WSL is enabled,
// but the named pipe is the canonical interface.
const defaultPipePath = `\\.\pipe\docker_engine`

func socketPath() string {
	if h := os.Getenv("DOCKER_HOST"); h != "" {
		// e.g. "npipe:////./pipe/docker_engine"
		return strings.TrimPrefix(h, "npipe://")
	}
	return defaultPipePath
}

func client() *http.Client {
	pipe := socketPath()
	return &http.Client{
		Timeout: 200 * time.Millisecond,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return winio.DialPipeContext(ctx, pipe)
			},
		},
	}
}
