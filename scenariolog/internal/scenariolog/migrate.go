package scenariolog

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

const schemaVersionV1 = 1

func Migrate(ctx context.Context, db *sql.DB) error {
	v, err := getUserVersion(ctx, db)
	if err != nil {
		return err
	}

	switch {
	case v == 0:
		if err := migrateToV1(ctx, db); err != nil {
			return err
		}
		if err := setUserVersion(ctx, db, schemaVersionV1); err != nil {
			return err
		}
		// Best-effort: create FTS table if available (some sqlite builds omit fts5).
		_ = ensureFTS5(ctx, db)
		return nil
	case v == schemaVersionV1:
		// Best-effort: a DB may have been created on a system without FTS5 support.
		// If we later run on a build that has FTS5, we can create the table on-demand.
		_ = ensureFTS5(ctx, db)
		return nil
	default:
		return errors.Errorf("unsupported schema version: %d", v)
	}
}

func getUserVersion(ctx context.Context, db *sql.DB) (int, error) {
	var v int
	if err := db.QueryRowContext(ctx, "PRAGMA user_version;").Scan(&v); err != nil {
		return 0, errors.Wrap(err, "read PRAGMA user_version")
	}
	return v, nil
}

func setUserVersion(ctx context.Context, db *sql.DB, v int) error {
	// PRAGMA user_version doesn't accept parameters reliably in all sqlite shells/drivers.
	stmt := fmt.Sprintf("PRAGMA user_version = %d;", v)
	if _, err := db.ExecContext(ctx, stmt); err != nil {
		return errors.Wrap(err, "set PRAGMA user_version")
	}
	return nil
}

func migrateToV1(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "begin migration tx")
	}
	defer func() { _ = tx.Rollback() }()

	ddl := []string{
		// scenario_runs: one row per run
		`
CREATE TABLE IF NOT EXISTS scenario_runs (
    run_id TEXT PRIMARY KEY,
    root_dir TEXT NOT NULL,
    suite TEXT,
    started_at TEXT NOT NULL,
    completed_at TEXT,
    exit_code INTEGER,
    duration_ms INTEGER
);`,

		// steps: one row per step/script
		`
CREATE TABLE IF NOT EXISTS steps (
    step_id TEXT PRIMARY KEY,
    run_id TEXT NOT NULL,
    step_num INTEGER NOT NULL,
    step_name TEXT NOT NULL,
    script_path TEXT,
    started_at TEXT NOT NULL,
    completed_at TEXT,
    exit_code INTEGER,
    duration_ms INTEGER,
    FOREIGN KEY (run_id) REFERENCES scenario_runs(run_id) ON DELETE CASCADE
);`,
		`CREATE INDEX IF NOT EXISTS idx_steps_run ON steps(run_id);`,
		`CREATE INDEX IF NOT EXISTS idx_steps_num ON steps(run_id, step_num);`,

		// commands: optional, but included in schema so kv/artifacts can reference it safely later.
		`
CREATE TABLE IF NOT EXISTS commands (
    command_id TEXT PRIMARY KEY,
    step_id TEXT NOT NULL,
    command_num INTEGER NOT NULL,
    argv0 TEXT,
    argv_json TEXT,
    cwd TEXT,
    started_at TEXT NOT NULL,
    completed_at TEXT,
    exit_code INTEGER,
    duration_ms INTEGER,
    FOREIGN KEY (step_id) REFERENCES steps(step_id) ON DELETE CASCADE
);`,
		`CREATE INDEX IF NOT EXISTS idx_commands_step ON commands(step_id);`,
		`CREATE INDEX IF NOT EXISTS idx_commands_num ON commands(step_id, command_num);`,

		// kv: arbitrary tags at run/step/command scope
		`
CREATE TABLE IF NOT EXISTS kv (
    kv_id INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id TEXT NOT NULL,
    step_id TEXT,
    command_id TEXT,
    k TEXT NOT NULL,
    v TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    CHECK (command_id IS NULL OR step_id IS NOT NULL),
    FOREIGN KEY (run_id) REFERENCES scenario_runs(run_id) ON DELETE CASCADE,
    FOREIGN KEY (step_id) REFERENCES steps(step_id) ON DELETE CASCADE,
    FOREIGN KEY (command_id) REFERENCES commands(command_id) ON DELETE CASCADE
);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_kv_scope_key ON kv(run_id, step_id, command_id, k);`,
		`CREATE INDEX IF NOT EXISTS idx_kv_key ON kv(k);`,
		`CREATE INDEX IF NOT EXISTS idx_kv_run ON kv(run_id);`,

		// artifacts: captured files (stdout/stderr/etc) + metadata
		`
CREATE TABLE IF NOT EXISTS artifacts (
    artifact_id INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id TEXT NOT NULL,
    step_id TEXT,
    command_id TEXT,
    kind TEXT NOT NULL,
    path TEXT NOT NULL,
    is_text INTEGER NOT NULL DEFAULT 1,
    size_bytes INTEGER,
    sha256 TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    CHECK (command_id IS NULL OR step_id IS NOT NULL),
    FOREIGN KEY (run_id) REFERENCES scenario_runs(run_id) ON DELETE CASCADE,
    FOREIGN KEY (step_id) REFERENCES steps(step_id) ON DELETE CASCADE,
    FOREIGN KEY (command_id) REFERENCES commands(command_id) ON DELETE CASCADE
);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_artifacts_unique_path ON artifacts(run_id, step_id, command_id, kind, path);`,
		`CREATE INDEX IF NOT EXISTS idx_artifacts_run ON artifacts(run_id);`,
		`CREATE INDEX IF NOT EXISTS idx_artifacts_kind ON artifacts(kind);`,
		`CREATE INDEX IF NOT EXISTS idx_artifacts_is_text ON artifacts(is_text);`,

	}

	for _, stmt := range ddl {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return errors.Wrapf(err, "apply migration DDL: %s", strings.TrimSpace(stmt))
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "commit migration tx")
	}
	return nil
}

func ensureFTS5(ctx context.Context, db *sql.DB) error {
	// If FTS5 is unavailable, CREATE VIRTUAL TABLE will fail with "no such module: fts5".
	// We treat that as a supported degraded mode: the rest of the DB remains useful.
	//
	// NOTE: We intentionally run this outside the main migration transaction because
	// some drivers/builds behave differently for virtual tables. It's idempotent.
	stmt := `
CREATE VIRTUAL TABLE IF NOT EXISTS log_lines_fts USING fts5(
    run_id UNINDEXED,
    artifact_id UNINDEXED,
    line_num UNINDEXED,
    text,
    tokenize = 'unicode61'
);`
	_, err := db.ExecContext(ctx, stmt)
	if err == nil {
		return nil
	}
	if isFTS5Unavailable(err) {
		return nil
	}
	return errors.Wrap(err, "ensure fts5 table")
}

func isFTS5Unavailable(err error) bool {
	// sqlite error strings vary slightly; keep this intentionally conservative.
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "no such module: fts5")
}


