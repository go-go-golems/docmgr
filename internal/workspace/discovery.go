package workspace

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FindTicketScaffoldsMissingIndex finds directories that look like a ticket workspace
// (they contain scaffold subdirectories such as design/ or .meta/) but are missing index.md.
// The optional skipDir predicate can be used to omit ignored paths.
//
// NOTE: Doctor uses this scan to report "missing_index" issues. This state cannot be derived
// purely from the Workspace index because missing index.md means there is no doc to index.
func FindTicketScaffoldsMissingIndex(ctx context.Context, root string, skipDir func(relPath, baseName string) bool) ([]string, error) {
	missing := []string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
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
