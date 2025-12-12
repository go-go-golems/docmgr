package workspace

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestWorkspaceInitIndex_IngestsDocsTopicsAndRelatedFiles(t *testing.T) {
	ctx := context.Background()

	repoRoot := t.TempDir()
	docsRoot := filepath.Join(repoRoot, "ttmp")

	// Ticket layout
	ticketDir := filepath.Join(docsRoot, "2025", "12", "12", "MEN-1--x")
	writeFile(t, filepath.Join(ticketDir, "index.md"), `---
Title: Ticket Index
Ticket: MEN-1
Status: active
Topics: [a, B]
DocType: index
Intent: long-term
Owners: [manuel]
RelatedFiles:
  - Path: backend/main.go
    Note: entrypoint
LastUpdated: 2025-12-12T00:00:00Z
---

# Index
`)
	writeFile(t, filepath.Join(ticketDir, "tasks.md"), `---
Title: Tasks
Ticket: MEN-1
Status: active
Topics: [a]
DocType: tasks
Intent: long-term
Owners: [manuel]
LastUpdated: 2025-12-12T00:00:00Z
---

# Tasks
`)

	// Should be skipped: .meta/
	writeFile(t, filepath.Join(ticketDir, ".meta", "internal.md"), `---
Title: Meta
Ticket: MEN-1
DocType: reference
---
`)
	// Should be skipped: _guidelines/
	writeFile(t, filepath.Join(docsRoot, "_guidelines", "reference.md"), `---
Title: Guideline
Ticket: MEN-1
DocType: reference
---
`)

	// Broken frontmatter doc (parse error) should be indexed with parse_ok=0
	writeFile(t, filepath.Join(ticketDir, "reference", "zz-broken.md"), `---
Title: Broken
Ticket: MEN-1
DocType: reference
Topics: [a
---
broken
`)

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
	db := ws.DB()
	if db == nil {
		t.Fatalf("expected db to be initialized")
	}

	// docs should include index.md, tasks.md, and broken doc; but exclude .meta and _guidelines
	var docsCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM docs`).Scan(&docsCount); err != nil {
		t.Fatalf("count docs: %v", err)
	}
	if docsCount != 3 {
		t.Fatalf("expected 3 docs indexed, got %d", docsCount)
	}

	var skippedCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM docs WHERE path LIKE '%/.meta/%' OR path LIKE '%/_guidelines/%'`).Scan(&skippedCount); err != nil {
		t.Fatalf("count skipped paths: %v", err)
	}
	if skippedCount != 0 {
		t.Fatalf("expected 0 docs under skipped dirs, got %d", skippedCount)
	}

	// tasks.md should be tagged as control doc (sibling index.md exists).
	var controlCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM docs WHERE path LIKE '%/tasks.md' AND is_control_doc=1`).Scan(&controlCount); err != nil {
		t.Fatalf("count control docs: %v", err)
	}
	if controlCount != 1 {
		t.Fatalf("expected tasks.md to be is_control_doc=1, got %d", controlCount)
	}

	// Broken doc parse_ok should be 0 and parse_err non-empty.
	var parseOK int
	var parseErr string
	if err := db.QueryRowContext(ctx, `SELECT parse_ok, COALESCE(parse_err,'') FROM docs WHERE path LIKE '%/reference/zz-broken.md'`).Scan(&parseOK, &parseErr); err != nil {
		t.Fatalf("select broken doc: %v", err)
	}
	if parseOK != 0 {
		t.Fatalf("expected broken doc parse_ok=0, got %d", parseOK)
	}
	if parseErr == "" {
		t.Fatalf("expected broken doc parse_err to be non-empty")
	}

	// Related files row should exist for index.md.
	var rfCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM related_files WHERE raw_path='backend/main.go'`).Scan(&rfCount); err != nil {
		t.Fatalf("count related_files: %v", err)
	}
	if rfCount != 1 {
		t.Fatalf("expected 1 related_files row for backend/main.go, got %d", rfCount)
	}

	// Topics should be lowercased in doc_topics.
	var topicCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM doc_topics WHERE topic_lower='b'`).Scan(&topicCount); err != nil {
		t.Fatalf("count doc_topics: %v", err)
	}
	if topicCount != 1 {
		t.Fatalf("expected topic_lower='b' to exist once, got %d", topicCount)
	}
}

func writeFile(t *testing.T, path string, contents string) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}


