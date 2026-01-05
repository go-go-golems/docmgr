package workspace

import (
	"context"
	"database/sql"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-go-golems/docmgr/internal/documents"
	"github.com/go-go-golems/docmgr/internal/paths"
	"github.com/go-go-golems/docmgr/pkg/models"
	"github.com/pkg/errors"
)

// BuildIndexOptions controls which data is stored in the in-memory index.
type BuildIndexOptions struct {
	// IncludeBody stores the full markdown body into docs.body.
	// Default false to keep memory usage low.
	IncludeBody bool
}

// InitIndex initializes (or rebuilds) the in-memory SQLite index for this workspace.
//
// Current policy: rebuild from scratch per CLI invocation (Decision Q16).
func (w *Workspace) InitIndex(ctx context.Context, opts BuildIndexOptions) error {
	if ctx == nil {
		return errors.New("nil context")
	}
	if w.ctx.Root == "" {
		return errors.New("workspace has empty Root")
	}

	if w.db != nil {
		_ = w.db.Close()
		w.db = nil
	}

	db, err := openInMemorySQLite(ctx)
	if err != nil {
		return err
	}
	if err := createWorkspaceSchema(ctx, db); err != nil {
		_ = db.Close()
		return err
	}

	ftsOK, err := ensureDocsFTS5(ctx, db)
	if err != nil {
		_ = db.Close()
		return err
	}
	// Defensive: verify table existence in sqlite_master in case CREATE VIRTUAL TABLE
	// succeeded partially or was silently ignored in some environments.
	if ftsOK {
		if exists, err := sqliteTableExists(ctx, db, "docs_fts"); err == nil {
			ftsOK = exists
		} else {
			_ = db.Close()
			return err
		}
	}

	if err := ingestWorkspaceDocs(ctx, db, w.ctx, opts, ftsOK); err != nil {
		_ = db.Close()
		return err
	}

	w.db = db
	w.ftsAvailable = ftsOK
	return nil
}

