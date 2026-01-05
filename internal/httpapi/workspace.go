package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/go-go-golems/docmgr/internal/searchsvc"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/docmgr/pkg/commands"
	"github.com/go-go-golems/docmgr/pkg/models"
)

type workspaceSummaryStats struct {
	TicketsTotal    int `json:"ticketsTotal"`
	TicketsActive   int `json:"ticketsActive"`
	TicketsComplete int `json:"ticketsComplete"`
	TicketsReview   int `json:"ticketsReview"`
	TicketsDraft    int `json:"ticketsDraft"`
}

type ticketListItemStats struct {
	DocsTotal         int `json:"docsTotal"`
	TasksTotal        int `json:"tasksTotal"`
	TasksDone         int `json:"tasksDone"`
	RelatedFilesTotal int `json:"relatedFilesTotal"`
}

type ticketListItem struct {
	Ticket    string   `json:"ticket"`
	Title     string   `json:"title"`
	Status    string   `json:"status"`
	Topics    []string `json:"topics"`
	Owners    []string `json:"owners"`
	Intent    string   `json:"intent"`
	CreatedAt string   `json:"createdAt"`
	UpdatedAt string   `json:"updatedAt"`

	TicketDir string `json:"ticketDir"`
	IndexPath string `json:"indexPath"`

	Snippet string               `json:"snippet"`
	Stats   *ticketListItemStats `json:"stats"`
}

type workspaceRecentDocItem struct {
	Path      string   `json:"path"`
	Ticket    string   `json:"ticket"`
	Title     string   `json:"title"`
	DocType   string   `json:"docType"`
	Status    string   `json:"status"`
	Topics    []string `json:"topics"`
	UpdatedAt string   `json:"updatedAt"`
}

type workspaceSummaryResponse struct {
	Root        string                `json:"root"`
	RepoRoot    string                `json:"repoRoot"`
	IndexedAt   string                `json:"indexedAt"`
	DocsIndexed int                   `json:"docsIndexed"`
	Stats       workspaceSummaryStats `json:"stats"`
	Recent      struct {
		Tickets []ticketListItem         `json:"tickets"`
		Docs    []workspaceRecentDocItem `json:"docs"`
	} `json:"recent"`
}

func (s *Server) handleWorkspaceSummary(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return NewHTTPError(http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
	}

	snap := s.mgr.Snapshot()
	if snap.Workspace == nil {
		return NewHTTPError(http.StatusServiceUnavailable, "index_not_ready", "index not ready", nil)
	}

	var resp workspaceSummaryResponse
	if err := s.mgr.WithWorkspace(func(ws *workspace.Workspace) error {
		tickets, stats, err := listTicketIndexDocs(r.Context(), ws, workspaceTicketsQuery{
			IncludeArchived: true,
			IncludeStats:    false,
			OrderBy:         "last_updated",
			Reverse:         true,
			PageSize:        1000,
		})
		if err != nil {
			return err
		}

		recentDocs, err := listRecentDocs(r.Context(), ws, 10, true)
		if err != nil {
			return err
		}

		resp = workspaceSummaryResponse{
			Root:        ws.Context().Root,
			RepoRoot:    ws.Context().RepoRoot,
			IndexedAt:   snap.IndexedAt.Format(time.RFC3339Nano),
			DocsIndexed: snap.DocsIndexed,
			Stats:       stats,
		}
		if len(tickets) > 10 {
			tickets = tickets[:10]
		}
		resp.Recent.Tickets = tickets
		resp.Recent.Docs = recentDocs
		return nil
	}); err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, resp)
}

type workspaceTicketsQuery struct {
	Q               string   `json:"q"`
	Status          string   `json:"status"`
	Ticket          string   `json:"ticket"`
	Topics          []string `json:"topics"`
	Owners          []string `json:"owners"`
	Intent          string   `json:"intent"`
	OrderBy         string   `json:"orderBy"`
	Reverse         bool     `json:"reverse"`
	IncludeArchived bool     `json:"includeArchived"`
	IncludeStats    bool     `json:"includeStats"`
	PageSize        int      `json:"pageSize"`
	Cursor          string   `json:"cursor"`
}

type workspaceTicketsResponse struct {
	Query      workspaceTicketsQuery `json:"query"`
	Total      int                   `json:"total"`
	Results    []ticketListItem      `json:"results"`
	NextCursor string                `json:"nextCursor"`
}

