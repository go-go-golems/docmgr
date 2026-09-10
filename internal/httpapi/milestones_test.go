package httpapi

import (
	"context"
	"encoding/json"
	"github.com/go-go-golems/docmgr/pkg/commands"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestMilestoneHTTPServiceParity(t *testing.T) {
	s := setupWriteTestServer(t)
	dir := filepath.Join("ttmp", "2026", "01", "03", "WRT-9--writes")
	mustWriteFile(t, filepath.Join(dir, "tasks.md"), "# Tasks\n\n- [ ] Verify <!-- t:ab12 -->\n")
	req := commands.MilestoneRequest{Ticket: "WRT-9", OperationID: "http-test", Summary: "Verified", Phase: "checkpoint", Next: "Review", TaskIDs: []string{"ab12"}}
	rr := doJSON(t, s, http.MethodPost, "/api/v1/tickets/milestone?dry_run=true", req)
	if rr.Code != 200 {
		t.Fatal(rr.Body.String())
	}
	if _, err := os.Stat(filepath.Join(dir, ".docmgr-operations")); !os.IsNotExist(err) {
		t.Fatal("dry run wrote")
	}
	rr = doJSON(t, s, http.MethodPost, "/api/v1/tickets/milestone", req)
	if rr.Code != 200 {
		t.Fatal(rr.Body.String())
	}
	receipt, err := commands.RecordMilestone(context.Background(), dir, req, false)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]json.RawMessage
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	var a, b any
	if err := json.Unmarshal(got["receipt"], &a); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(encoded, &b); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Fatal("service/HTTP receipts differ")
	}
	rr = doJSON(t, s, http.MethodGet, "/api/v1/tickets/resume?ticket=WRT-9", nil)
	if rr.Code != 200 {
		t.Fatal(rr.Body.String())
	}
	var view commands.ResumeView
	if err := json.Unmarshal(rr.Body.Bytes(), &view); err != nil {
		t.Fatal(err)
	}
	if len(view.Remaining) != 0 || view.Next != "Review" {
		t.Fatal(view)
	}
	req.Summary = "Different"
	rr = doJSON(t, s, http.MethodPost, "/api/v1/tickets/milestone", req)
	if rr.Code != 409 {
		t.Fatal("expected ID conflict", rr.Code)
	}
	req.OperationID = "bad-path"
	req.Evidence = []commands.EvidenceRef{{Kind: "file", Path: "doc://../../../../../secret", Revision: "bad", Claim: "x"}}
	rr = doJSON(t, s, http.MethodPost, "/api/v1/tickets/milestone", req)
	if rr.Code != 409 {
		t.Fatal("escape accepted", rr.Code)
	}
	rr = doJSON(t, s, http.MethodPost, "/api/v1/tickets/milestone", map[string]any{"ticket": "WRT-9", "unknown": true})
	if rr.Code != 400 {
		t.Fatal("unknown field accepted")
	}
}
