package scenariolog

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

func StartRun(ctx context.Context, db *sql.DB, runID string, rootDir string, suite string, startedAt time.Time) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO scenario_runs (run_id, root_dir, suite, started_at) VALUES (?, ?, ?, ?);`,
		runID,
		rootDir,
		nullIfEmpty(suite),
		startedAt.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return errors.Wrap(err, "insert scenario_runs")
	}
	return nil
}

func EndRun(ctx context.Context, db *sql.DB, runID string, exitCode int, completedAt time.Time) error {
	startedAt, err := getRunStartedAt(ctx, db, runID)
	if err != nil {
		return err
	}

	durationMs := int64(completedAt.Sub(startedAt).Milliseconds())
	if durationMs < 0 {
		// Clock weirdness shouldn't break completion.
		durationMs = 0
	}

	_, err = db.ExecContext(ctx,
		`UPDATE scenario_runs
		 SET completed_at = ?, exit_code = ?, duration_ms = ?
		 WHERE run_id = ?;`,
		completedAt.UTC().Format(time.RFC3339Nano),
		exitCode,
		durationMs,
		runID,
	)
	if err != nil {
		return errors.Wrap(err, "update scenario_runs completion")
	}
	return nil
}

func getRunStartedAt(ctx context.Context, db *sql.DB, runID string) (time.Time, error) {
	var s string
	err := db.QueryRowContext(ctx, `SELECT started_at FROM scenario_runs WHERE run_id = ?;`, runID).Scan(&s)
	if err != nil {
		return time.Time{}, errors.Wrap(err, "select scenario_runs.started_at")
	}

	// Stored as RFC3339Nano by StartRun. Be tolerant if it was written differently.
	t, err := time.Parse(time.RFC3339Nano, s)
	if err == nil {
		return t, nil
	}
	t, err2 := time.Parse(time.RFC3339, s)
	if err2 == nil {
		return t, nil
	}
	return time.Time{}, errors.Wrapf(err, "parse started_at timestamp: %q", s)
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}