func (s *Server) handleWorkspaceTickets(w http.ResponseWriter, r *http.Request) error {
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

	q := workspaceTicketsQuery{
		Q:               strings.TrimSpace(r.URL.Query().Get("q")),
		Status:          strings.TrimSpace(r.URL.Query().Get("status")),
		Ticket:          strings.TrimSpace(r.URL.Query().Get("ticket")),
		Topics:          splitCSV(r.URL.Query().Get("topics")),
		Owners:          splitCSV(r.URL.Query().Get("owners")),
		Intent:          strings.TrimSpace(r.URL.Query().Get("intent")),
		OrderBy:         strings.TrimSpace(r.URL.Query().Get("orderBy")),
		Reverse:         parseBoolDefault(r.URL.Query().Get("reverse"), false),
		IncludeArchived: parseBoolDefault(r.URL.Query().Get("includeArchived"), true),
		IncludeStats:    parseBoolDefault(r.URL.Query().Get("includeStats"), false),
		PageSize:        pageSize,
		Cursor:          r.URL.Query().Get("cursor"),
	}
	if q.OrderBy == "" {
		q.OrderBy = "last_updated"
	}
	switch q.OrderBy {
	case "last_updated", "ticket", "title":
	default:
		return NewHTTPError(http.StatusBadRequest, "invalid_argument", "invalid orderBy", map[string]any{
			"field": "orderBy",
			"value": q.OrderBy,
		})
	}

	var resp workspaceTicketsResponse
	if err := s.mgr.WithWorkspace(func(ws *workspace.Workspace) error {
		all, _, err := listTicketIndexDocs(r.Context(), ws, q)
		if err != nil {
			if errors.Is(err, workspace.ErrFTSNotAvailable) {
				return NewHTTPError(http.StatusBadRequest, "fts_not_available", err.Error(), nil)
			}
			return err
		}

		total := len(all)
		if offset > total {
			offset = total
		}
		end := offset + pageSize
		if end > total {
			end = total
		}
		page := all[offset:end]

		next := ""
		if end < total {
			next, err = encodeCursor(end)
			if err != nil {
				return err
			}
		}

		resp = workspaceTicketsResponse{
			Query:      q,
			Total:      total,
			Results:    page,
			NextCursor: next,
		}
		return nil
	}); err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, resp)
}

type workspaceFacetsResponse struct {
	Statuses []string `json:"statuses"`
	DocTypes []string `json:"docTypes"`
	Intents  []string `json:"intents"`
	Topics   []string `json:"topics"`
	Owners   []string `json:"owners"`
}

func (s *Server) handleWorkspaceFacets(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return NewHTTPError(http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
	}

	includeArchived := parseBoolDefault(r.URL.Query().Get("includeArchived"), true)

	var resp workspaceFacetsResponse
	if err := s.mgr.WithWorkspace(func(ws *workspace.Workspace) error {
		vocab, _ := commands.LoadVocabulary()

		// Prefer vocabulary if available.
		resp.Statuses = vocabSlugs(vocab.Status)
		resp.DocTypes = vocabSlugs(vocab.DocTypes)
		resp.Intents = vocabSlugs(vocab.Intent)
		resp.Topics = vocabSlugs(vocab.Topics)

		db := ws.DB()
		if db == nil {
			return errors.New("workspace db is nil")
		}

		if len(resp.Statuses) == 0 {
			resp.Statuses = distinctDocColumn(r.Context(), db, "status", includeArchived)
		}
		if len(resp.DocTypes) == 0 {
			resp.DocTypes = distinctDocColumn(r.Context(), db, "doc_type", includeArchived)
		}
		if len(resp.Intents) == 0 {
			resp.Intents = distinctDocColumn(r.Context(), db, "intent", includeArchived)
		}
		if len(resp.Topics) == 0 {
			resp.Topics = distinctTopics(r.Context(), db, includeArchived)
		}
		resp.Owners = distinctOwners(r.Context(), db, includeArchived)

		sort.Strings(resp.Statuses)
		sort.Strings(resp.DocTypes)
		sort.Strings(resp.Intents)
		sort.Strings(resp.Topics)
		sort.Strings(resp.Owners)

		return nil
	}); err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, resp)
}

type workspaceRecentResponse struct {
	Tickets []ticketListItem         `json:"tickets"`
	Docs    []workspaceRecentDocItem `json:"docs"`
}

