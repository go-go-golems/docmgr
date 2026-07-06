package commands

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-go-golems/docmgr/internal/documents"
	"github.com/go-go-golems/docmgr/internal/paths"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/docmgr/pkg/models"
)

// RelatedFileChange describes one file to relate to a document, with an
// optional note.
type RelatedFileChange struct {
	Path string
	Note string
}

// RelatedFilesUpdateResult summarizes an ApplyRelatedFilesUpdate call.
type RelatedFilesUpdateResult struct {
	Added   int
	Updated int
	Removed int
	Total   int
	Changed bool
}

// ApplyRelatedFilesUpdate applies add/remove changes to the RelatedFiles
// frontmatter of the document at targetDocPath, using the same identity and
// anchored-write rules as 'docmgr doc relate' (resolveKey/anchoredForWrite):
// entries are deduplicated by resolved absolute path, notes are merged with
// appendNote, and new entries are persisted in anchored form (repo://...,
// ws://..., docs://..., abs://...).
//
// When no effective change results, the document is left untouched and the
// returned result has Changed=false. This is the shared write primitive
// behind the HTTP API's POST /docs/relate endpoint.
func ApplyRelatedFilesUpdate(ws *workspace.Workspace, targetDocPath string, add []RelatedFileChange, remove []string) (*RelatedFilesUpdateResult, error) {
	if ws == nil {
		return nil, fmt.Errorf("nil workspace")
	}
	targetDocPath = filepath.Clean(strings.TrimSpace(targetDocPath))
	if targetDocPath == "" {
		return nil, fmt.Errorf("empty target document path")
	}

	resolver := paths.NewResolver(paths.ResolverOptions{
		DocsRoot:      ws.Context().Root,
		DocPath:       targetDocPath,
		ConfigDir:     ws.Context().ConfigDir,
		RepoRoot:      ws.Context().RepoRoot,
		WorkspaceRoot: ws.Context().WorkspaceRoot,
	})

	doc, content, err := documents.ReadDocumentWithFrontmatter(targetDocPath)
	if err != nil {
		return nil, err
	}

	// Existing entries keyed by resolved absolute path so anchored, legacy and
	// raw forms of the same file collapse into one entry (mirrors relate.go).
	current := map[string]models.RelatedFile{}
	for _, rf := range doc.RelatedFiles {
		trimmedPath := strings.TrimSpace(rf.Path)
		if trimmedPath == "" {
			continue
		}
		key := resolveKey(resolver, trimmedPath)
		if key == "" {
			continue
		}
		rf.Path = trimmedPath
		if existing, ok := current[key]; ok {
			if merged, changed := appendNote(existing.Note, rf.Note); changed {
				existing.Note = merged
				current[key] = existing
			}
			continue
		}
		current[key] = rf
	}

	res := &RelatedFilesUpdateResult{}

	// Removals first (same order as the CLI path).
	for _, raw := range remove {
		key := resolveKey(resolver, raw)
		if key == "" {
			key = filepath.ToSlash(strings.TrimSpace(raw))
		}
		if key == "" {
			continue
		}
		if _, ok := current[key]; ok {
			delete(current, key)
			res.Removed++
		}
	}

	// Additions / note updates. New entries are written in anchored form.
	for _, chg := range add {
		rawPath := strings.TrimSpace(chg.Path)
		if rawPath == "" {
			continue
		}
		key := resolveKey(resolver, rawPath)
		if key == "" {
			continue
		}
		note := strings.TrimSpace(chg.Note)
		if rf, ok := current[key]; ok {
			if note != "" {
				if merged, changed := appendNote(rf.Note, note); changed {
					rf.Note = merged
					current[key] = rf
					res.Updated++
				}
			}
			continue
		}
		current[key] = models.RelatedFile{Path: anchoredForWrite(resolver, rawPath), Note: note}
		res.Added++
	}

	out := make(models.RelatedFiles, 0, len(current))
	for _, rf := range current {
		out = append(out, rf)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	res.Total = len(out)

	if res.Added == 0 && res.Updated == 0 && res.Removed == 0 {
		return res, nil
	}

	doc.RelatedFiles = out
	if err := documents.WriteDocumentWithFrontmatter(targetDocPath, doc, content, true); err != nil {
		return nil, fmt.Errorf("failed to write document: %w", err)
	}
	res.Changed = true
	return res, nil
}
