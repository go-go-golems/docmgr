package scenariolog

import (
	"context"
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


