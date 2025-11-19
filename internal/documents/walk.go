package documents

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/docmgr/pkg/models"
)

// WalkDocumentFunc is invoked for every markdown document encountered while walking.
// doc and body are nil if readErr is non-nil.
type WalkDocumentFunc func(path string, doc *models.Document, body string, readErr error) error

type walkConfig struct {
	skipDir func(path string, d fs.DirEntry) bool
}

// WalkOption customizes the behavior of WalkDocuments.
type WalkOption func(*walkConfig)

// WithSkipDir configures a predicate that, when it returns true, skips the directory.
func WithSkipDir(skip func(path string, d fs.DirEntry) bool) WalkOption {
	return func(cfg *walkConfig) {
		cfg.skipDir = skip
	}
}

// WalkDocuments walks the root directory, invoking fn for every markdown document encountered.
// Directories beginning with "_" are skipped by default. Use WithSkipDir to customize.
func WalkDocuments(root string, fn WalkDocumentFunc, opts ...WalkOption) error {
	cfg := &walkConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() != "." && strings.HasPrefix(d.Name(), "_") {
				return fs.SkipDir
			}
			if cfg.skipDir != nil && cfg.skipDir(path, d) {
				return fs.SkipDir
			}
			return nil
		}
		if strings.ToLower(filepath.Ext(d.Name())) != ".md" {
			return nil
		}

		doc, body, readErr := ReadDocumentWithFrontmatter(path)
		return fn(path, doc, body, readErr)
	})
}
