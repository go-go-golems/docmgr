package scenariolog

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
)

func SetKV(ctx context.Context, db *sql.DB, runID string, stepID string, commandID string, k string, v string) error {
	if runID == "" {
		return errors.New("SetKV: runID is required")
	}
	if k == "" {
		return errors.New("SetKV: key is required")
	}
	if v == "" {
		// Don't store empty values; keep semantics simple.
		return nil
	}

	_, err := db.ExecContext(ctx,
		`INSERT INTO kv (run_id, step_id, command_id, k, v)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(run_id, step_id, command_id, k) DO UPDATE SET v = excluded.v;`,
		runID,
		nullIfEmpty(stepID),
		nullIfEmpty(commandID),
		k,
		v,
	)
	if err != nil {
		return errors.Wrap(err, "upsert kv")
	}
	return nil
}