func (s *Server) handleWorkspaceRecent(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return NewHTTPError(http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
	}

	ticketsLimit := parseIntDefault(r.URL.Query().Get("ticketsLimit"), 20)
	if ticketsLimit <= 0 {
		ticketsLimit = 20
	}
	if ticketsLimit > 1000 {
		ticketsLimit = 1000
	}
	docsLimit := parseIntDefault(r.URL.Query().Get("docsLimit"), 20)
	if docsLimit <= 0 {
		docsLimit = 20
	}
	if docsLimit > 1000 {
		docsLimit = 1000
	}
	includeArchived := parseBoolDefault(r.URL.Query().Get("includeArchived"), true)

	var resp workspaceRecentResponse
	if err := s.mgr.WithWorkspace(func(ws *workspace.Workspace) error {
		tickets, _, err := listTicketIndexDocs(r.Context(), ws, workspaceTicketsQuery{
			IncludeArchived: includeArchived,
			OrderBy:         "last_updated",
			Reverse:         true,
			PageSize:        1000,
		})
		if err != nil {
			return err
		}
		if len(tickets) > ticketsLimit {
			tickets = tickets[:ticketsLimit]
		}

		docs, err := listRecentDocs(r.Context(), ws, docsLimit, includeArchived)
		if err != nil {
			return err
		}

		resp = workspaceRecentResponse{
			Tickets: tickets,
			Docs:    docs,
		}
		return nil
	}); err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, resp)
}

type workspaceTopicListItem struct {
	Topic        string `json:"topic"`
	DocsTotal    int    `json:"docsTotal"`
	TicketsTotal int    `json:"ticketsTotal"`
	UpdatedAt    string `json:"updatedAt"`
}

type workspaceTopicsResponse struct {
	Total   int                      `json:"total"`
	Results []workspaceTopicListItem `json:"results"`
}

func (s *Server) handleWorkspaceTopics(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return NewHTTPError(http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
	}
	includeArchived := parseBoolDefault(r.URL.Query().Get("includeArchived"), true)

	var resp workspaceTopicsResponse
	if err := s.mgr.WithWorkspace(func(ws *workspace.Workspace) error {
		db := ws.DB()
		if db == nil {
			return errors.New("workspace db is nil")
		}

		where := "d.parse_ok = 1"
		if !includeArchived {
			where += " AND d.is_archived_path = 0"
		}
		// #nosec G202 -- statement is static.
		rows, err := db.QueryContext(r.Context(), `
SELECT
  t.topic_lower,
  MIN(t.topic_original) AS topic,
  COUNT(1) AS docs_total,
  COUNT(DISTINCT d.ticket_id) AS tickets_total,
  MAX(COALESCE(d.last_updated,'')) AS max_last_updated
FROM doc_topics t
JOIN docs d ON d.doc_id = t.doc_id
WHERE `+where+`
GROUP BY t.topic_lower
ORDER BY tickets_total DESC, docs_total DESC, topic_lower ASC;
`)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()

		var out []workspaceTopicListItem
		for rows.Next() {
			var (
				_topicLower    string
				topic          string
				docsTotal      int
				ticketsTotal   int
				maxLastUpdated string
			)
			if err := rows.Scan(&_topicLower, &topic, &docsTotal, &ticketsTotal, &maxLastUpdated); err != nil {
				return err
			}
			out = append(out, workspaceTopicListItem{
				Topic:        strings.TrimSpace(topic),
				DocsTotal:    docsTotal,
				TicketsTotal: ticketsTotal,
				UpdatedAt:    strings.TrimSpace(maxLastUpdated),
			})
		}
		if err := rows.Err(); err != nil {
			return err
		}

		resp = workspaceTopicsResponse{Total: len(out), Results: out}
		return nil
	}); err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, resp)
}

type workspaceTopicDetailResponse struct {
	Topic   string                   `json:"topic"`
	Stats   workspaceSummaryStats    `json:"stats"`
	Tickets []ticketListItem         `json:"tickets"`
	Docs    []workspaceRecentDocItem `json:"docs"`
}

