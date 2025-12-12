package workspace

import (
	"context"
	"path/filepath"
	"testing"
)

func TestWorkspaceQueryDocs_BasicFiltersAndReverseLookup(t *testing.T) {
	ctx := context.Background()

	repoRoot := t.TempDir()
	docsRoot := filepath.Join(repoRoot, "ttmp")
	ticketDir := filepath.Join(docsRoot, "2025", "12", "12", "MEN-1--x")

	// Ticket layout
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

	// Default query (ticket scope): control docs hidden, errors hidden => only index.md
	res, err := ws.QueryDocs(ctx, DocQuery{
		Scope: Scope{Kind: ScopeTicket, TicketID: "MEN-1"},
	})
	if err != nil {
		t.Fatalf("QueryDocs: %v", err)
	}
	if len(res.Docs) != 1 {
		t.Fatalf("expected 1 doc (index.md) with defaults, got %d", len(res.Docs))
	}
	if res.Docs[0].Doc == nil || res.Docs[0].Doc.Ticket != "MEN-1" {
		t.Fatalf("expected a valid doc with Ticket=MEN-1")
	}

	// Include control docs => index.md + tasks.md
	res, err = ws.QueryDocs(ctx, DocQuery{
		Scope: Scope{Kind: ScopeTicket, TicketID: "MEN-1"},
		Options: DocQueryOptions{
			IncludeControlDocs: true,
		},
	})
	if err != nil {
		t.Fatalf("QueryDocs include control: %v", err)
	}
	if len(res.Docs) != 2 {
		t.Fatalf("expected 2 docs with IncludeControlDocs=true, got %d", len(res.Docs))
	}

	// Include errors => now includes broken doc with ReadErr and Doc=nil.
	res, err = ws.QueryDocs(ctx, DocQuery{
		Scope: Scope{Kind: ScopeTicket, TicketID: "MEN-1"},
		Options: DocQueryOptions{
			IncludeControlDocs: true,
			IncludeErrors:      true,
		},
	})
	if err != nil {
		t.Fatalf("QueryDocs include errors: %v", err)
	}
	if len(res.Docs) != 3 {
		t.Fatalf("expected 3 docs with IncludeErrors=true, got %d", len(res.Docs))
	}
	var foundBroken bool
	for _, h := range res.Docs {
		if filepath.Base(h.Path) == "zz-broken.md" {
			foundBroken = true
			if h.Doc != nil {
				t.Fatalf("expected broken doc to have Doc=nil")
			}
			if h.ReadErr == nil {
				t.Fatalf("expected broken doc to have ReadErr")
			}
		}
	}
	if !foundBroken {
		t.Fatalf("expected broken doc to be present when IncludeErrors=true")
	}

	// TopicsAny: match topic "b" (case-insensitive) => should return index.md only.
	res, err = ws.QueryDocs(ctx, DocQuery{
		Scope: Scope{Kind: ScopeTicket, TicketID: "MEN-1"},
		Filters: DocFilters{
			TopicsAny: []string{"b"},
		},
		Options: DocQueryOptions{
			IncludeControlDocs: true,
		},
	})
	if err != nil {
		t.Fatalf("QueryDocs topics: %v", err)
	}
	if len(res.Docs) != 1 {
		t.Fatalf("expected 1 doc matching topic 'b', got %d", len(res.Docs))
	}
	if res.Docs[0].Doc == nil || res.Docs[0].Doc.DocType != "index" {
		t.Fatalf("expected topic 'b' match to be index.md doc")
	}

	// RelatedFile reverse lookup: match index.md by referenced file.
	res, err = ws.QueryDocs(ctx, DocQuery{
		Scope: Scope{Kind: ScopeRepo},
		Filters: DocFilters{
			RelatedFile: []string{"backend/main.go"},
		},
	})
	if err != nil {
		t.Fatalf("QueryDocs related file: %v", err)
	}
	if len(res.Docs) != 1 {
		t.Fatalf("expected 1 doc matching related file, got %d", len(res.Docs))
	}
	if res.Docs[0].Doc == nil || res.Docs[0].Doc.DocType != "index" {
		t.Fatalf("expected related file match to be index doc")
	}
}

func TestWorkspaceQueryDocs_ContradictoryScopeAndFilterIsError(t *testing.T) {
	ctx := context.Background()

	repoRoot := t.TempDir()
	docsRoot := filepath.Join(repoRoot, "ttmp")
	ticketDir := filepath.Join(docsRoot, "2025", "12", "12", "MEN-1--x")
	writeFile(t, filepath.Join(ticketDir, "index.md"), `---
Title: Ticket Index
Ticket: MEN-1
DocType: index
---
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

	_, err = ws.QueryDocs(ctx, DocQuery{
		Scope: Scope{Kind: ScopeTicket, TicketID: "MEN-1"},
		Filters: DocFilters{
			Ticket: "MEN-2",
		},
	})
	if err == nil {
		t.Fatalf("expected contradictory query to error")
	}
}


