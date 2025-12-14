//go:build sqlite_fts5

package scenariolog

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSearchFTSFindsIndexedLines(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	root := t.TempDir()
	dbPath := filepath.Join(root, "run.db")

	db, err := Open(ctx, dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = db.Close() }()

	if err := Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	runID := "run-fts-1"
	now := time.Now().UTC()
	if err := StartRun(ctx, db, runID, root, "suite", now); err != nil {
		t.Fatalf("StartRun: %v", err)
	}

	// Create a step row (minimal) and artifact row, then index the artifact.
	stepID := runID + "-step-01"
	_, err = db.ExecContext(ctx,
		`INSERT INTO steps (step_id, run_id, step_num, step_name, started_at) VALUES (?, ?, ?, ?, ?);`,
		stepID, runID, 1, "step", now.Format(time.RFC3339Nano))
	if err != nil {
		t.Fatalf("insert step: %v", err)
	}

	abs := filepath.Join(root, "stdout.txt")
	if err := os.WriteFile(abs, []byte("hello world\nwarning: something\n"), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
	sha, size, err := fileSHA256AndSize(abs)
	if err != nil {
		t.Fatalf("sha: %v", err)
	}
	artifactID, err := insertArtifact(ctx, db, runID, stepID, "", "stdout", "stdout.txt", true, size, sha)
	if err != nil {
		t.Fatalf("insertArtifact: %v", err)
	}
	if err := indexArtifactLinesFTS(ctx, db, runID, artifactID, abs); err != nil {
		t.Fatalf("indexArtifactLinesFTS: %v", err)
	}

	hits, err := SearchFTS(ctx, db, runID, "warning", 10)
	if err != nil {
		t.Fatalf("SearchFTS: %v", err)
	}
	if len(hits) == 0 {
		t.Fatalf("expected at least 1 hit")
	}
	found := false
	for _, h := range hits {
		if h.ArtifactID == artifactID && h.LineNum == 2 {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected to find hit at artifact=%d line=2; got %+v", artifactID, hits)
	}
}


