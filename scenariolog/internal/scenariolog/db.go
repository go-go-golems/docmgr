package scenariolog

import (
	"context"
	"database/sql"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

func Open(ctx context.Context, path string) (*sql.DB, error) {
	// Keep DSN simple and set pragmas explicitly (better error surfacing and easier evolution).
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, errors.Wrap(err, "open sqlite db")
	}

	// Single-writer is sufficient for the MVP and avoids surprising lock behavior.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := applyPragmas(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func applyPragmas(ctx context.Context, db *sql.DB) error {
	stmts := []string{
		"PRAGMA foreign_keys = ON;",
		"PRAGMA journal_mode = WAL;",
		"PRAGMA synchronous = NORMAL;",
		"PRAGMA busy_timeout = 5000;",
	}
	for _, stmt := range stmts {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return errors.Wrapf(err, "apply pragma: %s", strings.TrimSpace(stmt))
		}
	}
	return nil
}


