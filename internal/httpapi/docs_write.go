package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/docmgr/pkg/commands"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/core"
)

type docsMetaRequest struct {
	Path  string `json:"path"`
	Field string `json:"field"`
	Value string `json:"value"`
}

type docsMetaResponse struct {
	Path   string `json:"path"`
	Field  string `json:"field"`
	Value  string `json:"value"`
	Status string `json:"status"`
}

// handleDocsMeta wraps the 'docmgr meta update' write primitive
// (commands.UpdateDocumentField): POST {path, field, value} updates one
// frontmatter field of a document under the docs root and refreshes the
// in-memory index.
func (s *Server) handleDocsMeta(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return NewHTTPError(http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
	}

	var req docsMetaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return NewHTTPError(http.StatusBadRequest, "invalid_argument", "invalid json body", nil)
	}
	req.Path = strings.TrimSpace(req.Path)
	req.Field = strings.TrimSpace(req.Field)
	if req.Path == "" {
		return NewHTTPError(http.StatusBadRequest, "invalid_argument", "missing path", map[string]any{"field": "path"})
	}
	if req.Field == "" {
		return NewHTTPError(http.StatusBadRequest, "invalid_argument", "missing field", map[string]any{"field": "field"})
	}

	var resp docsMetaResponse
	if err := s.mgr.WithWorkspace(func(ws *workspace.Workspace) error {
		abs, rel, err := resolveDocWithin(ws.Context().Root, req.Path)
		if err != nil {
			return err
		}

		if err := commands.UpdateDocumentField(abs, req.Field, req.Value); err != nil {
			if errors.Is(err, commands.ErrUnknownMetaField) {
				return NewHTTPError(http.StatusBadRequest, "invalid_argument", err.Error(), map[string]any{
					"field": "field",
					"value": req.Field,
				})
			}
			if t, ok := core.AsTaxonomy(err); ok {
				return NewHTTPError(http.StatusUnprocessableEntity, "invalid_frontmatter", err.Error(), map[string]any{
					"path":     rel,
					"taxonomy": t,
				})
			}
			return err
		}

		resp = docsMetaResponse{Path: rel, Field: req.Field, Value: req.Value, Status: "updated"}
		return nil
	}); err != nil {
		return err
	}

	if _, err := s.mgr.Refresh(r.Context()); err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, resp)
}

type docsRelateAddItem struct {
	Path string `json:"path"`
	Note string `json:"note"`
}

type docsRelateRequest struct {
	Path   string              `json:"path"`
	Add    []docsRelateAddItem `json:"add"`
	Remove []string            `json:"remove"`
}

type docsRelateResponse struct {
	Path    string `json:"path"`
	Added   int    `json:"added"`
	Updated int    `json:"updated"`
	Removed int    `json:"removed"`
	Total   int    `json:"total"`
	Status  string `json:"status"`
}

// handleDocsRelate wraps the 'docmgr doc relate' write primitive
// (commands.ApplyRelatedFilesUpdate), including anchored writes for new
// entries, and refreshes the in-memory index afterwards.
func (s *Server) handleDocsRelate(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return NewHTTPError(http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
	}

	var req docsRelateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return NewHTTPError(http.StatusBadRequest, "invalid_argument", "invalid json body", nil)
	}
	req.Path = strings.TrimSpace(req.Path)
	if req.Path == "" {
		return NewHTTPError(http.StatusBadRequest, "invalid_argument", "missing path", map[string]any{"field": "path"})
	}
	if len(req.Add) == 0 && len(req.Remove) == 0 {
		return NewHTTPError(http.StatusBadRequest, "invalid_argument", "nothing to do: provide add and/or remove", map[string]any{"field": "add"})
	}
	for _, item := range req.Add {
		if strings.TrimSpace(item.Path) == "" {
			return NewHTTPError(http.StatusBadRequest, "invalid_argument", "add entries need a non-empty path", map[string]any{"field": "add"})
		}
	}

	var resp docsRelateResponse
	if err := s.mgr.WithWorkspace(func(ws *workspace.Workspace) error {
		abs, rel, err := resolveDocWithin(ws.Context().Root, req.Path)
		if err != nil {
			return err
		}

		add := make([]commands.RelatedFileChange, 0, len(req.Add))
		for _, item := range req.Add {
			add = append(add, commands.RelatedFileChange{Path: item.Path, Note: item.Note})
		}

		res, err := commands.ApplyRelatedFilesUpdate(ws, abs, add, req.Remove)
		if err != nil {
			if t, ok := core.AsTaxonomy(err); ok {
				return NewHTTPError(http.StatusUnprocessableEntity, "invalid_frontmatter", err.Error(), map[string]any{
					"path":     rel,
					"taxonomy": t,
				})
			}
			return err
		}

		status := "updated"
		if !res.Changed {
			status = "noop"
		}
		resp = docsRelateResponse{
			Path:    rel,
			Added:   res.Added,
			Updated: res.Updated,
			Removed: res.Removed,
			Total:   res.Total,
			Status:  status,
		}
		return nil
	}); err != nil {
		return err
	}

	if _, err := s.mgr.Refresh(r.Context()); err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, resp)
}

// resolveDocWithin resolves a markdown document path within rootDir via the
// traversal-safe resolveFileWithin and rejects directories and non-markdown
// files.
func resolveDocWithin(rootDir string, rawPath string) (string, string, error) {
	abs, rel, fi, err := resolveFileWithin(rootDir, rawPath)
	if err != nil {
		return "", "", err
	}
	if fi.IsDir() {
		return "", "", NewHTTPError(http.StatusBadRequest, "invalid_argument", "path is a directory", map[string]any{
			"field": "path",
			"value": rawPath,
		})
	}
	if !strings.HasSuffix(strings.ToLower(rel), ".md") {
		return "", "", NewHTTPError(http.StatusBadRequest, "invalid_argument", "path is not a markdown document", map[string]any{
			"field": "path",
			"value": rawPath,
		})
	}
	return abs, rel, nil
}
