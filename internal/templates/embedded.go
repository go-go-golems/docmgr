package templates

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed embedded/_templates/*.md
var embeddedTemplatesFS embed.FS

//go:embed embedded/_guidelines/*.md
var embeddedGuidelinesFS embed.FS

// LoadEmbeddedTemplate loads a template from the embedded filesystem
func LoadEmbeddedTemplate(docType string) (string, bool) {
	path := fmt.Sprintf("embedded/_templates/%s.md", docType)
	content, err := embeddedTemplatesFS.ReadFile(path)
	if err != nil {
		return "", false
	}
	return string(content), true
}

// LoadEmbeddedGuideline loads a guideline from the embedded filesystem
func LoadEmbeddedGuideline(docType string) (string, bool) {
	path := fmt.Sprintf("embedded/_guidelines/%s.md", docType)
	content, err := embeddedGuidelinesFS.ReadFile(path)
	if err != nil {
		return "", false
	}
	return string(content), true
}

// LoadTemplate loads a template from filesystem only (for user customization)
// Note: Embedded templates are ONLY used for scaffolding via docmgr init, NOT in runtime resolution
// If no template is found, callers should create a minimal document with just frontmatter
func LoadTemplate(root, docType string) (string, bool) {
	// Only try filesystem (allows user customization)
	path := filepath.Join(root, "_templates", docType+".md")
	if b, err := os.ReadFile(path); err == nil {
		return string(b), true
	}

	// No template found - caller will create minimal doc
	return "", false
}

// LoadGuideline loads a guideline from filesystem only (for user customization)
// Note: Embedded guidelines are ONLY used for scaffolding via docmgr init, NOT in runtime resolution
// If no guideline is found, callers should handle gracefully (no guideline shown)
func LoadGuideline(root, docType string) (string, bool) {
	// Only try filesystem (allows user customization)
	path := filepath.Join(root, "_guidelines", docType+".md")
	if b, err := os.ReadFile(path); err == nil {
		return string(b), true
	}

	// No guideline found - caller will handle gracefully
	return "", false
}

// ListEmbeddedTemplates returns all template doc types available in embedded FS
func ListEmbeddedTemplates() ([]string, error) {
	var docTypes []string
	err := fs.WalkDir(embeddedTemplatesFS, "embedded/_templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".md" {
			base := filepath.Base(path)
			docType := base[:len(base)-len(".md")]
			docTypes = append(docTypes, docType)
		}
		return nil
	})
	return docTypes, err
}

// ListEmbeddedGuidelines returns all guideline doc types available in embedded FS
func ListEmbeddedGuidelines() ([]string, error) {
	var docTypes []string
	err := fs.WalkDir(embeddedGuidelinesFS, "embedded/_guidelines", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".md" {
			base := filepath.Base(path)
			docType := base[:len(base)-len(".md")]
			docTypes = append(docTypes, docType)
		}
		return nil
	})
	return docTypes, err
}
