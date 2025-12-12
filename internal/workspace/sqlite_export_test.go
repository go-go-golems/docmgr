package workspace

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestExportIndexToSQLiteFile_CreatesREADME(t *testing.T) {
	ctx := context.Background()

	repoRoot := t.TempDir()
	docsRoot := filepath.Join(repoRoot, "ttmp")
	ticketDir := filepath.Join(docsRoot, "2025", "12", "12", "MEN-1--x")
	if err := os.MkdirAll(ticketDir, 0o755); err != nil {
		t.Fatalf("mkdir ticketDir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ticketDir, "index.md"), []byte(`---
Title: Ticket Index
Ticket: MEN-1
DocType: index
Topics: [a]
---
body
`), 0o644); err != nil {
		t.Fatalf("write index.md: %v", err)
	}

	ws, err := NewWorkspaceFromContext(WorkspaceContext{
		Root:      docsRoot,
		ConfigDir: repoRoot,
		RepoRoot:  repoRoot,
	})
	if err != nil {
		t.Fatalf("NewWorkspaceFromContext: %v", err)
	}
	if err := ws.InitIndex(ctx, BuildIndexOptions{}); err != nil {
		t.Fatalf("InitIndex: %v", err)
	}

	outPath := filepath.Join(repoRoot, "export.sqlite")
	if err := ws.ExportIndexToSQLiteFile(ctx, ExportSQLiteOptions{OutPath: outPath, Force: true}); err != nil {
		t.Fatalf("ExportIndexToSQLiteFile: %v", err)
	}

	db, err := sql.Open("sqlite3", outPath)
	if err != nil {
		t.Fatalf("open exported sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	// README table exists and has at least the __about__.md record.
	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM README`).Scan(&count); err != nil {
		t.Fatalf("count README: %v", err)
	}
	if count < 1 {
		t.Fatalf("expected README rows >= 1, got %d", count)
	}

	var about string
	if err := db.QueryRowContext(ctx, `SELECT content FROM README WHERE name='__about__.md'`).Scan(&about); err != nil {
		t.Fatalf("select about: %v", err)
	}
	if about == "" {
		t.Fatalf("expected __about__.md content to be non-empty")
	}
}


