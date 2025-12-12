package workspace

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-go-golems/docmgr/internal/paths"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/core"
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

	sqlQ, err := compileDocQuery(ctx, w, q)
	if err != nil {
		return DocQueryResult{}, err
	}

	rows, err := w.db.QueryContext(ctx, sqlQ.SQL, sqlQ.Args...)
	if err != nil {
		return DocQueryResult{}, errors.Wrap(err, "query docs")
	}
	defer func() { _ = rows.Close() }()

	var handles []DocHandle

	// Hydration helpers (best-effort; if they fail we still return base rows).
	topicStmt, _ := w.db.PrepareContext(ctx, `SELECT COALESCE(topic_original,'') FROM doc_topics WHERE doc_id=? ORDER BY topic_lower;`)
	rfStmt, _ := w.db.PrepareContext(ctx, `SELECT COALESCE(raw_path,''), COALESCE(note,'') FROM related_files WHERE doc_id=? ORDER BY rf_id;`)
	if topicStmt != nil {
		defer func() { _ = topicStmt.Close() }()
	}
	if rfStmt != nil {
		defer func() { _ = rfStmt.Close() }()
	}

	for rows.Next() {
		if err := ctx.Err(); err != nil {
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
			&parseOK,
			&parseErr,
			&body,
		); err != nil {
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
			handles = append(handles, handle)
			continue
		}

		doc := &models.Document{
			Ticket:  ticketID.String,
			DocType: docType.String,
			Status:  status.String,
			Intent:  intent.String,
			Title:   title.String,
		}

		if strings.TrimSpace(lastUpdated.String) != "" {
			if t, err := time.Parse(time.RFC3339Nano, lastUpdated.String); err == nil {
				doc.LastUpdated = t
			}
		}

		// Best-effort hydrate topics.
		if topicStmt != nil {
			if topics, err := fetchTopics(ctx, topicStmt, docID); err == nil {
				doc.Topics = topics
			}
		}

		// Best-effort hydrate related files (raw path + note for display/UX).
		if rfStmt != nil {
			if rfs, err := fetchRelatedFiles(ctx, rfStmt, docID); err == nil {
				doc.RelatedFiles = rfs
			}
		}

		handle.Doc = doc
		handles = append(handles, handle)
	}
	if err := rows.Err(); err != nil {
		return DocQueryResult{}, errors.Wrap(err, "iterate docs rows")
	}

	return DocQueryResult{
		Docs:        handles,
		Diagnostics: nil, // Task 8: fill this when IncludeDiagnostics=true
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

func fetchTopics(ctx context.Context, stmt *sql.Stmt, docID int64) ([]string, error) {
	rows, err := stmt.QueryContext(ctx, docID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		t = strings.TrimSpace(t)
		if t != "" {
			out = append(out, t)
		}
	}
	return out, rows.Err()
}

func fetchRelatedFiles(ctx context.Context, stmt *sql.Stmt, docID int64) (models.RelatedFiles, error) {
	rows, err := stmt.QueryContext(ctx, docID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out models.RelatedFiles
	for rows.Next() {
		var raw, note string
		if err := rows.Scan(&raw, &note); err != nil {
			return nil, err
		}
		raw = strings.TrimSpace(raw)
		note = strings.TrimSpace(note)
		if raw == "" {
			continue
		}
		out = append(out, models.RelatedFile{Path: raw, Note: note})
	}
	return out, rows.Err()
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


