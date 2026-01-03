package workspace

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-go-golems/docmgr/internal/paths"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/core"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/docmgrctx"
	"github.com/go-go-golems/docmgr/pkg/models"
	"github.com/pkg/errors"
)

// QueryDocs is the primary API for document lookup.
//
// Spec: §5.2, §10.1–§10.4.
func (w *Workspace) QueryDocs(ctx context.Context, q DocQuery) (DocQueryResult, error) {
	if ctx == nil {
		return DocQueryResult{}, errors.New("nil context")
	}
	if w.db == nil {
		return DocQueryResult{}, errors.New("workspace index not initialized (db is nil); call InitIndex first")
	}

	if q.Options.OrderBy == "" {
		q.Options.OrderBy = OrderByPath
	}

	if err := validateDocQuery(q); err != nil {
		return DocQueryResult{}, err
	}

	var diags []core.Taxonomy
	if q.Options.IncludeDiagnostics {
		// Reverse-lookup normalization diagnostics (best-effort).
		// We emit a warning when we can't derive strong keys (canonical/repo-rel/abs) and must rely
		// on weaker fallbacks (clean/raw). Matching still proceeds (fallback strategy), but we want
		// to explain why results may be surprising.
		for _, raw := range q.Filters.RelatedFile {
			raw = strings.TrimSpace(raw)
			if raw == "" {
				continue
			}
			n := w.resolver.Normalize(raw)
			canon := strings.TrimSpace(n.Canonical)
			repoRel := strings.TrimSpace(n.RepoRelative)
			abs := strings.TrimSpace(n.Abs)
			clean := strings.TrimSpace(normalizeCleanPath(raw))
			if canon == "" && repoRel == "" && abs == "" && clean != "" {
				if t := docmgrctx.NewWorkspaceQueryNormalizationFallbackTaxonomy("file", raw, "no canonical/repo/abs key; falling back to cleaned/raw matching"); t != nil {
					diags = append(diags, *t)
				}
			}
		}
		for _, raw := range q.Filters.RelatedDir {
			raw = strings.TrimSpace(raw)
			if raw == "" {
				continue
			}
			n := w.resolver.Normalize(raw)
			canon := strings.TrimSpace(n.Canonical)
			repoRel := strings.TrimSpace(n.RepoRelative)
			abs := strings.TrimSpace(n.Abs)
			clean := strings.TrimSpace(normalizeCleanPath(raw))
			if canon == "" && repoRel == "" && abs == "" && clean != "" {
				if t := docmgrctx.NewWorkspaceQueryNormalizationFallbackTaxonomy("dir", raw, "no canonical/repo/abs key; falling back to cleaned/raw prefix matching"); t != nil {
					diags = append(diags, *t)
				}
			}
		}
	}

	sqlQ, err := compileDocQuery(ctx, w, q)
	if err != nil {
		return DocQueryResult{}, err
	}

	rows, err := w.db.QueryContext(ctx, sqlQ.SQL, sqlQ.Args...)
	if err != nil {
		return DocQueryResult{}, errors.Wrap(err, "query docs")
	}
	// NOTE: We intentionally avoid nested queries while iterating `rows` to prevent
	// N+1 behavior and connection-pool deadlocks (especially if MaxOpenConns is low).
	// We scan all rows first, then batch-hydrate topics and related_files in 1 query each.

	type pendingRow struct {
		docID   int64
		parseOK bool
		handle  DocHandle
	}

	var pending []pendingRow
	var okDocIDs []int64

	for rows.Next() {
		if err := ctx.Err(); err != nil {
			_ = rows.Close()
			return DocQueryResult{}, err
		}

		var (
			docID       int64
			path        string
			ticketID    sql.NullString
			docType     sql.NullString
			status      sql.NullString
			intent      sql.NullString
			title       sql.NullString
			lastUpdated sql.NullString
			whatFor     sql.NullString
			whenToUse   sql.NullString
			parseOK     int
			parseErr    sql.NullString
			body        sql.NullString
		)

		if err := rows.Scan(
			&docID,
			&path,
			&ticketID,
			&docType,
			&status,
			&intent,
			&title,
			&lastUpdated,
			&whatFor,
			&whenToUse,
			&parseOK,
			&parseErr,
			&body,
		); err != nil {
			_ = rows.Close()
			return DocQueryResult{}, errors.Wrap(err, "scan docs row")
		}

		handle := DocHandle{
			Path: filepath.ToSlash(filepath.Clean(path)),
		}

		if q.Options.IncludeBody && body.Valid {
			handle.Body = body.String
		}

		if parseOK == 0 {
			if parseErr.Valid && strings.TrimSpace(parseErr.String) != "" {
				handle.ReadErr = errors.New(parseErr.String)
			} else {
				handle.ReadErr = errors.New("document parse failed")
			}
			// Best-effort: keep ticket_id available even for parse-error docs so callers
			// (notably `doctor`) can group findings by ticket without needing a separate lookup.
			if strings.TrimSpace(ticketID.String) != "" {
				handle.Doc = &models.Document{Ticket: ticketID.String}
			}
			pending = append(pending, pendingRow{docID: docID, parseOK: false, handle: handle})
			continue
		}

		doc := &models.Document{
			Ticket:    ticketID.String,
			DocType:   docType.String,
			Status:    status.String,
			Intent:    intent.String,
			Title:     title.String,
			WhatFor:   whatFor.String,
			WhenToUse: whenToUse.String,
		}

		if strings.TrimSpace(lastUpdated.String) != "" {
			if t, err := time.Parse(time.RFC3339Nano, lastUpdated.String); err == nil {
				doc.LastUpdated = t
			}
		}

		handle.Doc = doc
		pending = append(pending, pendingRow{docID: docID, parseOK: true, handle: handle})
		okDocIDs = append(okDocIDs, docID)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return DocQueryResult{}, errors.Wrap(err, "iterate docs rows")
	}
	_ = rows.Close()

	// Batch hydrate topics + related_files for parse_ok docs.
	topicsByDocID := map[int64][]string{}
	rfsByDocID := map[int64]models.RelatedFiles{}

	if len(okDocIDs) > 0 {
		if topics, err := fetchTopicsByDocIDs(ctx, w.db, okDocIDs); err == nil {
			topicsByDocID = topics
		}
		if rfs, err := fetchRelatedFilesByDocIDs(ctx, w.db, okDocIDs); err == nil {
			rfsByDocID = rfs
		}
	}

	handles := make([]DocHandle, 0, len(pending))
	for _, p := range pending {
		if p.parseOK && p.handle.Doc != nil {
			if topics, ok := topicsByDocID[p.docID]; ok {
				p.handle.Doc.Topics = topics
			}
			if rfs, ok := rfsByDocID[p.docID]; ok {
				p.handle.Doc.RelatedFiles = rfs
			}
		}
		handles = append(handles, p.handle)
	}

	// Diagnostics: include parse-error docs as taxonomy entries (but do not include them in Docs)
	// when IncludeErrors=false and IncludeDiagnostics=true.
	if q.Options.IncludeDiagnostics && !q.Options.IncludeErrors {
		zero := 0
		diagQ, err := compileDocQueryWithParseFilter(ctx, w, q, &zero)
		if err == nil {
			diagRows, qerr := w.db.QueryContext(ctx, diagQ.SQL, diagQ.Args...)
			if qerr == nil {
				defer func() { _ = diagRows.Close() }()
				for diagRows.Next() {
					var (
						_docID     int64
						_path      string
						_ticketID  sql.NullString
						_docType   sql.NullString
						_status    sql.NullString
						_intent    sql.NullString
						_title     sql.NullString
						_lastUpd   sql.NullString
						_whatFor   sql.NullString
						_whenToUse sql.NullString
						_parseOK   int
						_parseErr  sql.NullString
						_body      sql.NullString
					)
					if err := diagRows.Scan(
						&_docID,
						&_path,
						&_ticketID,
						&_docType,
						&_status,
						&_intent,
						&_title,
						&_lastUpd,
						&_whatFor,
						&_whenToUse,
						&_parseOK,
						&_parseErr,
						&_body,
					); err != nil {
						continue
					}
					reason := strings.TrimSpace(_parseErr.String)
					if reason == "" {
						reason = "invalid frontmatter / parse failed"
					}
					if t := docmgrctx.NewWorkspaceQuerySkippedParseTaxonomy(filepath.ToSlash(filepath.Clean(_path)), reason, errors.New(reason)); t != nil {
						diags = append(diags, *t)
					}
				}
			}
		}
	}

	return DocQueryResult{
		Docs:        handles,
		Diagnostics: diags,
	}, nil
}

