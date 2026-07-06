package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestTicketsEndpoints_BasicFlow(t *testing.T) {
	t.Parallel()

	root := filepath.Join(t.TempDir(), "ttmp")
	ticketDir := filepath.Join(root, "2026", "01", "03", "TST-123--example")
	mustMkdirAll(t, filepath.Join(ticketDir, "design"))

	mustWriteFile(t, filepath.Join(ticketDir, "index.md"), `---
Title: Test Ticket
Ticket: TST-123
Status: active
DocType: index
Topics: [docmgr, test]
LastUpdated: 2026-01-05T00:00:00Z
RelatedFiles:
  - Path: internal/httpapi/server.go
    Note: server entry
---

# Test Ticket
`)
	mustWriteFile(t, filepath.Join(ticketDir, "design", "01-design.md"), `---
Title: Design Doc
Ticket: TST-123
Status: active
DocType: design-doc
RelatedFiles:
  - Path: ui/src/App.tsx
---

# Design
`)
	mustWriteFile(t, filepath.Join(ticketDir, "tasks.md"), `# Tasks

## TODO

- [ ] First task <!-- t:ab12 -->
- [x] Done task
`)

	mgr := NewIndexManager(root)
	if _, err := mgr.Refresh(context.Background()); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	s := NewServer(mgr, ServerOptions{})

	// Summary
	{
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tickets/get?ticket=TST-12", nil)
		rr := httptest.NewRecorder()
		s.Handler().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("tickets/get: expected %d, got %d (%s)", http.StatusOK, rr.Code, rr.Body.String())
		}
		var got ticketGetResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
			t.Fatalf("tickets/get short ref decode: %v", err)
		}
		if got.Ticket != "TST-123" || got.Stats.DocsTotal == 0 {
			t.Fatalf("tickets/get short ref used raw ticket instead of canonical ticket: %+v", got)
		}
	}

	// Docs
	{
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tickets/docs?ticket=TST-123&pageSize=50", nil)
		rr := httptest.NewRecorder()
		s.Handler().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("tickets/docs: expected %d, got %d (%s)", http.StatusOK, rr.Code, rr.Body.String())
		}
	}

	// Docs with a forgiving ticket prefix must query by the resolved canonical ID.
	{
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tickets/docs?ticket=TST-12&pageSize=50", nil)
		rr := httptest.NewRecorder()
		s.Handler().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("tickets/docs short ref: expected %d, got %d (%s)", http.StatusOK, rr.Code, rr.Body.String())
		}
		var got ticketDocsResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
			t.Fatalf("tickets/docs short ref decode: %v", err)
		}
		if got.Ticket != "TST-123" || got.Total == 0 {
			t.Fatalf("tickets/docs short ref used raw ticket instead of canonical ticket: %+v", got)
		}
	}

	// Tasks list
	{
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tickets/tasks?ticket=TST-123", nil)
		rr := httptest.NewRecorder()
		s.Handler().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("tickets/tasks: expected %d, got %d (%s)", http.StatusOK, rr.Code, rr.Body.String())
		}
	}

	// Toggle task stable ref to checked.
	{
		body, _ := json.Marshal(map[string]any{
			"ticket":  "TST-123",
			"refs":    []string{"ab12"},
			"checked": true,
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/tickets/tasks/check", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		s.Handler().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("tickets/tasks/check: expected %d, got %d (%s)", http.StatusOK, rr.Code, rr.Body.String())
		}

		req2 := httptest.NewRequest(http.MethodGet, "/api/v1/tickets/tasks?ticket=TST-123", nil)
		rr2 := httptest.NewRecorder()
		s.Handler().ServeHTTP(rr2, req2)
		if rr2.Code != http.StatusOK {
			t.Fatalf("tickets/tasks after check: expected %d, got %d (%s)", http.StatusOK, rr2.Code, rr2.Body.String())
		}
		if !bytes.Contains(rr2.Body.Bytes(), []byte(`"stableId":"ab12"`)) {
			t.Fatalf("tickets/tasks did not expose stableId after check: %s", rr2.Body.String())
		}
	}

	// Legacy positional ids still work.
	{
		body, _ := json.Marshal(map[string]any{
			"ticket":  "TST-123",
			"ids":     []int{1},
			"checked": false,
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/tickets/tasks/check", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		s.Handler().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("tickets/tasks/check legacy ids: expected %d, got %d (%s)", http.StatusOK, rr.Code, rr.Body.String())
		}
	}

	// Graph
	{
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tickets/graph?ticket=TST-12&direction=TD", nil)
		rr := httptest.NewRecorder()
		s.Handler().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("tickets/graph: expected %d, got %d (%s)", http.StatusOK, rr.Code, rr.Body.String())
		}
		var got ticketGraphResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
			t.Fatalf("tickets/graph short ref decode: %v", err)
		}
		if got.Ticket != "TST-123" {
			t.Fatalf("tickets/graph short ref used raw ticket instead of canonical ticket: %+v", got)
		}
	}
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
