package httpapi

import (
	"path/filepath"
	"strings"

	"github.com/go-go-golems/docmgr/internal/paths"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/docmgr/pkg/models"
)

// relatedFileItem is the API representation of a RelatedFiles entry, resolved
// server-side via the one path resolver (design doc DOCMGR-200 §8.1).
//
// Path keeps the stored frontmatter form (anchored like "repo://pkg/foo.go" or
// legacy). Root + ResolvedPath describe how the existing /file viewer can load
// the file: root=repo|docs with a root-relative path, or root=abs with an
// absolute path (which the viewer cannot serve; UIs should disable "open").
type relatedFileItem struct {
	Path   string `json:"path"`
	Note   string `json:"note,omitempty"`
	Anchor string `json:"anchor,omitempty"`
	// Root is "repo", "docs" or "abs".
	Root string `json:"root,omitempty"`
	// ResolvedPath is relative to Root for repo/docs, absolute for abs.
	ResolvedPath string `json:"resolvedPath,omitempty"`
	Exists       bool   `json:"exists"`
}

// resolveRelatedFiles resolves the RelatedFiles entries of the document at
// docAbsPath against the workspace's anchors.
func resolveRelatedFiles(ws *workspace.Workspace, docAbsPath string, rfs models.RelatedFiles) []relatedFileItem {
	items := make([]relatedFileItem, 0, len(rfs))
	if ws == nil {
		for _, rf := range rfs {
			items = append(items, relatedFileItem{Path: rf.Path, Note: rf.Note})
		}
		return items
	}

	resolver := paths.NewResolver(paths.ResolverOptions{
		DocsRoot:      ws.Context().Root,
		ConfigDir:     ws.Context().ConfigDir,
		RepoRoot:      ws.Context().RepoRoot,
		WorkspaceRoot: ws.Context().WorkspaceRoot,
		DocPath:       docAbsPath,
	})

	for _, rf := range rfs {
		item := relatedFileItem{
			Path: rf.Path,
			Note: rf.Note,
		}
		raw := strings.TrimSpace(rf.Path)
		if raw != "" {
			n := resolver.Resolve(raw)
			item.Anchor = string(n.Anchor)
			item.Exists = n.Exists
			switch {
			case strings.TrimSpace(n.RepoRelative) != "":
				item.Root = "repo"
				item.ResolvedPath = filepath.ToSlash(n.RepoRelative)
			case strings.TrimSpace(n.DocsRelative) != "":
				item.Root = "docs"
				item.ResolvedPath = filepath.ToSlash(n.DocsRelative)
			case strings.TrimSpace(n.Abs) != "":
				item.Root = "abs"
				item.ResolvedPath = filepath.ToSlash(n.Abs)
			}
		}
		items = append(items, item)
	}
	return items
}
