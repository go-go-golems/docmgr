package searchsvc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/docmgr/internal/documents"
	"github.com/go-go-golems/docmgr/internal/paths"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/core"
	"github.com/go-go-golems/docmgr/pkg/models"
)

type SearchQuery struct {
	// TextQuery is an FTS5 query string (no compatibility guarantees).
	TextQuery string

	Ticket  string
	Topics  []string
	DocType string
	Status  string

	File string
	Dir  string

	ExternalSource string
	Since          string
	Until          string
	CreatedSince   string
	UpdatedSince   string

	OrderBy workspace.OrderBy
	Reverse bool

	IncludeArchivedPath bool
	IncludeScriptsPath  bool
	IncludeControlDocs  bool
	IncludeDiagnostics  bool
	IncludeErrors       bool
}

type SearchResult struct {
	Ticket  string
	Title   string
	DocType string
	Status  string
	Topics  []string
	Path    string
	Snippet string

	MatchedFiles []string
	MatchedNotes []string
}

type SearchResponse struct {
	Total       int
	Results     []SearchResult
	Diagnostics []core.Taxonomy
}

func SearchDocs(ctx context.Context, ws *workspace.Workspace, q SearchQuery) (SearchResponse, error) {
	if ctx == nil {
		return SearchResponse{}, fmt.Errorf("nil context")
	}
	if ws == nil {
		return SearchResponse{}, fmt.Errorf("nil workspace")
	}

	if strings.TrimSpace(q.TextQuery) == "" &&
		strings.TrimSpace(q.Ticket) == "" &&
		len(q.Topics) == 0 &&
		strings.TrimSpace(q.DocType) == "" &&
		strings.TrimSpace(q.Status) == "" &&
		strings.TrimSpace(q.File) == "" &&
		strings.TrimSpace(q.Dir) == "" &&
		strings.TrimSpace(q.ExternalSource) == "" &&
		strings.TrimSpace(q.Since) == "" &&
		strings.TrimSpace(q.Until) == "" &&
		strings.TrimSpace(q.CreatedSince) == "" &&
		strings.TrimSpace(q.UpdatedSince) == "" {
		return SearchResponse{}, fmt.Errorf("must provide at least a query or filter")
	}

	sinceTime, err := ParseDate(q.Since)
	if err != nil {
		return SearchResponse{}, fmt.Errorf("invalid --since date: %w", err)
	}
	untilTime, err := ParseDate(q.Until)
	if err != nil {
		return SearchResponse{}, fmt.Errorf("invalid --until date: %w", err)
	}
	createdSinceTime, err := ParseDate(q.CreatedSince)
	if err != nil {
		return SearchResponse{}, fmt.Errorf("invalid --created-since date: %w", err)
	}
	updatedSinceTime, err := ParseDate(q.UpdatedSince)
	if err != nil {
		return SearchResponse{}, fmt.Errorf("invalid --updated-since date: %w", err)
	}

	scope := workspace.Scope{Kind: workspace.ScopeRepo}
	if strings.TrimSpace(q.Ticket) != "" {
		scope = workspace.Scope{Kind: workspace.ScopeTicket, TicketID: strings.TrimSpace(q.Ticket)}
	}

	docQuery := workspace.DocQuery{
		Scope: scope,
		Filters: workspace.DocFilters{
			Ticket:    strings.TrimSpace(q.Ticket),
			Status:    strings.TrimSpace(q.Status),
			DocType:   strings.TrimSpace(q.DocType),
			TopicsAny: q.Topics,
			TextQuery: strings.TrimSpace(q.TextQuery),
			RelatedFile: func() []string {
				if strings.TrimSpace(q.File) == "" {
					return nil
				}
				return []string{strings.TrimSpace(q.File)}
			}(),
			RelatedDir: func() []string {
				if strings.TrimSpace(q.Dir) == "" {
					return nil
				}
				return []string{strings.TrimSpace(q.Dir)}
			}(),
		},
		Options: workspace.DocQueryOptions{
			IncludeBody:         true,
			IncludeErrors:       q.IncludeErrors,
			IncludeDiagnostics:  q.IncludeDiagnostics,
			IncludeArchivedPath: q.IncludeArchivedPath,
			IncludeScriptsPath:  q.IncludeScriptsPath,
			IncludeControlDocs:  q.IncludeControlDocs,
			OrderBy:             q.OrderBy,
			Reverse:             q.Reverse,
		},
	}

	res, err := ws.QueryDocs(ctx, docQuery)
	if err != nil {
		return SearchResponse{}, err
	}

	out := make([]SearchResult, 0, len(res.Docs))

	fileQueryRaw := strings.TrimSpace(q.File)
	for _, h := range res.Docs {
		if h.Doc == nil {
			continue
		}

		doc := h.Doc
		content := h.Body
		if strings.TrimSpace(content) == "" {
			// Fallback: load body from disk if not included in the index.
			_, body, rerr := documents.ReadDocumentWithFrontmatter(h.Path)
			if rerr == nil {
				content = body
			}
		}

		// External source filter (best-effort; re-read frontmatter).
		if strings.TrimSpace(q.ExternalSource) != "" {
			fm, ferr := readFrontmatter(h.Path)
			if ferr != nil {
				continue
			}
			if !externalSourceMatch(fm.ExternalSources, q.ExternalSource) {
				continue
			}
		}

		// Date filters.
		if fi, err := os.Stat(h.Path); err == nil {
			createdTime := fi.ModTime()
			if !createdSinceTime.IsZero() && createdTime.Before(createdSinceTime) {
				continue
			}
		}
		if !doc.LastUpdated.IsZero() {
			if !sinceTime.IsZero() && doc.LastUpdated.Before(sinceTime) {
				continue
			}
			if !untilTime.IsZero() && doc.LastUpdated.After(untilTime) {
				continue
			}
			if !updatedSinceTime.IsZero() && doc.LastUpdated.Before(updatedSinceTime) {
				continue
			}
		}

		relPath, err := filepath.Rel(ws.Context().Root, h.Path)
		if err != nil {
			relPath = h.Path
		}
		relPath = filepath.ToSlash(relPath)

		snippet := ExtractSnippet(content, q.TextQuery, 100)

		var matchedFiles []string
		var matchedNotes []string
		if fileQueryRaw != "" {
			matchedFiles, matchedNotes = matchRelatedFiles(ws, h.Path, doc.RelatedFiles, fileQueryRaw)
		}

		out = append(out, SearchResult{
			Ticket:       doc.Ticket,
			Title:        doc.Title,
			DocType:      doc.DocType,
			Status:       doc.Status,
			Topics:       append([]string{}, doc.Topics...),
			Path:         relPath,
			Snippet:      snippet,
			MatchedFiles: matchedFiles,
			MatchedNotes: matchedNotes,
		})
	}

	return SearchResponse{
		Total:       len(out),
		Results:     out,
		Diagnostics: res.Diagnostics,
	}, nil
}

