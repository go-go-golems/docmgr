package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupWriteTestServer builds a temp repo (go.mod + .ttmp.yaml + ttmp docs
// root + a source file) and chdirs into it so repo-root discovery — and thus
// anchored related-file writes — are deterministic. Tests using it must NOT
// call t.Parallel (they change the process working directory).
func setupWriteTestServer(t *testing.T) *Server {
	t.Helper()

	repo := t.TempDir()
	root := filepath.Join(repo, "ttmp")
	ticketDir := filepath.Join(root, "2026", "01", "03", "WRT-9--writes")
	mustMkdirAll(t, ticketDir)
	mustMkdirAll(t, filepath.Join(repo, "src"))

	mustWriteFile(t, filepath.Join(repo, "go.mod"), "module example.com/writetest\n\ngo 1.23\n")
	mustWriteFile(t, filepath.Join(repo, ".ttmp.yaml"), "root: ttmp\n")
	mustWriteFile(t, filepath.Join(repo, "src", "main.go"), "package main\n")

	mustWriteFile(t, filepath.Join(ticketDir, "index.md"), `---
Title: Write Endpoints
Ticket: WRT-9
Status: active
DocType: index
Topics: [docmgr]
LastUpdated: 2026-01-05T00:00:00Z
---

# Write Endpoints
`)

	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(repo); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldCwd) })

	mgr := NewIndexManager(root)
	if _, err := mgr.Refresh(context.Background()); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	return NewServer(mgr, ServerOptions{})
}

func doJSON(t *testing.T, s *Server, method, url string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var reader *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(b)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, url, reader)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)
	return rr
}

func TestDocsMeta_WriteAndReadBack(t *testing.T) {
	s := setupWriteTestServer(t)

	docPath := "2026/01/03/WRT-9--writes/index.md"

	rr := doJSON(t, s, http.MethodPost, "/api/v1/docs/meta", map[string]any{
		"path":  docPath,
		"field": "Status",
		"value": "review",
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("docs/meta: expected %d, got %d (%s)", http.StatusOK, rr.Code, rr.Body.String())
	}

	// Read back through the (refreshed) API.
	rr2 := doJSON(t, s, http.MethodGet, "/api/v1/docs/get?path="+docPath, nil)
	if rr2.Code != http.StatusOK {
		t.Fatalf("docs/get: expected %d, got %d (%s)", http.StatusOK, rr2.Code, rr2.Body.String())
	}
	var got struct {
		Doc struct {
			Status string `json:"status"`
		} `json:"doc"`
	}
	if err := json.Unmarshal(rr2.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal docs/get: %v", err)
	}
	if got.Doc.Status != "review" {
		t.Fatalf("expected status review, got %q", got.Doc.Status)
	}
}

func TestDocsMeta_UnknownFieldIs400(t *testing.T) {
	s := setupWriteTestServer(t)

	rr := doJSON(t, s, http.MethodPost, "/api/v1/docs/meta", map[string]any{
		"path":  "2026/01/03/WRT-9--writes/index.md",
		"field": "NotAField",
		"value": "x",
	})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d (%s)", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
}

func TestDocsMeta_ForbidsTraversal(t *testing.T) {
	s := setupWriteTestServer(t)

	rr := doJSON(t, s, http.MethodPost, "/api/v1/docs/meta", map[string]any{
		"path":  "../outside.md",
		"field": "Status",
		"value": "review",
	})
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected %d, got %d (%s)", http.StatusForbidden, rr.Code, rr.Body.String())
	}
}

