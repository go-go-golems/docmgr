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
	for _, table := range []string{"docs", "doc_topics", "doc_owners", "related_files"} {
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

	// Verify docs table has skill-specific columns.
	for _, col := range []string{"what_for", "when_to_use"} {
		var name string
		if err := db.QueryRowContext(ctx,
			`SELECT name FROM pragma_table_info('docs') WHERE name=?`,
			col,
		).Scan(&name); err != nil {
			t.Fatalf("column %q missing in docs table: %v", col, err)
		}
		if name != col {
			t.Fatalf("expected column %q, got %q", col, name)
		}
	}
}