func (s *Server) handleWorkspaceTopicsGet(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return NewHTTPError(http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
	}

	topic := strings.TrimSpace(r.URL.Query().Get("topic"))
	if topic == "" {
		return NewHTTPError(http.StatusBadRequest, "invalid_argument", "missing topic", map[string]any{"field": "topic"})
	}
	includeArchived := parseBoolDefault(r.URL.Query().Get("includeArchived"), true)
	docsLimit := parseIntDefault(r.URL.Query().Get("docsLimit"), 20)
	if docsLimit <= 0 {
		docsLimit = 20
	}
	if docsLimit > 1000 {
		docsLimit = 1000
	}

	var resp workspaceTopicDetailResponse
	if err := s.mgr.WithWorkspace(func(ws *workspace.Workspace) error {
		tickets, stats, err := listTicketIndexDocs(r.Context(), ws, workspaceTicketsQuery{
			Topics:          []string{topic},
			IncludeArchived: includeArchived,
			OrderBy:         "ticket",
			Reverse:         false,
			PageSize:        1000,
		})
		if err != nil {
			return err
		}

		docs, err := listDocsByTopic(r.Context(), ws, topic, docsLimit, includeArchived)
		if err != nil {
			return err
		}

		resp = workspaceTopicDetailResponse{
			Topic:   topic,
			Stats:   stats,
			Tickets: tickets,
			Docs:    docs,
		}
		return nil
	}); err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, resp)
}

