package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-go-golems/docmgr/internal/paths"
	"github.com/go-go-golems/docmgr/internal/tasksmd"
	"github.com/go-go-golems/docmgr/internal/ticketgraph"
	"github.com/go-go-golems/docmgr/internal/tickets"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/docmgr/pkg/models"
)

type ticketStats struct {
	DocsTotal         int `json:"docsTotal"`
	TasksTotal        int `json:"tasksTotal"`
	TasksDone         int `json:"tasksDone"`
	RelatedFilesTotal int `json:"relatedFilesTotal"`
}

type ticketGetResponse struct {
	Ticket    string   `json:"ticket"`
	Title     string   `json:"title"`
	Status    string   `json:"status"`
	Intent    string   `json:"intent"`
	Owners    []string `json:"owners"`
	Topics    []string `json:"topics"`
	CreatedAt string   `json:"createdAt"`
	UpdatedAt string   `json:"updatedAt"`

	TicketDir string `json:"ticketDir"`
	IndexPath string `json:"indexPath"`

	Stats ticketStats `json:"stats"`
}

func (s *Server) handleTicketsGet(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return NewHTTPError(http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
	}

	ticketID := strings.TrimSpace(r.URL.Query().Get("ticket"))
	if ticketID == "" {
		return NewHTTPError(http.StatusBadRequest, "invalid_argument", "missing ticket", map[string]any{"field": "ticket"})
	}

	var resp ticketGetResponse
	if err := s.mgr.WithWorkspace(func(ws *workspace.Workspace) error {
		res, err := tickets.Resolve(r.Context(), ws, ticketID)
		if err != nil {
			if errors.Is(err, tickets.ErrNotFound) {
				return NewHTTPError(http.StatusNotFound, "not_found", "ticket not found", map[string]any{"field": "ticket", "value": ticketID})
			}
			if errors.Is(err, tickets.ErrAmbiguous) {
				return NewHTTPError(http.StatusBadRequest, "ambiguous", "ticket is ambiguous", map[string]any{"field": "ticket", "value": ticketID})
			}
			return err
		}

		docsTotal, relatedFilesTotal, err := ticketDocStats(r.Context(), ws, ticketID)
		if err != nil {
			return err
		}

		tasksTotal, tasksDone, _ := ticketTaskCounts(ws, res.TicketDirRel)

		updatedAt := ""
		if res.IndexDoc != nil && !res.IndexDoc.LastUpdated.IsZero() {
			updatedAt = res.IndexDoc.LastUpdated.Format(time.RFC3339Nano)
		} else if fi, err := os.Stat(res.IndexPathAbs); err == nil {
			updatedAt = fi.ModTime().Format(time.RFC3339Nano)
		}

		resp = ticketGetResponse{
			Ticket:    ticketID,
			Title:     strings.TrimSpace(res.IndexDoc.Title),
			Status:    strings.TrimSpace(res.IndexDoc.Status),
			Intent:    strings.TrimSpace(res.IndexDoc.Intent),
			Owners:    append([]string{}, res.IndexDoc.Owners...),
			Topics:    append([]string{}, res.IndexDoc.Topics...),
			CreatedAt: res.CreatedAt,
			UpdatedAt: updatedAt,
			TicketDir: res.TicketDirRel,
			IndexPath: res.IndexPathRel,
			Stats: ticketStats{
				DocsTotal:         docsTotal,
				TasksTotal:        tasksTotal,
				TasksDone:         tasksDone,
				RelatedFilesTotal: relatedFilesTotal,
			},
		}
		return nil
	}); err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, resp)
}

type ticketDocItem struct {
	Path         string               `json:"path"`
	Title        string               `json:"title"`
	DocType      string               `json:"docType"`
	Status       string               `json:"status"`
	Topics       []string             `json:"topics"`
	Summary      string               `json:"summary"`
	LastUpdated  *time.Time           `json:"lastUpdated,omitempty"`
	RelatedFiles []models.RelatedFile `json:"relatedFiles"`
}

type ticketDocsResponse struct {
	Ticket     string          `json:"ticket"`
	Total      int             `json:"total"`
	Results    []ticketDocItem `json:"results"`
	NextCursor string          `json:"nextCursor"`
}