func TestDocsRelate_WritesAnchoredEntry(t *testing.T) {
	s := setupWriteTestServer(t)

	docPath := "2026/01/03/WRT-9--writes/index.md"

	rr := doJSON(t, s, http.MethodPost, "/api/v1/docs/relate", map[string]any{
		"path": docPath,
		"add":  []map[string]string{{"path": "src/main.go", "note": "entry point"}},
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("docs/relate: expected %d, got %d (%s)", http.StatusOK, rr.Code, rr.Body.String())
	}
	var relateResp struct {
		Added  int    `json:"added"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &relateResp); err != nil {
		t.Fatalf("unmarshal relate: %v", err)
	}
	if relateResp.Added != 1 || relateResp.Status != "updated" {
		t.Fatalf("expected added=1/updated, got %+v (%s)", relateResp, rr.Body.String())
	}

	rr2 := doJSON(t, s, http.MethodGet, "/api/v1/docs/get?path="+docPath, nil)
	if rr2.Code != http.StatusOK {
		t.Fatalf("docs/get: expected %d, got %d (%s)", http.StatusOK, rr2.Code, rr2.Body.String())
	}
	var got struct {
		RelatedFiles []struct {
			Path         string `json:"path"`
			Note         string `json:"note"`
			Root         string `json:"root"`
			ResolvedPath string `json:"resolvedPath"`
			Exists       bool   `json:"exists"`
		} `json:"relatedFiles"`
	}
	if err := json.Unmarshal(rr2.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal docs/get: %v", err)
	}
	if len(got.RelatedFiles) != 1 {
		t.Fatalf("expected 1 related file, got %d (%s)", len(got.RelatedFiles), rr2.Body.String())
	}
	rf := got.RelatedFiles[0]
	if rf.Path != "repo://src/main.go" {
		t.Fatalf("expected anchored path repo://src/main.go, got %q", rf.Path)
	}
	if rf.Note != "entry point" || rf.Root != "repo" || rf.ResolvedPath != "src/main.go" || !rf.Exists {
		t.Fatalf("unexpected related file resolution: %+v", rf)
	}

	// Remove it again.
	rr3 := doJSON(t, s, http.MethodPost, "/api/v1/docs/relate", map[string]any{
		"path":   docPath,
		"remove": []string{"src/main.go"},
	})
	if rr3.Code != http.StatusOK {
		t.Fatalf("docs/relate remove: expected %d, got %d (%s)", http.StatusOK, rr3.Code, rr3.Body.String())
	}
	var removeResp struct {
		Removed int `json:"removed"`
		Total   int `json:"total"`
	}
	if err := json.Unmarshal(rr3.Body.Bytes(), &removeResp); err != nil {
		t.Fatalf("unmarshal remove: %v", err)
	}
	if removeResp.Removed != 1 || removeResp.Total != 0 {
		t.Fatalf("expected removed=1 total=0, got %+v", removeResp)
	}
}

func TestTicketsChangelog_AppendAndParse(t *testing.T) {
	s := setupWriteTestServer(t)

	// No changelog.md yet.
	rr := doJSON(t, s, http.MethodGet, "/api/v1/tickets/changelog?ticket=WRT-9", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("changelog get: expected %d, got %d (%s)", http.StatusOK, rr.Code, rr.Body.String())
	}
	var empty struct {
		Exists  bool  `json:"exists"`
		Entries []any `json:"entries"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &empty); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if empty.Exists || len(empty.Entries) != 0 {
		t.Fatalf("expected exists=false with no entries, got %s", rr.Body.String())
	}

	// Append.
	rr2 := doJSON(t, s, http.MethodPost, "/api/v1/tickets/changelog", map[string]any{
		"ticket": "WRT-9",
		"title":  "Testing",
		"entry":  "Implemented the write endpoints.",
	})
	if rr2.Code != http.StatusOK {
		t.Fatalf("changelog post: expected %d, got %d (%s)", http.StatusOK, rr2.Code, rr2.Body.String())
	}

	// Read back parsed entries.
	rr3 := doJSON(t, s, http.MethodGet, "/api/v1/tickets/changelog?ticket=WRT-9", nil)
	if rr3.Code != http.StatusOK {
		t.Fatalf("changelog get 2: expected %d, got %d (%s)", http.StatusOK, rr3.Code, rr3.Body.String())
	}
	var got struct {
		Exists  bool `json:"exists"`
		Entries []struct {
			Date  string `json:"date"`
			Title string `json:"title"`
			Body  string `json:"body"`
		} `json:"entries"`
	}
	if err := json.Unmarshal(rr3.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !got.Exists || len(got.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %s", rr3.Body.String())
	}
	e := got.Entries[0]
	if e.Title != "Testing" || e.Date == "" || !strings.Contains(e.Body, "Implemented the write endpoints.") {
		t.Fatalf("unexpected entry: %+v", e)
	}

	// Empty entry rejected.
	rr4 := doJSON(t, s, http.MethodPost, "/api/v1/tickets/changelog", map[string]any{
		"ticket": "WRT-9",
		"entry":  "  ",
	})
	if rr4.Code != http.StatusBadRequest {
		t.Fatalf("expected %d for empty entry, got %d (%s)", http.StatusBadRequest, rr4.Code, rr4.Body.String())
	}
}

func TestWorkspaceDoctor_ReturnsFindingsAndRollup(t *testing.T) {
	s := setupWriteTestServer(t)

	rr := doJSON(t, s, http.MethodGet, "/api/v1/workspace/doctor?ticket=WRT-9", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("doctor: expected %d, got %d (%s)", http.StatusOK, rr.Code, rr.Body.String())
	}
	var got struct {
		Ticket string `json:"ticket"`
		Totals struct {
			Findings int `json:"findings"`
		} `json:"totals"`
		Rollup []struct {
			Ticket string `json:"ticket"`
			Status string `json:"status"`
		} `json:"rollup"`
		Findings []struct {
			Ticket   string `json:"ticket"`
			Issue    string `json:"issue"`
			Severity string `json:"severity"`
		} `json:"findings"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal doctor: %v", err)
	}
	if got.Ticket != "WRT-9" {
		t.Fatalf("expected ticket WRT-9, got %q", got.Ticket)
	}
	if got.Totals.Findings == 0 || len(got.Findings) == 0 {
		t.Fatalf("expected at least one finding, got %s", rr.Body.String())
	}
	found := false
	for _, item := range got.Rollup {
		if item.Ticket == "WRT-9" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected rollup entry for WRT-9, got %s", rr.Body.String())
	}

	// Unknown ticket -> 404.
	rr2 := doJSON(t, s, http.MethodGet, "/api/v1/workspace/doctor?ticket=NOPE-404", nil)
	if rr2.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d (%s)", http.StatusNotFound, rr2.Code, rr2.Body.String())
	}
}
