package scenariolog

import (
	"context"
	"database/sql"
	"testing"
)

func TestMigrateV1CreatesExpectedTables(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = db.Close() }()

	if err := Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	want := []string{
		"scenario_runs",
		"steps",
		"commands",
		"kv",
		"artifacts",
	}
	for _, table := range want {
		exists, err := tableExists(ctx, db, table)
		if err != nil {
			t.Fatalf("tableExists(%q): %v", table, err)
		}
		if !exists {
			t.Fatalf("expected table %q to exist", table)
		}
	}

	// FTS is best-effort and may be unavailable depending on sqlite compile options.
	// Migrate() must still succeed in degraded mode.
	_, _ = tableExists(ctx, db, "log_lines_fts")
}

func tableExists(ctx context.Context, db *sql.DB, name string) (bool, error) {
	var c int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type IN ('table','view') AND name = ?;", name).Scan(&c)
	if err != nil {
		return false, err
	}
	return c > 0, nil
}