func listDocsByTopic(ctx context.Context, ws *workspace.Workspace, topic string, limit int, includeArchived bool) ([]workspaceRecentDocItem, error) {
	res, err := ws.QueryDocs(ctx, workspace.DocQuery{
		Scope: workspace.Scope{Kind: workspace.ScopeRepo},
		Filters: workspace.DocFilters{
			TopicsAny: []string{topic},
		},
		Options: workspace.DocQueryOptions{
			IncludeBody:         false,
			IncludeErrors:       false,
			IncludeDiagnostics:  false,
			IncludeArchivedPath: includeArchived,
			IncludeScriptsPath:  true,
			IncludeControlDocs:  false,
			OrderBy:             workspace.OrderByLastUpdated,
			Reverse:             true,
		},
	})
	if err != nil {
		return nil, err
	}

	items := make([]workspaceRecentDocItem, 0, len(res.Docs))
	for _, h := range res.Docs {
		if h.Doc == nil {
			continue
		}
		rel := relPath(ws, h.Path)
		updatedAt := docUpdatedAt(ws, h.Path, h.Doc)
		items = append(items, workspaceRecentDocItem{
			Path:      rel,
			Ticket:    strings.TrimSpace(h.Doc.Ticket),
			Title:     strings.TrimSpace(h.Doc.Title),
			DocType:   strings.TrimSpace(h.Doc.DocType),
			Status:    strings.TrimSpace(h.Doc.Status),
			Topics:    append([]string{}, h.Doc.Topics...),
			UpdatedAt: updatedAt,
		})
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].UpdatedAt == items[j].UpdatedAt {
			return items[i].Path < items[j].Path
		}
		return items[i].UpdatedAt > items[j].UpdatedAt
	})
	if len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func listRecentDocs(ctx context.Context, ws *workspace.Workspace, limit int, includeArchived bool) ([]workspaceRecentDocItem, error) {
	res, err := ws.QueryDocs(ctx, workspace.DocQuery{
		Scope:   workspace.Scope{Kind: workspace.ScopeRepo},
		Filters: workspace.DocFilters{},
		Options: workspace.DocQueryOptions{
			IncludeBody:         false,
			IncludeErrors:       false,
			IncludeDiagnostics:  false,
			IncludeArchivedPath: includeArchived,
			IncludeScriptsPath:  true,
			IncludeControlDocs:  false,
			OrderBy:             workspace.OrderByLastUpdated,
			Reverse:             true,
		},
	})
	if err != nil {
		return nil, err
	}
	items := make([]workspaceRecentDocItem, 0, len(res.Docs))
	for _, h := range res.Docs {
		if h.Doc == nil {
			continue
		}
		rel := relPath(ws, h.Path)
		updatedAt := docUpdatedAt(ws, h.Path, h.Doc)
		items = append(items, workspaceRecentDocItem{
			Path:      rel,
			Ticket:    strings.TrimSpace(h.Doc.Ticket),
			Title:     strings.TrimSpace(h.Doc.Title),
			DocType:   strings.TrimSpace(h.Doc.DocType),
			Status:    strings.TrimSpace(h.Doc.Status),
			Topics:    append([]string{}, h.Doc.Topics...),
			UpdatedAt: updatedAt,
		})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].UpdatedAt == items[j].UpdatedAt {
			return items[i].Path < items[j].Path
		}
		return items[i].UpdatedAt > items[j].UpdatedAt
	})
	if len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func listTicketIndexDocs(ctx context.Context, ws *workspace.Workspace, q workspaceTicketsQuery) ([]ticketListItem, workspaceSummaryStats, error) {
	if ws == nil {
		return nil, workspaceSummaryStats{}, errors.New("nil workspace")
	}
	if strings.TrimSpace(q.Q) != "" && !ws.FTSAvailable() {
		return nil, workspaceSummaryStats{}, workspace.ErrFTSNotAvailable
	}

	docQ := workspace.DocQuery{
		Scope: workspace.Scope{Kind: workspace.ScopeRepo},
		Filters: workspace.DocFilters{
			DocType:   "index",
			Status:    strings.TrimSpace(q.Status),
			Ticket:    strings.TrimSpace(q.Ticket),
			TopicsAny: q.Topics,
			OwnersAny: q.Owners,
			Intent:    strings.TrimSpace(q.Intent),
			TextQuery: strings.TrimSpace(q.Q),
		},
		Options: workspace.DocQueryOptions{
			IncludeBody:         strings.TrimSpace(q.Q) != "",
			IncludeErrors:       false,
			IncludeDiagnostics:  false,
			IncludeArchivedPath: q.IncludeArchived,
			IncludeScriptsPath:  true,
			IncludeControlDocs:  true,
			OrderBy:             workspace.OrderByLastUpdated,
			Reverse:             q.Reverse,
		},
	}

	if q.OrderBy == "ticket" || q.OrderBy == "title" {
		docQ.Options.OrderBy = workspace.OrderByPath
	}
	res, err := ws.QueryDocs(ctx, docQ)
	if err != nil {
		return nil, workspaceSummaryStats{}, err
	}

	items := make([]ticketListItem, 0, len(res.Docs))
	stats := workspaceSummaryStats{}

	for _, h := range res.Docs {
		if h.Doc == nil {
			continue
		}
		doc := h.Doc
		ticketID := strings.TrimSpace(doc.Ticket)
		if ticketID == "" {
			continue
		}
		indexRel := relPath(ws, h.Path)
		ticketDir := filepath.ToSlash(filepath.Dir(indexRel))
		createdAt := inferCreatedAt(ticketDir)
		updatedAt := docUpdatedAt(ws, h.Path, doc)

		item := ticketListItem{
			Ticket:    ticketID,
			Title:     strings.TrimSpace(doc.Title),
			Status:    strings.TrimSpace(doc.Status),
			Intent:    strings.TrimSpace(doc.Intent),
			Owners:    append([]string{}, doc.Owners...),
			Topics:    append([]string{}, doc.Topics...),
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
			TicketDir: ticketDir,
			IndexPath: indexRel,
			Snippet:   "",
			Stats:     nil,
		}
		if strings.TrimSpace(q.Q) != "" {
			item.Snippet = searchsvc.ExtractSnippet(h.Body, q.Q, 100)
		}

		if q.IncludeStats {
			docsTotal, relatedFilesTotal, _ := ticketDocStats(ctx, ws, ticketID)
			tasksTotal, tasksDone, _ := ticketTaskCounts(ws, ticketDir)
			item.Stats = &ticketListItemStats{
				DocsTotal:         docsTotal,
				TasksTotal:        tasksTotal,
				TasksDone:         tasksDone,
				RelatedFilesTotal: relatedFilesTotal,
			}
		}

		stats.TicketsTotal++
		switch strings.ToLower(item.Status) {
		case "active":
			stats.TicketsActive++
		case "review":
			stats.TicketsReview++
		case "complete":
			stats.TicketsComplete++
		case "draft":
			stats.TicketsDraft++
		}

		items = append(items, item)
	}

	switch q.OrderBy {
	case "ticket":
		sort.SliceStable(items, func(i, j int) bool {
			if items[i].Ticket == items[j].Ticket {
				return items[i].IndexPath < items[j].IndexPath
			}
			if q.Reverse {
				return items[i].Ticket > items[j].Ticket
			}
			return items[i].Ticket < items[j].Ticket
		})
	case "title":
		sort.SliceStable(items, func(i, j int) bool {
			ai := strings.ToLower(items[i].Title)
			aj := strings.ToLower(items[j].Title)
			if ai == aj {
				if q.Reverse {
					return items[i].Ticket > items[j].Ticket
				}
				return items[i].Ticket < items[j].Ticket
			}
			if q.Reverse {
				return ai > aj
			}
			return ai < aj
		})
	case "last_updated":
		sort.SliceStable(items, func(i, j int) bool {
			if items[i].UpdatedAt == items[j].UpdatedAt {
				if q.Reverse {
					return items[i].Ticket > items[j].Ticket
				}
				return items[i].Ticket < items[j].Ticket
			}
			if q.Reverse {
				return items[i].UpdatedAt > items[j].UpdatedAt
			}
			return items[i].UpdatedAt < items[j].UpdatedAt
		})
	}

	return items, stats, nil
}