// ---- Public request/response types (Spec §5.2) ----

type DocQuery struct {
	Scope   Scope
	Filters DocFilters
	Options DocQueryOptions
}

type ScopeKind int

const (
	ScopeRepo ScopeKind = iota
	ScopeTicket
	ScopeDoc
)

type Scope struct {
	Kind     ScopeKind
	TicketID string // if ScopeTicket
	DocPath  string // if ScopeDoc (user-provided)
}

type DocFilters struct {
	Ticket  string
	DocType string
	Status  string

	RelatedFile []string
	RelatedDir  []string

	TopicsAny []string
}

type OrderBy string

const (
	OrderByPath        OrderBy = "path"
	OrderByLastUpdated OrderBy = "last_updated"
)

type DocQueryOptions struct {
	IncludeBody   bool
	IncludeErrors bool
	OrderBy       OrderBy
	Reverse       bool

	IncludeArchivedPath bool
	IncludeControlDocs  bool
	IncludeScriptsPath  bool

	IncludeDiagnostics bool
}

type DocHandle struct {
	Path    string
	Doc     *models.Document
	Body    string
	ReadErr error
}

type DocQueryResult struct {
	Docs        []DocHandle
	Diagnostics []core.Taxonomy
}

// ---- Internal helpers ----

func validateDocQuery(q DocQuery) error {
	switch q.Scope.Kind {
	case ScopeRepo:
	case ScopeTicket:
		if strings.TrimSpace(q.Scope.TicketID) == "" {
			return errors.New("scope ticket requires TicketID")
		}
		if strings.TrimSpace(q.Filters.Ticket) != "" && strings.TrimSpace(q.Filters.Ticket) != strings.TrimSpace(q.Scope.TicketID) {
			return errors.Errorf("contradictory query: ScopeTicket=%q but Filters.Ticket=%q", q.Scope.TicketID, q.Filters.Ticket)
		}
	case ScopeDoc:
		if strings.TrimSpace(q.Scope.DocPath) == "" {
			return errors.New("scope doc requires DocPath")
		}
	default:
		return errors.Errorf("unknown scope kind: %d", q.Scope.Kind)
	}

	return nil
}

