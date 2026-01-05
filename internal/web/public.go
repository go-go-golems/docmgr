package web

import (
	"io/fs"
)

func PublicFS() (fs.FS, error) {
	return publicFS()
}
