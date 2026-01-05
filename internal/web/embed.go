//go:build embed

package web

import (
	"embed"
	"io/fs"
)

//go:embed embed/public
var embeddedPublicFS embed.FS

func publicFS() (fs.FS, error) {
	return fs.Sub(embeddedPublicFS, "embed/public")
}
