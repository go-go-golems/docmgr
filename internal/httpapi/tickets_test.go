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

- [ ] First task
- [x] Done task
`)

	mgr := NewIndexManager(root)
	if _, err := mgr.Refresh(context.Background()); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	s := NewServer(mgr, ServerOptions{})

	// Summary
	{
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tickets/get?ticket=TST-123", nil)
		rr := httptest.NewRecorder()
		s.Handler().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("tickets/get: expected %d, got %d (%s)", http.StatusOK, rr.Code, rr.Body.String())
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

	// Tasks list
	{
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tickets/tasks?ticket=TST-123", nil)
		rr := httptest.NewRecorder()
		s.Handler().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("tickets/tasks: expected %d, got %d (%s)", http.StatusOK, rr.Code, rr.Body.String())
		}
	}

	// Toggle task id=1 to checked.
	{
		body, _ := json.Marshal(map[string]any{
			"ticket":  "TST-123",
			"ids":     []int{1},
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
	}

	// Graph
	{
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tickets/graph?ticket=TST-123&direction=TD", nil)
		rr := httptest.NewRecorder()
		s.Handler().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("tickets/graph: expected %d, got %d (%s)", http.StatusOK, rr.Code, rr.Body.String())
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
