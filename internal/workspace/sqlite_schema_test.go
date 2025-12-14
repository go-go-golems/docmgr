package workspace

import (
	"context"
	"testing"
)

func TestCreateWorkspaceSchema_InMemory(t *testing.T) {
	ctx := context.Background()

	db, err := openInMemorySQLite(ctx)
	if err != nil {
		t.Fatalf("openInMemorySQLite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := createWorkspaceSchema(ctx, db); err != nil {
		t.Fatalf("createWorkspaceSchema: %v", err)
	}

	// Sanity: ensure key tables exist by querying sqlite_master.
	for _, table := range []string{"docs", "doc_topics", "related_files"} {
		var name string
		if err := db.QueryRowContext(ctx,
			`SELECT name FROM sqlite_master WHERE type='table' AND name=?`,
			table,
		).Scan(&name); err != nil {
			t.Fatalf("table %q missing (sqlite_master scan err): %v", table, err)
		}
		if name != table {
			t.Fatalf("expected table %q, got %q", table, name)
		}
	}
}
