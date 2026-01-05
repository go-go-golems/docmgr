package httpapi

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-go-golems/docmgr/internal/searchsvc"
	"github.com/go-go-golems/docmgr/internal/workspace"
)

var ErrIndexNotReady = errors.New("index not ready; call /api/v1/index/refresh or restart with successful startup indexing")

type ServerOptions struct {
	CORSOrigin string
}

type Server struct {
	mgr  *IndexManager
	opts ServerOptions
	mux  *http.ServeMux
}

func NewServer(mgr *IndexManager, opts ServerOptions) *Server {
	s := &Server{
		mgr:  mgr,
		opts: opts,
		mux:  http.NewServeMux(),
	}

	s.mux.HandleFunc("/api/v1/healthz", s.wrap(s.handleHealthz))
	s.mux.HandleFunc("/api/v1/workspace/status", s.wrap(s.handleWorkspaceStatus))
	s.mux.HandleFunc("/api/v1/index/refresh", s.wrap(s.handleIndexRefresh))
	s.mux.HandleFunc("/api/v1/search/docs", s.wrap(s.handleSearchDocs))
	s.mux.HandleFunc("/api/v1/search/files", s.wrap(s.handleSearchFiles))
	s.mux.HandleFunc("/api/v1/docs/get", s.wrap(s.handleDocsGet))
	s.mux.HandleFunc("/api/v1/files/get", s.wrap(s.handleFilesGet))
	s.mux.HandleFunc("/api/v1/tickets/get", s.wrap(s.handleTicketsGet))
	s.mux.HandleFunc("/api/v1/tickets/docs", s.wrap(s.handleTicketsDocs))
	s.mux.HandleFunc("/api/v1/tickets/tasks", s.wrap(s.handleTicketsTasks))
	s.mux.HandleFunc("/api/v1/tickets/tasks/check", s.wrap(s.handleTicketsTasksCheck))
	s.mux.HandleFunc("/api/v1/tickets/tasks/add", s.wrap(s.handleTicketsTasksAdd))
	s.mux.HandleFunc("/api/v1/tickets/graph", s.wrap(s.handleTicketsGraph))

	return s
}

func (s *Server) Handler() http.Handler { return s.mux }

func (s *Server) wrap(fn func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.opts.CORSOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", s.opts.CORSOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}

		if err := fn(w, r); err != nil {
			s.writeError(w, r, err)
		}
	}
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return NewHTTPError(http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
	}
	return writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleWorkspaceStatus(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return NewHTTPError(http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
	}

	snap := s.mgr.Snapshot()
	if snap.Workspace == nil {
		return NewHTTPError(http.StatusServiceUnavailable, "index_not_ready", "index not ready", nil)
	}

	ctx := snap.Workspace.Context()
	cfgPath, _ := workspace.FindTTMPConfigPath()
	vocabPath := filepath.Join(ctx.Root, "vocabulary.yaml")

	return writeJSON(w, http.StatusOK, map[string]any{
		"root":           ctx.Root,
		"configDir":      ctx.ConfigDir,
		"repoRoot":       ctx.RepoRoot,
		"configPath":     cfgPath,
		"vocabularyPath": vocabPath,
		"indexedAt":      snap.IndexedAt.Format(time.RFC3339Nano),
		"docsIndexed":    snap.DocsIndexed,
		"ftsAvailable":   snap.Workspace.FTSAvailable(),
	})
}

func (s *Server) handleIndexRefresh(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return NewHTTPError(http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
	}

	snap, err := s.mgr.Refresh(r.Context())
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, map[string]any{
		"refreshed":    true,
		"indexedAt":    snap.IndexedAt.Format(time.RFC3339Nano),
		"docsIndexed":  snap.DocsIndexed,
		"ftsAvailable": snap.Workspace.FTSAvailable(),
	})
}

type cursorPayload struct {
	V int `json:"v"`
	O int `json:"o"`
}

func encodeCursor(offset int) (string, error) {
	if offset <= 0 {
		return "", nil
	}
	b, err := json.Marshal(cursorPayload{V: 1, O: offset})
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func decodeCursor(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid cursor: %w", err)
	}
	var p cursorPayload
	if err := json.Unmarshal(b, &p); err != nil {
		return 0, fmt.Errorf("invalid cursor: %w", err)
	}
	if p.V != 1 {
		return 0, fmt.Errorf("invalid cursor version: %d", p.V)
	}
	if p.O < 0 {
		return 0, fmt.Errorf("invalid cursor offset: %d", p.O)
	}
	return p.O, nil
}

