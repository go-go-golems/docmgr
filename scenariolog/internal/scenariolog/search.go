package scenariolog

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
)

var ErrFTSNotAvailable = errors.New("fts5 not available (log_lines_fts missing)")

type SearchHit struct {
	ArtifactID int64
	LineNum    int
	Text       string
}

func SearchFTS(ctx context.Context, db *sql.DB, runID string, query string, limit int) ([]SearchHit, error) {
	if runID == "" {
		return nil, errors.New("SearchFTS: runID is required")
	}
	if query == "" {
		return nil, errors.New("SearchFTS: query is required")
	}
	if limit <= 0 {
		limit = 100
	}

	ok, err := hasTable(ctx, db, "log_lines_fts")
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrFTSNotAvailable
	}

	rows, err := db.QueryContext(ctx,
		`SELECT artifact_id, line_num, text
		 FROM log_lines_fts
		 WHERE run_id = ?
		   AND log_lines_fts MATCH ?
		 LIMIT ?;`,
		runID,
		query,
		limit,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query log_lines_fts")
	}
	defer func() { _ = rows.Close() }()

	hits := []SearchHit{}
	for rows.Next() {
		var h SearchHit
		if err := rows.Scan(&h.ArtifactID, &h.LineNum, &h.Text); err != nil {
			return nil, errors.Wrap(err, "scan fts row")
		}
		hits = append(hits, h)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "iterate fts rows")
	}
	return hits, nil
}