func (s *Server) handleTicketsDocs(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return NewHTTPError(http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
	}

	ticketID := strings.TrimSpace(r.URL.Query().Get("ticket"))
	if ticketID == "" {
		return NewHTTPError(http.StatusBadRequest, "invalid_argument", "missing ticket", map[string]any{"field": "ticket"})
	}

	pageSize := parseIntDefault(r.URL.Query().Get("pageSize"), 200)
	if pageSize <= 0 {
		pageSize = 200
	}
	if pageSize > 1000 {
		pageSize = 1000
	}
	offset, err := decodeCursor(r.URL.Query().Get("cursor"))
	if err != nil {
		return NewHTTPError(http.StatusBadRequest, "invalid_cursor", err.Error(), nil)
	}

	orderByRaw := strings.TrimSpace(r.URL.Query().Get("orderBy"))
	orderBy := workspace.OrderBy(orderByRaw)
	if orderBy == "" {
		orderBy = workspace.OrderByPath
	}
	switch orderBy {
	case workspace.OrderByPath, workspace.OrderByLastUpdated:
	case workspace.OrderByRank:
		return NewHTTPError(http.StatusBadRequest, "invalid_argument", "invalid orderBy", map[string]any{
			"field": "orderBy",
			"value": orderByRaw,
		})
	default:
		return NewHTTPError(http.StatusBadRequest, "invalid_argument", "invalid orderBy", map[string]any{
			"field": "orderBy",
			"value": orderByRaw,
		})
	}

	includeArchived := parseBoolDefault(r.URL.Query().Get("includeArchived"), true)
	includeScripts := parseBoolDefault(r.URL.Query().Get("includeScripts"), true)
	includeControl := parseBoolDefault(r.URL.Query().Get("includeControlDocs"), true)

	var resp ticketDocsResponse
	if err := s.mgr.WithWorkspace(func(ws *workspace.Workspace) error {
		if _, err := tickets.Resolve(r.Context(), ws, ticketID); err != nil {
			if errors.Is(err, tickets.ErrNotFound) {
				return NewHTTPError(http.StatusNotFound, "not_found", "ticket not found", map[string]any{"field": "ticket", "value": ticketID})
			}
			if errors.Is(err, tickets.ErrAmbiguous) {
				return NewHTTPError(http.StatusBadRequest, "ambiguous", "ticket is ambiguous", map[string]any{"field": "ticket", "value": ticketID})
			}
			return err
		}

		qr, err := ws.QueryDocs(r.Context(), workspace.DocQuery{
			Scope: workspace.Scope{Kind: workspace.ScopeTicket, TicketID: ticketID},
			Options: workspace.DocQueryOptions{
				IncludeBody:         false,
				IncludeErrors:       false,
				IncludeDiagnostics:  false,
				IncludeArchivedPath: includeArchived,
				IncludeScriptsPath:  includeScripts,
				IncludeControlDocs:  includeControl,
				OrderBy:             orderBy,
			},
		})
		if err != nil {
			return err
		}

		docs := make([]workspace.DocHandle, 0, len(qr.Docs))
		for _, h := range qr.Docs {
			if h.Doc == nil {
				continue
			}
			docs = append(docs, h)
		}

		total := len(docs)
		if offset > total {
			offset = total
		}
		end := offset + pageSize
		if end > total {
			end = total
		}
		page := docs[offset:end]

		next := ""
		if end < total {
			next, err = encodeCursor(end)
			if err != nil {
				return err
			}
		}

		items := make([]ticketDocItem, 0, len(page))
		for _, h := range page {
			rel := h.Path
			if r, err := filepath.Rel(ws.Context().Root, h.Path); err == nil {
				rel = filepath.ToSlash(r)
			} else {
				rel = filepath.ToSlash(rel)
			}
			var lastUpdated *time.Time
			if !h.Doc.LastUpdated.IsZero() {
				t := h.Doc.LastUpdated
				lastUpdated = &t
			}
			items = append(items, ticketDocItem{
				Path:         rel,
				Title:        strings.TrimSpace(h.Doc.Title),
				DocType:      strings.TrimSpace(h.Doc.DocType),
				Status:       strings.TrimSpace(h.Doc.Status),
				Topics:       append([]string{}, h.Doc.Topics...),
				Summary:      strings.TrimSpace(h.Doc.Summary),
				LastUpdated:  lastUpdated,
				RelatedFiles: append([]models.RelatedFile{}, h.Doc.RelatedFiles...),
			})
		}

		resp = ticketDocsResponse{
			Ticket:     ticketID,
			Total:      total,
			Results:    items,
			NextCursor: next,
		}
		return nil
	}); err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, resp)
}

