package scenariolog

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
)

func insertArtifact(
	ctx context.Context,
	db *sql.DB,
	runID string,
	stepID string,
	commandID string,
	kind string,
	path string,
	isText bool,
	sizeBytes int64,
	sha256 string,
) (int64, error) {
	if runID == "" {
		return 0, errors.New("insertArtifact: runID is required")
	}
	if stepID == "" {
		return 0, errors.New("insertArtifact: stepID is required")
	}
	if kind == "" {
		return 0, errors.New("insertArtifact: kind is required")
	}
	if path == "" {
		return 0, errors.New("insertArtifact: path is required")
	}

	isTextInt := 0
	if isText {
		isTextInt = 1
	}

	res, err := db.ExecContext(ctx,
		`INSERT INTO artifacts (run_id, step_id, command_id, kind, path, is_text, size_bytes, sha256)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?);`,
		runID,
		stepID,
		nullIfEmpty(commandID),
		kind,
		path,
		isTextInt,
		sizeBytes,
		nullIfEmpty(sha256),
	)
	if err != nil {
		return 0, errors.Wrap(err, "insert artifacts")
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, errors.Wrap(err, "artifacts LastInsertId")
	}
	return id, nil
}


