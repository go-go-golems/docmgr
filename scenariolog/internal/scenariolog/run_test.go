package scenariolog

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestRunLifecycleStartEnd(t *testing.T) {
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

	start := time.Date(2025, 12, 13, 10, 0, 0, 0, time.UTC)
	end := start.Add(1500 * time.Millisecond)

	runID := "run-test-1"
	if err := StartRun(ctx, db, runID, "/tmp/scenario", "suite-1", start); err != nil {
		t.Fatalf("StartRun: %v", err)
	}
	if err := EndRun(ctx, db, runID, 7, end); err != nil {
		t.Fatalf("EndRun: %v", err)
	}

	// Best-effort: confirm a known kv tag exists (suite).
	ok, err := kvExists(ctx, db, runID, "suite")
	if err != nil {
		t.Fatalf("kvExists: %v", err)
	}
	if !ok {
		t.Fatalf("expected kv tag 'suite' to exist for run")
	}

	var exitCode int
	var durationMs int64
	err = db.QueryRowContext(ctx, "SELECT exit_code, duration_ms FROM scenario_runs WHERE run_id = ?;", runID).Scan(&exitCode, &durationMs)
	if err != nil {
		t.Fatalf("select run row: %v", err)
	}
	if exitCode != 7 {
		t.Fatalf("exit_code=%d, want 7", exitCode)
	}
	if durationMs != 1500 {
		t.Fatalf("duration_ms=%d, want 1500", durationMs)
	}
}

func kvExists(ctx context.Context, db *sql.DB, runID string, k string) (bool, error) {
	var c int
	err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM kv WHERE run_id = ? AND step_id IS NULL AND command_id IS NULL AND k = ?;",
		runID,
		k,
	).Scan(&c)
	if err != nil {
		return false, err
	}
	return c > 0, nil
}