func ingestWorkspaceDocs(ctx context.Context, db *sql.DB, wctx WorkspaceContext, opts BuildIndexOptions, ftsOK bool) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "begin ingest tx")
	}
	defer func() { _ = tx.Rollback() }()

	insertDocStmt, err := tx.PrepareContext(ctx, `
INSERT INTO docs (
  path, ticket_id, doc_type, status, intent, title, last_updated,
  what_for, when_to_use,
  parse_ok, parse_err,
  is_index, is_archived_path, is_scripts_path, is_sources_path, is_control_doc,
  body
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`)
	if err != nil {
		return errors.Wrap(err, "prepare insert docs")
	}
	defer func() { _ = insertDocStmt.Close() }()

	insertTopicStmt, err := tx.PrepareContext(ctx, `
INSERT OR IGNORE INTO doc_topics (doc_id, topic_lower, topic_original)
VALUES (?, ?, ?)
`)
	if err != nil {
		return errors.Wrap(err, "prepare insert doc_topics")
	}
	defer func() { _ = insertTopicStmt.Close() }()

	insertOwnerStmt, err := tx.PrepareContext(ctx, `
INSERT OR IGNORE INTO doc_owners (doc_id, owner_lower, owner_original)
VALUES (?, ?, ?)
`)
	if err != nil {
		return errors.Wrap(err, "prepare insert doc_owners")
	}
	defer func() { _ = insertOwnerStmt.Close() }()

	insertRFStmt, err := tx.PrepareContext(ctx, `
INSERT INTO related_files (
  doc_id, note,
  norm_canonical, norm_repo_rel, norm_docs_rel, norm_doc_rel, norm_abs, norm_clean,
  anchor, raw_path
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`)
	if err != nil {
		return errors.Wrap(err, "prepare insert related_files")
	}
	defer func() { _ = insertRFStmt.Close() }()

	var insertFTSStmt *sql.Stmt
	if ftsOK {
		insertFTSStmt, err = tx.PrepareContext(ctx, `
INSERT INTO docs_fts (rowid, title, body, topics, doc_type, ticket_id)
VALUES (?, ?, ?, ?, ?, ?)
`)
		if err != nil {
			return errors.Wrap(err, "prepare insert docs_fts")
		}
		defer func() { _ = insertFTSStmt.Close() }()
	}

	walkErr := documents.WalkDocuments(wctx.Root, func(path string, doc *models.Document, body string, readErr error) error {
		if err := ctx.Err(); err != nil {
			return err
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			absPath = path
		}
		absPath = filepath.Clean(absPath)

		tags := ComputePathTags(absPath)

		parseOK := 1
		parseErr := ""
		var ticketID, docType, status, intent, title sql.NullString
		var lastUpdated sql.NullString
		var whatFor, whenToUse sql.NullString
		var bodyVal sql.NullString

		if readErr != nil || doc == nil {
			parseOK = 0
			// Fallback: infer ticket ID from directory structure so broken docs can still be
			// discovered by ticket-scoped queries (useful for diagnostics/repair flows).
			//
			// Example: <docsRoot>/YYYY/MM/DD/<TICKET--slug>/... -> ticket_id = <TICKET>
			ticketID = nullString(inferTicketIDFromPath(wctx.Root, absPath))
			if readErr != nil {
				parseErr = readErr.Error()
			} else {
				parseErr = "unknown read error"
			}
		} else {
			ticketID = nullString(doc.Ticket)
			docType = nullString(doc.DocType)
			status = nullString(doc.Status)
			intent = nullString(doc.Intent)
			title = nullString(doc.Title)
			whatFor = nullString(doc.WhatFor)
			whenToUse = nullString(doc.WhenToUse)
			if !doc.LastUpdated.IsZero() {
				lastUpdated = sql.NullString{String: doc.LastUpdated.UTC().Format(time.RFC3339Nano), Valid: true}
			}
			if opts.IncludeBody {
				bodyVal = sql.NullString{String: body, Valid: true}
			}
		}

		res, err := insertDocStmt.ExecContext(
			ctx,
			filepath.ToSlash(absPath),
			ticketID, docType, status, intent, title, lastUpdated,
			whatFor, whenToUse,
			parseOK, nullString(parseErr),
			boolToInt(tags.IsIndex),
			boolToInt(tags.IsArchivedPath),
			boolToInt(tags.IsScriptsPath),
			boolToInt(tags.IsSourcesPath),
			boolToInt(tags.IsControlDoc),
			bodyVal,
		)
		if err != nil {
			return errors.Wrap(err, "insert docs row")
		}
		docID, err := res.LastInsertId()
		if err != nil {
			return errors.Wrap(err, "docs last insert id")
		}

		if parseOK == 0 || doc == nil {
			return nil
		}

		if insertFTSStmt != nil {
			topicsText := strings.TrimSpace(strings.Join(doc.Topics, " "))
			_, err := insertFTSStmt.ExecContext(
				ctx,
				docID,
				nullString(doc.Title),
				nullString(body),
				nullString(topicsText),
				nullString(doc.DocType),
				nullString(doc.Ticket),
			)
			if err != nil {
				return errors.Wrap(err, "insert docs_fts row")
			}
		}

		for _, topic := range doc.Topics {
			topic = strings.TrimSpace(topic)
			if topic == "" {
				continue
			}
			_, err := insertTopicStmt.ExecContext(ctx, docID, strings.ToLower(topic), topic)
			if err != nil {
				return errors.Wrap(err, "insert doc_topics row")
			}
		}

		for _, owner := range doc.Owners {
			owner = strings.TrimSpace(owner)
			if owner == "" {
				continue
			}
			_, err := insertOwnerStmt.ExecContext(ctx, docID, strings.ToLower(owner), owner)
			if err != nil {
				return errors.Wrap(err, "insert doc_owners row")
			}
		}

		// Use a resolver anchored at this document path so doc-relative entries normalize correctly.
		resolver := paths.NewResolver(paths.ResolverOptions{
			DocsRoot:  wctx.Root,
			DocPath:   absPath,
			ConfigDir: wctx.ConfigDir,
			RepoRoot:  wctx.RepoRoot,
		})
		for _, rf := range doc.RelatedFiles {
			raw := strings.TrimSpace(rf.Path)
			if raw == "" {
				continue
			}
			n := normalizeRelatedFile(resolver, raw)
			_, err := insertRFStmt.ExecContext(
				ctx,
				docID,
				nullString(rf.Note),
				nullString(n.Canonical),
				nullString(n.RepoRelative),
				nullString(n.DocsRelative),
				nullString(n.DocRelative),
				nullString(n.Abs),
				nullString(n.Clean),
				nullString(n.Anchor),
				nullString(raw),
			)
			if err != nil {
				return errors.Wrap(err, "insert related_files row")
			}
		}

		return nil
	}, documents.WithSkipDir(func(_ string, d fs.DirEntry) bool {
		return DefaultIngestSkipDir("", d)
	}))

	if walkErr != nil {
		return errors.Wrap(walkErr, "walk documents for ingest")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "commit ingest tx")
	}
	return nil
}

// inferTicketIDFromPath best-effort extracts a ticket ID from a document path under the docs root.
//
// Expected docs layout:
//
//	<docsRoot>/<YYYY>/<MM>/<DD>/<TICKET--slug>/...
//
// It returns "" if it cannot infer a ticket ID.
func inferTicketIDFromPath(docsRoot string, absDocPath string) string {
	docsRoot = filepath.Clean(strings.TrimSpace(docsRoot))
	absDocPath = filepath.Clean(strings.TrimSpace(absDocPath))
	if docsRoot == "" || absDocPath == "" {
		return ""
	}
	rel, err := filepath.Rel(docsRoot, absDocPath)
	if err != nil {
		return ""
	}
	rel = filepath.Clean(rel)
	if rel == "." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." {
		return ""
	}
	parts := strings.Split(rel, string(filepath.Separator))
	// Need at least YYYY/MM/DD/<ticketDir>/<file>
	if len(parts) < 4 {
		return ""
	}
	ticketDir := strings.TrimSpace(parts[3])
	if ticketDir == "" {
		return ""
	}
	// ticket dir is typically "<TICKET>--<slug>"
	if i := strings.Index(ticketDir, "--"); i > 0 {
		return strings.TrimSpace(ticketDir[:i])
	}
	return ""
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func nullString(s string) sql.NullString {
	s = strings.TrimSpace(s)
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func normalizeCleanPath(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	// Preserve relative paths (including leading "..") but normalize separators and remove redundant segments.
	cleaned := filepath.ToSlash(filepath.Clean(raw))
	return cleaned
}
