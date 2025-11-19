package documents

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/frontmatter"
	"github.com/go-go-golems/docmgr/pkg/models"
	"gopkg.in/yaml.v3"
)

// ReadDocumentWithFrontmatter reads a markdown file that contains YAML frontmatter.
// It returns the parsed Document metadata along with the markdown body content.
func ReadDocumentWithFrontmatter(path string) (*models.Document, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer func() { _ = f.Close() }()

	var doc models.Document
	body, err := frontmatter.Parse(f, &doc)
	if err != nil {
		return nil, "", fmt.Errorf("parse frontmatter in %s: %w", path, err)
	}

	return &doc, string(body), nil
}

// WriteDocumentWithFrontmatter writes the provided document metadata and body
// to the target markdown path using YAML frontmatter.
func WriteDocumentWithFrontmatter(path string, doc *models.Document, body string, force bool) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".docmgr-*")
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(tmp.Name())
	}()

	if _, err := tmp.WriteString("---\n"); err != nil {
		_ = tmp.Close()
		return err
	}

	enc := yaml.NewEncoder(tmp)
	if err := enc.Encode(doc); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := enc.Close(); err != nil {
		_ = tmp.Close()
		return err
	}

	if _, err := tmp.WriteString("---\n\n"); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.WriteString(body); err != nil {
		_ = tmp.Close()
		return err
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmp.Name(), path)
}
