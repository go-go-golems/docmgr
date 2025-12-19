package workspace

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync/atomic"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

var workspaceSQLiteCounter uint64

// openInMemorySQLite opens an in-memory SQLite database connection.
//
// Note: This uses the sqlite3 driver. The DSN uses a shared cache so multiple
// connections (if we ever use them) can still see the same in-memory DB.
func openInMemorySQLite(ctx context.Context) (*sql.DB, error) {
	// IMPORTANT: Use a unique named in-memory DB per Workspace instance.
	//
	// If we reuse a fixed name with cache=shared, independent Workspace instances
	// (including different unit tests in the same package) can unintentionally share
	// state inside the same process, leading to confusing, flaky results.
	name := atomic.AddUint64(&workspaceSQLiteCounter, 1)
	dsn := fmt.Sprintf("file:docmgr_workspace_%d?mode=memory&cache=shared", name)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, errors.Wrap(err, "open sqlite in-memory")
	}
	// Allow multiple connections because query code may issue nested statements while
	// iterating result sets (e.g., best-effort hydration of topics/related_files).
	// With SetMaxOpenConns(1), such patterns can deadlock/hang.
	//
	// NOTE: DSN uses a named shared in-memory DB (cache=shared) so multiple connections
	// see the same schema/data.
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(4)

	if err := applySQLitePragmas(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func applySQLitePragmas(ctx context.Context, db *sql.DB) error {
	// Pragmas are best-effort; if one fails we want to know because it impacts correctness/perf.
	stmts := []string{
		"PRAGMA foreign_keys = ON;",
		"PRAGMA journal_mode = OFF;",  // in-memory; journaling not needed
		"PRAGMA synchronous = OFF;",   // in-memory; ok to disable durability
		"PRAGMA temp_store = MEMORY;", // keep temp data in memory
	}
	for _, stmt := range stmts {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return errors.Wrapf(err, "apply pragma: %s", strings.TrimSpace(stmt))
		}
	}
	return nil
}

// createWorkspaceSchema creates the minimal schema for the Workspace in-memory index.
//
// Spec: §9.1–§9.2.
func createWorkspaceSchema(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "begin schema tx")
	}
	defer func() { _ = tx.Rollback() }()

	ddl := []string{
		// docs: one row per markdown document
		`
CREATE TABLE IF NOT EXISTS docs (
    doc_id INTEGER PRIMARY KEY,
    path TEXT NOT NULL UNIQUE,              -- absolute path to .md file
    ticket_id TEXT,                          -- from frontmatter Ticket field
    doc_type TEXT,                           -- from frontmatter DocType
    status TEXT,                             -- from frontmatter Status
    intent TEXT,                             -- from frontmatter Intent
    title TEXT,                              -- from frontmatter Title
    last_updated TEXT,                       -- ISO8601 timestamp from frontmatter
    what_for TEXT,                           -- from frontmatter WhatFor (for skills)
    when_to_use TEXT,                        -- from frontmatter WhenToUse (for skills)

    -- Parse state
    parse_ok INTEGER NOT NULL DEFAULT 1,     -- 1 if frontmatter parsed successfully, 0 otherwise
    parse_err TEXT,                          -- error message if parse_ok=0

    -- Path category tags (for filtering)
    is_index INTEGER NOT NULL DEFAULT 0,          -- 1 if path ends with /index.md
    is_archived_path INTEGER NOT NULL DEFAULT 0,  -- 1 if path contains /archive/
    is_scripts_path INTEGER NOT NULL DEFAULT 0,   -- 1 if path contains /scripts/
    is_sources_path INTEGER NOT NULL DEFAULT 0,   -- 1 if path contains /sources/
    is_control_doc INTEGER NOT NULL DEFAULT 0,    -- 1 if basename is README.md, tasks.md, or changelog.md

    -- Optional: full document body (can be skipped initially for memory)
    body TEXT
);
`,
		`CREATE INDEX IF NOT EXISTS idx_docs_ticket_id ON docs(ticket_id);`,
		`CREATE INDEX IF NOT EXISTS idx_docs_parse_ok ON docs(parse_ok);`,
		`CREATE INDEX IF NOT EXISTS idx_docs_path_tags ON docs(is_archived_path, is_scripts_path, is_control_doc);`,

		// doc_topics: many-to-many doc ↔ topic
		`
CREATE TABLE IF NOT EXISTS doc_topics (
    doc_id INTEGER NOT NULL,
    topic_lower TEXT NOT NULL,              -- lowercase topic for case-insensitive matching
    topic_original TEXT,                    -- original case (for display)
    PRIMARY KEY (doc_id, topic_lower),
    FOREIGN KEY (doc_id) REFERENCES docs(doc_id) ON DELETE CASCADE
);
`,
		`CREATE INDEX IF NOT EXISTS idx_doc_topics_topic ON doc_topics(topic_lower);`,

		// related_files: one row per RelatedFiles entry
		`
CREATE TABLE IF NOT EXISTS related_files (
    rf_id INTEGER PRIMARY KEY,
    doc_id INTEGER NOT NULL,
    note TEXT,                              -- optional note from RelatedFiles entry

    -- Normalized path keys (multiple representations for reliable matching)
    norm_canonical TEXT,                    -- canonical best-effort key (prefers repo_rel, then docs_rel, then doc_rel, then abs)
    norm_repo_rel TEXT,                     -- repo-relative path (preferred canonical key)
    norm_docs_rel TEXT,                     -- docs-root relative path (fallback)
    norm_doc_rel TEXT,                      -- doc-relative path (fallback; may include ../)
    norm_abs TEXT,                          -- absolute path (fallback)
    norm_clean TEXT,                        -- cleaned relative path derived from raw_path (fallback)
    anchor TEXT,                            -- which anchor was used (repo/doc/config/etc)

    -- Original raw path from frontmatter (for display/debugging)
    raw_path TEXT,

    FOREIGN KEY (doc_id) REFERENCES docs(doc_id) ON DELETE CASCADE
);
`,
		`CREATE INDEX IF NOT EXISTS idx_related_files_doc_id ON related_files(doc_id);`,
		`CREATE INDEX IF NOT EXISTS idx_related_files_norm_canonical ON related_files(norm_canonical);`,
		`CREATE INDEX IF NOT EXISTS idx_related_files_norm_repo_rel ON related_files(norm_repo_rel);`,
		`CREATE INDEX IF NOT EXISTS idx_related_files_norm_docs_rel ON related_files(norm_docs_rel);`,
		`CREATE INDEX IF NOT EXISTS idx_related_files_norm_doc_rel ON related_files(norm_doc_rel);`,
		`CREATE INDEX IF NOT EXISTS idx_related_files_norm_abs ON related_files(norm_abs);`,
		`CREATE INDEX IF NOT EXISTS idx_related_files_norm_clean ON related_files(norm_clean);`,
	}

	for _, stmt := range ddl {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return errors.Wrap(err, "apply schema DDL")
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "commit schema tx")
	}
	return nil
}

func sqliteQuoteStringLiteral(s string) (string, error) {
	// SQLite string literals are single-quoted. Escape by doubling single quotes.
	// Disallow NUL which sqlite treats oddly in some contexts.
	if strings.Contains(s, "\x00") {
		return "", errors.New("sqlite string literal contains NUL byte")
	}
	return "'" + strings.ReplaceAll(s, "'", "''") + "'", nil
}
