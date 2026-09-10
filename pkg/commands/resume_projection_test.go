package commands

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestResumeUsesLatestRecordedProjectionIncludingClose(t *testing.T) {
	dir := ticketFixture(t)
	ctx := context.Background()
	_, err := RecordMilestone(ctx, dir, MilestoneRequest{Ticket: "TEST", OperationID: "verified", Summary: "Verified", Phase: "close", TaskIDs: []string{"ab12"}}, false)
	mustMilestone(t, err)
	_, err = CloseTicket(ctx, dir, "closed", CloseRequest{Ticket: "TEST"})
	mustMilestone(t, err)
	view, err := TicketResume(ctx, dir)
	mustMilestone(t, err)
	if view.Status != "complete" || len(view.Conflicts) != 0 {
		t.Fatalf("recorded close produced stale warning: %+v", view)
	}
	mustMilestone(t, os.WriteFile(filepath.Join(dir, "changelog.md"), []byte("unrecorded edit"), 0644))
	view, err = TicketResume(ctx, dir)
	mustMilestone(t, err)
	if len(view.Conflicts) != 1 {
		t.Fatal("unrecorded edit not detected")
	}
}