type ticketTasksStats struct {
	Total int `json:"total"`
	Done  int `json:"done"`
}

type ticketTasksResponse struct {
	Ticket    string            `json:"ticket"`
	Exists    bool              `json:"exists"`
	TasksPath string            `json:"tasksPath"`
	Stats     ticketTasksStats  `json:"stats"`
	Sections  []tasksmd.Section `json:"sections"`
}

func (s *Server) handleTicketsTasks(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return NewHTTPError(http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
	}

	ticketID := strings.TrimSpace(r.URL.Query().Get("ticket"))
	if ticketID == "" {
		return NewHTTPError(http.StatusBadRequest, "invalid_argument", "missing ticket", map[string]any{"field": "ticket"})
	}

	var resp ticketTasksResponse
	if err := s.mgr.WithWorkspace(func(ws *workspace.Workspace) error {
		res, err := tickets.Resolve(r.Context(), ws, ticketID)
		if err != nil {
			if errors.Is(err, tickets.ErrNotFound) {
				return NewHTTPError(http.StatusNotFound, "not_found", "ticket not found", map[string]any{"field": "ticket", "value": ticketID})
			}
			if errors.Is(err, tickets.ErrAmbiguous) {
				return NewHTTPError(http.StatusBadRequest, "ambiguous", "ticket is ambiguous", map[string]any{"field": "ticket", "value": ticketID})
			}
			return err
		}

		rawPath := filepath.ToSlash(filepath.Join(res.TicketDirRel, "tasks.md"))
		abs, rel, _, err := resolveFileWithin(ws.Context().Root, rawPath)
		if err != nil {
			var he *HTTPError
			if errors.As(err, &he) && he.Status == http.StatusNotFound {
				resp = ticketTasksResponse{
					Ticket:    ticketID,
					Exists:    false,
					TasksPath: rawPath,
					Stats:     ticketTasksStats{Total: 0, Done: 0},
					Sections:  []tasksmd.Section{},
				}
				return nil
			}
			return err
		}

		lines, err := tasksmd.ReadFile(abs)
		if err != nil {
			return err
		}
		parsed, _ := tasksmd.Parse(lines)
		resp = ticketTasksResponse{
			Ticket:    ticketID,
			Exists:    true,
			TasksPath: rel,
			Stats:     ticketTasksStats{Total: parsed.Total, Done: parsed.Done},
			Sections:  parsed.Sections,
		}
		return nil
	}); err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, resp)
}

type ticketTasksCheckRequest struct {
	Ticket  string `json:"ticket"`
	IDs     []int  `json:"ids"`
	Checked bool   `json:"checked"`
}