func fetchTopicsByDocIDs(ctx context.Context, db *sql.DB, docIDs []int64) (map[int64][]string, error) {
	docIDs = uniqueInt64(docIDs...)
	if len(docIDs) == 0 || db == nil {
		return map[int64][]string{}, nil
	}
	placeholders := makePlaceholders(len(docIDs))
	// #nosec G202 -- placeholders are generated ("?,?") and values are bound via args, not string-interpolated.
	sqlQ := `SELECT doc_id, COALESCE(topic_original,'') FROM doc_topics WHERE doc_id IN (` + placeholders + `) ORDER BY doc_id, topic_lower;`
	args := make([]any, 0, len(docIDs))
	for _, id := range docIDs {
		args = append(args, id)
	}
	rows, err := db.QueryContext(ctx, sqlQ, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := map[int64][]string{}
	for rows.Next() {
		var docID int64
		var t string
		if err := rows.Scan(&docID, &t); err != nil {
			return nil, err
		}
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		out[docID] = append(out[docID], t)
	}
	return out, rows.Err()
}

func fetchRelatedFilesByDocIDs(ctx context.Context, db *sql.DB, docIDs []int64) (map[int64]models.RelatedFiles, error) {
	docIDs = uniqueInt64(docIDs...)
	if len(docIDs) == 0 || db == nil {
		return map[int64]models.RelatedFiles{}, nil
	}
	placeholders := makePlaceholders(len(docIDs))
	// #nosec G202 -- placeholders are generated ("?,?") and values are bound via args, not string-interpolated.
	sqlQ := `SELECT doc_id, COALESCE(raw_path,''), COALESCE(note,'') FROM related_files WHERE doc_id IN (` + placeholders + `) ORDER BY doc_id, rf_id;`
	args := make([]any, 0, len(docIDs))
	for _, id := range docIDs {
		args = append(args, id)
	}
	rows, err := db.QueryContext(ctx, sqlQ, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := map[int64]models.RelatedFiles{}
	for rows.Next() {
		var docID int64
		var raw, note string
		if err := rows.Scan(&docID, &raw, &note); err != nil {
			return nil, err
		}
		raw = strings.TrimSpace(raw)
		note = strings.TrimSpace(note)
		if raw == "" {
			continue
		}
		out[docID] = append(out[docID], models.RelatedFile{Path: raw, Note: note})
	}
	return out, rows.Err()
}

func uniqueInt64(values ...int64) []int64 {
	seen := map[int64]struct{}{}
	out := make([]int64, 0, len(values))
	for _, v := range values {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

// queryPathKeys returns comparable strings for matching query inputs against persisted norm_* columns.
func queryPathKeys(resolver *paths.Resolver, raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" || resolver == nil {
		return nil
	}
	n := resolver.Normalize(raw)
	keys := []string{
		strings.TrimSpace(n.Canonical),
		strings.TrimSpace(n.RepoRelative),
		strings.TrimSpace(n.DocsRelative),
		strings.TrimSpace(n.DocRelative),
		strings.TrimSpace(n.Abs),
		strings.TrimSpace(normalizeCleanPath(raw)),
		strings.TrimSpace(n.OriginalClean),
	}
	return uniqueNonEmptyStrings(keys...)
}

func uniqueNonEmptyStrings(values ...string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		k := filepath.ToSlash(v)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, k)
	}
	return out
}
