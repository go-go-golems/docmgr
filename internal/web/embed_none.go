//go:build !embed

package web

import (
	"fmt"
	"io/fs"
	"os"
)

func publicFS() (fs.FS, error) {
	// Best-effort: serve on-disk generated assets in dev builds (after `go generate ./internal/web`).
	const dir = "internal/web/embed/public"
	if _, err := os.Stat(dir); err != nil {
		return nil, fmt.Errorf("web assets not available at %s: %w", dir, err)
	}
	return os.DirFS(dir), nil
}