func (s *Server) handleTicketsTasksCheck(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return NewHTTPError(http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
	}

	var req ticketTasksCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return NewHTTPError(http.StatusBadRequest, "invalid_argument", "invalid json body", nil)
	}
	req.Ticket = strings.TrimSpace(req.Ticket)
	if req.Ticket == "" {
		return NewHTTPError(http.StatusBadRequest, "invalid_argument", "missing ticket", map[string]any{"field": "ticket"})
	}
	if len(req.IDs) == 0 {
		return NewHTTPError(http.StatusBadRequest, "invalid_argument", "missing ids", map[string]any{"field": "ids"})
	}

	if err := s.mgr.WithWorkspace(func(ws *workspace.Workspace) error {
		res, err := tickets.Resolve(r.Context(), ws, req.Ticket)
		if err != nil {
			if errors.Is(err, tickets.ErrNotFound) {
				return NewHTTPError(http.StatusNotFound, "not_found", "ticket not found", map[string]any{"field": "ticket", "value": req.Ticket})
			}
			if errors.Is(err, tickets.ErrAmbiguous) {
				return NewHTTPError(http.StatusBadRequest, "ambiguous", "ticket is ambiguous", map[string]any{"field": "ticket", "value": req.Ticket})
			}
			return err
		}

		rawPath := filepath.ToSlash(filepath.Join(res.TicketDirRel, "tasks.md"))
		abs, _, _, err := resolveFileWithin(ws.Context().Root, rawPath)
		if err != nil {
			return err
		}
		lines, err := tasksmd.ReadFile(abs)
		if err != nil {
			return err
		}
		updated, err := tasksmd.ToggleChecked(lines, req.IDs, req.Checked)
		if err != nil {
			return NewHTTPError(http.StatusBadRequest, "invalid_argument", err.Error(), nil)
		}
		if err := tasksmd.WriteFile(abs, updated); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

type ticketTasksAddRequest struct {
	Ticket  string `json:"ticket"`
	Section string `json:"section"`
	Text    string `json:"text"`
}

func (s *Server) handleTicketsTasksAdd(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return NewHTTPError(http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
	}

	var req ticketTasksAddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return NewHTTPError(http.StatusBadRequest, "invalid_argument", "invalid json body", nil)
	}
	req.Ticket = strings.TrimSpace(req.Ticket)
	req.Section = strings.TrimSpace(req.Section)
	req.Text = strings.TrimSpace(req.Text)
	if req.Ticket == "" {
		return NewHTTPError(http.StatusBadRequest, "invalid_argument", "missing ticket", map[string]any{"field": "ticket"})
	}
	if req.Text == "" {
		return NewHTTPError(http.StatusBadRequest, "invalid_argument", "missing text", map[string]any{"field": "text"})
	}

	if err := s.mgr.WithWorkspace(func(ws *workspace.Workspace) error {
		res, err := tickets.Resolve(r.Context(), ws, req.Ticket)
		if err != nil {
			if errors.Is(err, tickets.ErrNotFound) {
				return NewHTTPError(http.StatusNotFound, "not_found", "ticket not found", map[string]any{"field": "ticket", "value": req.Ticket})
			}
			if errors.Is(err, tickets.ErrAmbiguous) {
				return NewHTTPError(http.StatusBadRequest, "ambiguous", "ticket is ambiguous", map[string]any{"field": "ticket", "value": req.Ticket})
			}
			return err
		}

		rawPath := filepath.ToSlash(filepath.Join(res.TicketDirRel, "tasks.md"))
		abs, _, _, err := resolveFileWithin(ws.Context().Root, rawPath)
		if err != nil {
			return err
		}
		lines, err := tasksmd.ReadFile(abs)
		if err != nil {
			return err
		}
		updated, err := tasksmd.AppendTask(lines, req.Section, req.Text)
		if err != nil {
			return NewHTTPError(http.StatusBadRequest, "invalid_argument", err.Error(), nil)
		}
		if err := tasksmd.WriteFile(abs, updated); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

type ticketGraphResponse struct {
	Ticket    string            `json:"ticket"`
	Direction string            `json:"direction"`
	Mermaid   string            `json:"mermaid"`
	Stats     ticketgraph.Stats `json:"stats"`
}

func (s *Server) handleTicketsGraph(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return NewHTTPError(http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
	}

	ticketID := strings.TrimSpace(r.URL.Query().Get("ticket"))
	if ticketID == "" {
		return NewHTTPError(http.StatusBadRequest, "invalid_argument", "missing ticket", map[string]any{"field": "ticket"})
	}
	direction := strings.TrimSpace(r.URL.Query().Get("direction"))
	if direction == "" {
		direction = "TD"
	}

	includeArchived := parseBoolDefault(r.URL.Query().Get("includeArchived"), false)
	includeScripts := parseBoolDefault(r.URL.Query().Get("includeScripts"), false)
	includeControl := parseBoolDefault(r.URL.Query().Get("includeControlDocs"), true)

	var resp ticketGraphResponse
	if err := s.mgr.WithWorkspace(func(ws *workspace.Workspace) error {
		if _, err := tickets.Resolve(r.Context(), ws, ticketID); err != nil {
			if errors.Is(err, tickets.ErrNotFound) {
				return NewHTTPError(http.StatusNotFound, "not_found", "ticket not found", map[string]any{"field": "ticket", "value": ticketID})
			}
			if errors.Is(err, tickets.ErrAmbiguous) {
				return NewHTTPError(http.StatusBadRequest, "ambiguous", "ticket is ambiguous", map[string]any{"field": "ticket", "value": ticketID})
			}
			return err
		}

		mermaid, stats, err := ticketgraph.BuildMermaid(r.Context(), ws, ticketID, direction, includeArchived, includeScripts, includeControl)
		if err != nil {
			return NewHTTPError(http.StatusBadRequest, "invalid_argument", err.Error(), nil)
		}
		resp = ticketGraphResponse{
			Ticket:    ticketID,
			Direction: direction,
			Mermaid:   mermaid,
			Stats:     stats,
		}
		return nil
	}); err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, resp)
}

func ticketTaskCounts(ws *workspace.Workspace, ticketDirRel string) (int, int, error) {
	if ws == nil {
		return 0, 0, errors.New("nil workspace")
	}
	rawPath := filepath.ToSlash(filepath.Join(ticketDirRel, "tasks.md"))
	abs, _, _, err := resolveFileWithin(ws.Context().Root, rawPath)
	if err != nil {
		var he *HTTPError
		if errors.As(err, &he) && he.Status == http.StatusNotFound {
			return 0, 0, nil
		}
		return 0, 0, err
	}
	lines, err := tasksmd.ReadFile(abs)
	if err != nil {
		return 0, 0, err
	}
	p, _ := tasksmd.Parse(lines)
	return p.Total, p.Done, nil
}

func ticketDocStats(ctx context.Context, ws *workspace.Workspace, ticketID string) (int, int, error) {
	if ws == nil {
		return 0, 0, errors.New("nil workspace")
	}
	qr, err := ws.QueryDocs(ctx, workspace.DocQuery{
		Scope: workspace.Scope{Kind: workspace.ScopeTicket, TicketID: ticketID},
		Options: workspace.DocQueryOptions{
			IncludeBody:         false,
			IncludeErrors:       false,
			IncludeDiagnostics:  false,
			IncludeArchivedPath: true,
			IncludeScriptsPath:  true,
			IncludeControlDocs:  true,
			OrderBy:             workspace.OrderByPath,
		},
	})
	if err != nil {
		return 0, 0, err
	}

	fileSet := map[string]struct{}{}
	docsTotal := 0
	for _, h := range qr.Docs {
		if h.Doc == nil {
			continue
		}
		docsTotal++
		docResolver := paths.NewResolver(paths.ResolverOptions{
			DocsRoot:  ws.Context().Root,
			ConfigDir: ws.Context().ConfigDir,
			RepoRoot:  ws.Context().RepoRoot,
			DocPath:   h.Path,
		})
		for _, rf := range h.Doc.RelatedFiles {
			key := canonicalizeForStats(docResolver, rf.Path)
			if strings.TrimSpace(key) == "" {
				continue
			}
			fileSet[key] = struct{}{}
		}
	}
	return docsTotal, len(fileSet), nil
}

func canonicalizeForStats(resolver *paths.Resolver, raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || resolver == nil {
		return ""
	}
	n := resolver.NormalizeNoFS(raw)
	switch {
	case strings.TrimSpace(n.Canonical) != "":
		return filepath.ToSlash(strings.TrimSpace(n.Canonical))
	case strings.TrimSpace(n.RepoRelative) != "":
		return filepath.ToSlash(strings.TrimSpace(n.RepoRelative))
	case strings.TrimSpace(n.OriginalClean) != "":
		return filepath.ToSlash(strings.TrimSpace(n.OriginalClean))
	default:
		return ""
	}
}
