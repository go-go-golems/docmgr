package skills

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/docmgr/internal/workspace"
)

func TestDiscoverPlansWorkspaceAndTicket(t *testing.T) {
	ctx := context.Background()
	rootDir := t.TempDir()
	docsRoot := filepath.Join(rootDir, "ttmp")
	if err := os.MkdirAll(docsRoot, 0o755); err != nil {
		t.Fatalf("mkdir docs root: %v", err)
	}

	ws, err := workspace.NewWorkspaceFromContext(workspace.WorkspaceContext{
		Root:      docsRoot,
		ConfigDir: rootDir,
		RepoRoot:  rootDir,
	})
	if err != nil {
		t.Fatalf("workspace: %v", err)
	}

	workspacePlanDir := filepath.Join(docsRoot, "skills", "glaze-help")
	if err := os.MkdirAll(workspacePlanDir, 0o755); err != nil {
		t.Fatalf("mkdir workspace plan dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workspacePlanDir, "skill.yaml"), []byte(`skill:
  name: glaze-help
  description: Help topics for Glazed.
  what_for: Provide help output.
  when_to_use: Use when working with Glazed.
  topics: [glaze]

sources: []
`), 0o644); err != nil {
		t.Fatalf("write workspace plan: %v", err)
	}

	ticketDir := filepath.Join(docsRoot, "2026", "01", "01", "TEST-123--sample")
	if err := os.MkdirAll(ticketDir, 0o755); err != nil {
		t.Fatalf("mkdir ticket dir: %v", err)
	}
	index := []byte(`---
Title: Sample Ticket
Ticket: TEST-123
DocType: index
Status: active
---
# Sample
`)
	if err := os.WriteFile(filepath.Join(ticketDir, "index.md"), index, 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}

	ticketPlanDir := filepath.Join(ticketDir, "skills", "ticket-help")
	if err := os.MkdirAll(ticketPlanDir, 0o755); err != nil {
		t.Fatalf("mkdir ticket plan dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ticketPlanDir, "skill.yaml"), []byte(`skill:
  name: ticket-help
  description: Ticket help.
  what_for: Provide ticket help.
  when_to_use: Use when working with ticket help.
  topics: [ticket]

sources: []
`), 0o644); err != nil {
		t.Fatalf("write ticket plan: %v", err)
	}

	if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{IncludeBody: false}); err != nil {
		t.Fatalf("init index: %v", err)
	}

	handles, err := DiscoverPlans(ctx, ws, DiscoverOptions{TicketID: "TEST-123", IncludeWorkspace: true})
	if err != nil {
		t.Fatalf("discover plans: %v", err)
	}
	if len(handles) != 2 {
		t.Fatalf("expected 2 plans, got %d", len(handles))
	}

	foundTicket := false
	for _, handle := range handles {
		if handle.TicketID == "TEST-123" {
			foundTicket = true
		}
	}
	if !foundTicket {
		t.Fatalf("expected ticket plan handle")
	}
}