func (s *Server) handleSearchDocs(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return NewHTTPError(http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
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
	case workspace.OrderByPath, workspace.OrderByLastUpdated, workspace.OrderByRank:
	default:
		return NewHTTPError(http.StatusBadRequest, "invalid_argument", "invalid orderBy", map[string]any{
			"field": "orderBy",
			"value": orderByRaw,
		})
	}

	reverse := parseBoolDefault(r.URL.Query().Get("reverse"), false)

	q := searchsvc.SearchQuery{
		TextQuery:           strings.TrimSpace(r.URL.Query().Get("query")),
		AllowEmpty:          true,
		Ticket:              strings.TrimSpace(r.URL.Query().Get("ticket")),
		Topics:              splitCSV(r.URL.Query().Get("topics")),
		DocType:             strings.TrimSpace(r.URL.Query().Get("docType")),
		Status:              strings.TrimSpace(r.URL.Query().Get("status")),
		File:                strings.TrimSpace(r.URL.Query().Get("file")),
		Dir:                 strings.TrimSpace(r.URL.Query().Get("dir")),
		ExternalSource:      strings.TrimSpace(r.URL.Query().Get("externalSource")),
		Since:               strings.TrimSpace(r.URL.Query().Get("since")),
		Until:               strings.TrimSpace(r.URL.Query().Get("until")),
		CreatedSince:        strings.TrimSpace(r.URL.Query().Get("createdSince")),
		UpdatedSince:        strings.TrimSpace(r.URL.Query().Get("updatedSince")),
		OrderBy:             orderBy,
		Reverse:             reverse,
		IncludeArchivedPath: parseBoolDefault(r.URL.Query().Get("includeArchived"), true),
		IncludeScriptsPath:  parseBoolDefault(r.URL.Query().Get("includeScripts"), true),
		IncludeControlDocs:  parseBoolDefault(r.URL.Query().Get("includeControlDocs"), true),
		IncludeDiagnostics:  parseBoolDefault(r.URL.Query().Get("includeDiagnostics"), true),
		IncludeErrors:       parseBoolDefault(r.URL.Query().Get("includeErrors"), false),
	}

	// Reverse lookup requires file/dir. As a convenience for UIs, treat `query` as `file`
	// when reverse=true and file/dir are empty.
	if q.Reverse && strings.TrimSpace(q.File) == "" && strings.TrimSpace(q.Dir) == "" {
		if strings.TrimSpace(q.TextQuery) != "" {
			q.File = strings.TrimSpace(q.TextQuery)
			q.TextQuery = ""
		} else {
			return NewHTTPError(http.StatusBadRequest, "invalid_argument", "reverse search requires file or dir", map[string]any{
				"field": "reverse",
			})
		}
	}

	// Rank ordering only makes sense with a non-empty text query; fall back to a stable default.
	if q.OrderBy == workspace.OrderByRank && strings.TrimSpace(q.TextQuery) == "" {
		q.OrderBy = workspace.OrderByPath
	}

	var resp searchsvc.SearchResponse
	if err := s.mgr.WithWorkspace(func(ws *workspace.Workspace) error {
		var err error
		resp, err = searchsvc.SearchDocs(r.Context(), ws, q)
		return err
	}); err != nil {
		if errors.Is(err, workspace.ErrFTSNotAvailable) {
			return NewHTTPError(http.StatusBadRequest, "fts_not_available", err.Error(), nil)
		}
		return err
	}

	total := resp.Total
	if offset > total {
		offset = total
	}
	end := offset + pageSize
	if end > total {
		end = total
	}
	page := resp.Results[offset:end]

	next := ""
	if end < total {
		next, err = encodeCursor(end)
		if err != nil {
			return err
		}
	}

	return writeJSON(w, http.StatusOK, map[string]any{
		"query": map[string]any{
			"query":          q.TextQuery,
			"ticket":         q.Ticket,
			"topics":         q.Topics,
			"docType":        q.DocType,
			"status":         q.Status,
			"file":           q.File,
			"dir":            q.Dir,
			"externalSource": q.ExternalSource,
			"since":          q.Since,
			"until":          q.Until,
			"createdSince":   q.CreatedSince,
			"updatedSince":   q.UpdatedSince,
			"orderBy":        string(q.OrderBy),
			"reverse":        q.Reverse,
			"pageSize":       pageSize,
			"cursor":         r.URL.Query().Get("cursor"),
		},
		"total":       total,
		"results":     page,
		"diagnostics": resp.Diagnostics,
		"nextCursor":  next,
	})
}

type fileSuggestion struct {
	File   string `json:"file"`
	Source string `json:"source"`
	Reason string `json:"reason"`
}

func (s *Server) handleSearchFiles(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return NewHTTPError(http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
	}

	limit := parseIntDefault(r.URL.Query().Get("limit"), 200)
	if limit <= 0 {
		limit = 200
	}
	if limit > 1000 {
		limit = 1000
	}

	q := searchsvc.SuggestFilesQuery{
		Ticket: strings.TrimSpace(r.URL.Query().Get("ticket")),
		Topics: splitCSV(r.URL.Query().Get("topics")),
		Query:  strings.TrimSpace(r.URL.Query().Get("query")),
	}

	var suggestions []searchsvc.FileSuggestion
	if err := s.mgr.WithWorkspace(func(ws *workspace.Workspace) error {
		var err error
		suggestions, err = searchsvc.SuggestFiles(r.Context(), ws, q)
		return err
	}); err != nil {
		return err
	}

	if len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}

	out := make([]fileSuggestion, 0, len(suggestions))
	for _, s := range suggestions {
		out = append(out, fileSuggestion{File: s.File, Source: s.Source, Reason: s.Reason})
	}

	return writeJSON(w, http.StatusOK, map[string]any{
		"total":   len(out),
		"results": out,
	})
}

func (s *Server) writeError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, ErrIndexNotReady) {
		_ = writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"error": map[string]any{
				"code":    "index_not_ready",
				"message": "index not ready",
			},
		})
		return
	}

	var he *HTTPError
	if errors.As(err, &he) {
		_ = writeJSON(w, he.Status, map[string]any{
			"error": map[string]any{
				"code":    he.Code,
				"message": he.Message,
				"details": he.Details,
			},
		})
		return
	}
	_ = writeJSON(w, http.StatusInternalServerError, map[string]any{
		"error": map[string]any{
			"code":    "internal",
			"message": err.Error(),
		},
	})
}

func parseBoolDefault(s string, def bool) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return def
	}
	switch s {
	case "1", "true", "t", "yes", "y", "on":
		return true
	case "0", "false", "f", "no", "n", "off":
		return false
	default:
		return def
	}
}

func parseIntDefault(s string, def int) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func splitCSV(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

type HTTPError struct {
	Status  int
	Code    string
	Message string
	Details any
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewHTTPError(status int, code, message string, details any) *HTTPError {
	return &HTTPError{Status: status, Code: code, Message: message, Details: details}
}
