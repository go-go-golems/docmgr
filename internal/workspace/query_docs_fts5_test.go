//go:build sqlite_fts5

package workspace

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestQueryDocs_TextQuery_OrderByRank(t *testing.T) {
	ctx := context.Background()

	repoRoot := t.TempDir()
	docsRoot := filepath.Join(repoRoot, "ttmp")
	if err := os.MkdirAll(docsRoot, 0o755); err != nil {
		t.Fatalf("mkdir docs root: %v", err)
	}

	ticketDir := filepath.Join(docsRoot, "2026", "01", "04", "TST-1--test")
	if err := os.MkdirAll(filepath.Join(ticketDir, "reference"), 0o755); err != nil {
		t.Fatalf("mkdir ticket dir: %v", err)
	}

	write := func(p string, s string) {
		t.Helper()
		if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
			t.Fatalf("write %s: %v", p, err)
		}
	}

	write(filepath.Join(ticketDir, "index.md"), `---
Title: Ticket Index
Ticket: TST-1
DocType: index
Status: active
Topics: [index]
---

# Index
`)

	write(filepath.Join(ticketDir, "reference", "01-a.md"), `---
Title: Alpha Title
Ticket: TST-1
DocType: reference
Status: active
Topics: [backend]
---

This document should match alpha via title.
`)

	write(filepath.Join(ticketDir, "reference", "02-b.md"), `---
Title: Other Title
Ticket: TST-1
DocType: reference
Status: active
Topics: [alpha]
---

This document should match alpha via topics.
`)

	ws, err := NewWorkspaceFromContext(WorkspaceContext{
		Root:      docsRoot,
		ConfigDir: repoRoot,
		RepoRoot:  repoRoot,
	})
	if err != nil {
		t.Fatalf("NewWorkspaceFromContext: %v", err)
	}
	if err := ws.InitIndex(ctx, BuildIndexOptions{IncludeBody: false}); err != nil {
		t.Fatalf("InitIndex: %v", err)
	}
	if !ws.FTSAvailable() {
		t.Fatalf("expected FTSAvailable() == true under sqlite_fts5 build tag")
	}

	res, err := ws.QueryDocs(ctx, DocQuery{
		Scope: Scope{Kind: ScopeRepo},
		Filters: DocFilters{
			TextQuery: "alpha",
		},
		Options: DocQueryOptions{
			IncludeArchivedPath: true,
			IncludeScriptsPath:  true,
			IncludeControlDocs:  true,
			OrderBy:             OrderByRank,
		},
	})
	if err != nil {
		t.Fatalf("QueryDocs: %v", err)
	}

	var gotTitles []string
	for _, h := range res.Docs {
		if h.Doc == nil {
			continue
		}
		gotTitles = append(gotTitles, h.Doc.Title)
	}
	if len(gotTitles) != 2 {
		t.Fatalf("expected 2 docs, got %d (%v)", len(gotTitles), gotTitles)
	}
}