func readFrontmatter(path string) (*models.Document, error) {
	doc, _, err := documents.ReadDocumentWithFrontmatter(path)
	return doc, err
}

func externalSourceMatch(externalSources []string, query string) bool {
	query = strings.TrimSpace(query)
	if query == "" {
		return true
	}
	for _, es := range externalSources {
		es = strings.TrimSpace(es)
		if es == "" {
			continue
		}
		if strings.Contains(es, query) || strings.Contains(query, es) {
			return true
		}
	}
	return false
}

func matchRelatedFiles(ws *workspace.Workspace, docPath string, relatedFiles models.RelatedFiles, fileQueryRaw string) ([]string, []string) {
	docResolver := paths.NewResolver(paths.ResolverOptions{
		DocsRoot:  ws.Context().Root,
		DocPath:   docPath,
		ConfigDir: ws.Context().ConfigDir,
		RepoRoot:  ws.Context().RepoRoot,
	})

	fileQueryNorm := docResolver.Normalize(fileQueryRaw)
	if fileQueryNorm.Empty() && strings.TrimSpace(fileQueryRaw) != "" {
		fileQueryNorm = paths.NormalizedPath{
			Canonical:     filepath.ToSlash(fileQueryRaw),
			OriginalClean: filepath.ToSlash(fileQueryRaw),
		}
	}
	baseQuery := filepath.Base(filepath.ToSlash(fileQueryRaw))

	var matchedFiles []string
	var matchedNotes []string
	for _, rf := range relatedFiles {
		n := docResolver.Normalize(rf.Path)
		if paths.MatchPaths(fileQueryNorm, n) ||
			(strings.TrimSpace(baseQuery) != "" && (filepath.Base(filepath.ToSlash(rf.Path)) == baseQuery)) ||
			(strings.TrimSpace(baseQuery) != "" && strings.HasSuffix(filepath.ToSlash(rf.Path), "/"+baseQuery)) {
			if best := strings.TrimSpace(n.Best()); best != "" {
				matchedFiles = append(matchedFiles, best)
			} else {
				matchedFiles = append(matchedFiles, filepath.ToSlash(strings.TrimSpace(rf.Path)))
			}
			if strings.TrimSpace(rf.Note) != "" {
				matchedNotes = append(matchedNotes, rf.Note)
			}
		}
	}
	return matchedFiles, matchedNotes
}