func relPath(ws *workspace.Workspace, abs string) string {
	rel, err := filepath.Rel(ws.Context().Root, abs)
	if err != nil {
		return filepath.ToSlash(filepath.Clean(abs))
	}
	return filepath.ToSlash(filepath.Clean(rel))
}

func docUpdatedAt(ws *workspace.Workspace, absPath string, doc *models.Document) string {
	if doc != nil && !doc.LastUpdated.IsZero() {
		return doc.LastUpdated.UTC().Format(time.RFC3339Nano)
	}
	if fi, err := os.Stat(absPath); err == nil {
		return fi.ModTime().UTC().Format(time.RFC3339Nano)
	}
	return ""
}

func inferCreatedAt(ticketDirRel string) string {
	parts := strings.Split(strings.Trim(filepath.ToSlash(ticketDirRel), "/"), "/")
	if len(parts) < 4 {
		return ""
	}
	yyyy := parts[0]
	mm := parts[1]
	dd := parts[2]
	if len(yyyy) != 4 || len(mm) != 2 || len(dd) != 2 {
		return ""
	}
	return yyyy + "-" + mm + "-" + dd
}

func vocabSlugs(items []models.VocabItem) []string {
	out := make([]string, 0, len(items))
	for _, it := range items {
		s := strings.TrimSpace(it.Slug)
		if s == "" {
			continue
		}
		out = append(out, s)
	}
	return out
}

func distinctDocColumn(ctx context.Context, db *sql.DB, col string, includeArchived bool) []string {
	col = strings.TrimSpace(col)
	if col == "" || db == nil {
		return nil
	}
	where := "parse_ok = 1 AND COALESCE(" + col + ",'') <> ''"
	if !includeArchived {
		where += " AND is_archived_path = 0"
	}
	// #nosec G202 -- column name is hard-coded by caller (not user input).
	rows, err := db.QueryContext(ctx, `SELECT DISTINCT `+col+` FROM docs WHERE `+where+` ORDER BY `+col+` ASC;`)
	if err != nil {
		return nil
	}
	defer func() { _ = rows.Close() }()
	var out []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			continue
		}
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		out = append(out, v)
	}
	return out
}

func distinctTopics(ctx context.Context, db *sql.DB, includeArchived bool) []string {
	if db == nil {
		return nil
	}
	where := "d.parse_ok = 1"
	if !includeArchived {
		where += " AND d.is_archived_path = 0"
	}
	// #nosec G202 -- statement is static.
	rows, err := db.QueryContext(ctx, `
SELECT DISTINCT COALESCE(t.topic_original,'') AS topic
FROM doc_topics t
JOIN docs d ON d.doc_id = t.doc_id
WHERE `+where+`
ORDER BY t.topic_lower ASC;
`)
	if err != nil {
		return nil
	}
	defer func() { _ = rows.Close() }()
	var out []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			continue
		}
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		out = append(out, v)
	}
	return out
}

func distinctOwners(ctx context.Context, db *sql.DB, includeArchived bool) []string {
	if db == nil {
		return nil
	}
	where := "d.parse_ok = 1"
	if !includeArchived {
		where += " AND d.is_archived_path = 0"
	}
	// #nosec G202 -- statement is static.
	rows, err := db.QueryContext(ctx, `
SELECT DISTINCT COALESCE(o.owner_original,'') AS owner
FROM doc_owners o
JOIN docs d ON d.doc_id = o.doc_id
WHERE `+where+`
ORDER BY o.owner_lower ASC;
`)
	if err != nil {
		return nil
	}
	defer func() { _ = rows.Close() }()
	var out []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			continue
		}
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		out = append(out, v)
	}
	return out
}
