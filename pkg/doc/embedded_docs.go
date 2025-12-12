package doc

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// EmbeddedDoc represents a single embedded documentation file.
type EmbeddedDoc struct {
	Name    string // file name (relative path within pkg/doc)
	Content string // UTF-8 text content
}

// ReadEmbeddedMarkdownDocs returns all embedded markdown docs shipped with docmgr.
//
// This is intended for debugging/export tooling (e.g. exporting a SQLite DB that is
// self-describing). We deliberately include only *.md files.
func ReadEmbeddedMarkdownDocs() ([]EmbeddedDoc, error) {
	var out []EmbeddedDoc
	err := fs.WalkDir(docFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.ToLower(filepath.Ext(path)) != ".md" {
			return nil
		}
		b, err := fs.ReadFile(docFS, path)
		if err != nil {
			return err
		}
		out = append(out, EmbeddedDoc{
			Name:    filepath.ToSlash(path),
			Content: string(b),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}


