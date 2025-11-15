package commands

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-go-golems/docmgr/pkg/models"
)

// TicketWorkspace represents a discovered workspace directory and its metadata.
type TicketWorkspace struct {
	Path           string
	Doc            *models.Document
	FrontmatterErr error
}

// collectTicketWorkspaces walks the docs root and returns directories that contain
// an index.md file with valid frontmatter. Directories whose base name starts with
// an underscore are ignored (for example, _templates, _guidelines). The optional
// skipDir predicate can be used to skip additional directories by relative path
// or base name.
func collectTicketWorkspaces(root string, skipDir func(relPath, baseName string) bool) ([]TicketWorkspace, error) {
	workspaces := []TicketWorkspace{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if path == root {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			rel = path
		}
		base := d.Name()
		if strings.HasPrefix(base, "_") {
			return fs.SkipDir
		}
		if skipDir != nil && skipDir(rel, base) {
			return fs.SkipDir
		}

		indexPath := filepath.Join(path, "index.md")
		if fi, err := os.Stat(indexPath); err == nil && !fi.IsDir() {
			doc, err := readDocumentFrontmatter(indexPath)
			if err != nil {
				workspaces = append(workspaces, TicketWorkspace{Path: path, FrontmatterErr: err})
			} else {
				workspaces = append(workspaces, TicketWorkspace{Path: path, Doc: doc})
			}
			return fs.SkipDir
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(workspaces, func(i, j int) bool {
		return workspaces[i].Path < workspaces[j].Path
	})
	return workspaces, nil
}

// collectTicketScaffoldsWithoutIndex finds directories that look like a ticket workspace
// (they contain scaffold subdirectories such as design/ or .meta/) but are missing index.md.
// The same skipDir predicate used in collectTicketWorkspaces can be provided to omit
// ignored paths.
func collectTicketScaffoldsWithoutIndex(root string, skipDir func(relPath, baseName string) bool) ([]string, error) {
	missing := []string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if path == root {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			rel = path
		}
		base := d.Name()
		if strings.HasPrefix(base, "_") {
			return fs.SkipDir
		}
		if skipDir != nil && skipDir(rel, base) {
			return fs.SkipDir
		}
		indexPath := filepath.Join(path, "index.md")
		if _, err := os.Stat(indexPath); err == nil {
			return fs.SkipDir
		}
		if hasWorkspaceScaffold(path) {
			missing = append(missing, path)
			return fs.SkipDir
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(missing)
	return missing, nil
}

var workspaceStructureMarkers = []string{"design", "reference", "playbooks", "scripts", "sources", "various", "archive", ".meta"}

func hasWorkspaceScaffold(path string) bool {
	for _, marker := range workspaceStructureMarkers {
		candidate := filepath.Join(path, marker)
		if fi, err := os.Stat(candidate); err == nil && fi.IsDir() {
			return true
		}
	}
	return false
}
