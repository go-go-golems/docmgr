package scenariolog

import (
	"bufio"
	"context"
	"database/sql"
	"os"

	"github.com/pkg/errors"
)

// indexArtifactLinesFTS indexes a text artifact into log_lines_fts if the table exists.
// In degraded mode (no FTS5), this is a no-op.
func indexArtifactLinesFTS(ctx context.Context, db *sql.DB, runID string, artifactID int64, absPath string) error {
	ok, err := hasTable(ctx, db, "log_lines_fts")
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	f, err := os.Open(absPath)
	if err != nil {
		return errors.Wrap(err, "open artifact for fts indexing")
	}
	defer func() { _ = f.Close() }()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "begin fts tx")
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO log_lines_fts (run_id, artifact_id, line_num, text) VALUES (?, ?, ?, ?);`)
	if err != nil {
		return errors.Wrap(err, "prepare fts insert")
	}
	defer func() { _ = stmt.Close() }()

	sc := bufio.NewScanner(f)
	// Default token limit is 64K; bump to tolerate long lines but keep a cap.
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)

	lineNum := 0
	for sc.Scan() {
		lineNum++
		text := sc.Text()
		// Keep inserts small and predictable; truncate extreme lines.
		if len(text) > 16*1024 {
			text = text[:16*1024]
		}
		if _, err := stmt.ExecContext(ctx, runID, artifactID, lineNum, text); err != nil {
			return errors.Wrap(err, "fts insert line")
		}
	}
	if err := sc.Err(); err != nil {
		return errors.Wrap(err, "scan artifact for fts")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "commit fts tx")
	}
	return nil
}

func hasTable(ctx context.Context, db *sql.DB, name string) (bool, error) {
	var c int
	err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM sqlite_master WHERE type IN ('table','view') AND name = ?;",
		name,
	).Scan(&c)
	if err != nil {
		return false, errors.Wrap(err, "sqlite_master lookup")
	}
	return c > 0, nil
}

