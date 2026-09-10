package commands

import (
	"context"
	"errors"
	"github.com/go-go-golems/docmgr/internal/documents"
	"github.com/go-go-golems/docmgr/internal/operations"
	"github.com/go-go-golems/docmgr/pkg/models"
	"os"
	"path/filepath"
	"testing"
)

func mustMilestone(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
func ticketFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	mustMilestone(t, documents.WriteDocumentWithFrontmatter(filepath.Join(dir, "index.md"), &models.Document{Title: "Test", Ticket: "TEST", Status: "active", DocType: "index"}, "\n# Test\n", true))
	mustMilestone(t, os.WriteFile(filepath.Join(dir, "tasks.md"), []byte("# Tasks\n\n- [ ] Test <!-- t:ab12 -->\n"), 0644))
	return dir
}
func TestMilestoneResumeAndStaleness(t *testing.T) {
	dir := ticketFixture(t)
	ctx := context.Background()
	evidence := []byte("PASS\n")
	mustMilestone(t, os.WriteFile(filepath.Join(dir, "proof.txt"), evidence, 0644))
	req := MilestoneRequest{Ticket: "TEST", OperationID: "verified", Summary: "Tests pass", Phase: "validate", Next: "Review", TaskIDs: []string{"ab12"}, Evidence: []EvidenceRef{{Kind: "test", Path: "doc://proof.txt", Revision: operations.Hash(evidence), Claim: "fixture passed"}}}
	_, err := RecordMilestone(ctx, dir, req, true)
	mustMilestone(t, err)
	r, err := RecordMilestone(ctx, dir, req, false)
	mustMilestone(t, err)
	if r.State != "committed" {
		t.Fatal(r)
	}
	view, err := TicketResume(ctx, dir)
	mustMilestone(t, err)
	if len(view.Remaining) != 0 || len(view.Conflicts) != 0 || view.Phase != "validate" {
		t.Fatalf("bad resume: %+v", view)
	}
	mustMilestone(t, os.WriteFile(filepath.Join(dir, "proof.txt"), []byte("changed"), 0644))
	view, err = TicketResume(ctx, dir)
	mustMilestone(t, err)
	if len(view.Conflicts) != 1 {
		t.Fatal("missed stale evidence")
	}
	_, err = RecordMilestone(ctx, dir, req, false)
	mustMilestone(t, err) // replay returns historical receipt, not new verification
	req.OperationID = "bad"
	req.TaskIDs = []string{"1"}
	if _, err := RecordMilestone(ctx, dir, req, false); err == nil {
		t.Fatal("bad request accepted")
	}
}
func TestClosePreflightAndRecovery(t *testing.T) {
	ctx := context.Background()
	req := CloseRequest{Ticket: "TEST", Status: "complete", Entry: "Ticket closed"}
	t.Run("directory preflight", func(t *testing.T) {
		dir := ticketFixture(t)
		mustMilestone(t, os.Mkdir(filepath.Join(dir, "changelog.md"), 0755))
		if _, err := CloseTicket(ctx, dir, "close1", req); err == nil {
			t.Fatal("expected error")
		}
		doc, _, err := documents.ReadDocumentWithFrontmatter(filepath.Join(dir, "index.md"))
		mustMilestone(t, err)
		if doc.Status != "active" {
			t.Fatal("partial complete")
		}
	})
	t.Run("index write recovery", func(t *testing.T) {
		dir := ticketFixture(t)
		_, err := closeTicket(ctx, dir, "close1", req, operations.Store{BeforeWrite: func(path string) error {
			if path == "index.md" {
				return errors.New("index failure")
			}
			return nil
		}})
		if err == nil {
			t.Fatal("expected fault")
		}
		doc, _, err := documents.ReadDocumentWithFrontmatter(filepath.Join(dir, "index.md"))
		mustMilestone(t, err)
		if doc.Status != "active" {
			t.Fatal("status changed")
		}
		r, err := CloseTicket(ctx, dir, "close1", req)
		mustMilestone(t, err)
		if r.Status != "complete" || r.OpenTasks != 1 {
			t.Fatal(r)
		}
		before, _, err := operations.Read(dir, "changelog.md")
		mustMilestone(t, err)
		r, err = CloseTicket(ctx, dir, "", req)
		mustMilestone(t, err)
		if r.Receipt.State != "unchanged" {
			t.Fatal(r)
		}
		after, _, err := operations.Read(dir, "changelog.md")
		mustMilestone(t, err)
		if string(before) != string(after) {
			t.Fatal("duplicate history")
		}
	})
}
