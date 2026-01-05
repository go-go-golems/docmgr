package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestWorkspaceEndpoints_BasicFlow(t *testing.T) {
	t.Parallel()

	root := filepath.Join(t.TempDir(), "ttmp")

	ticket1Dir := filepath.Join(root, "2026", "01", "03", "TST-123--example")
	mustMkdirAll(t, filepath.Join(ticket1Dir, "design"))
	mustWriteFile(t, filepath.Join(ticket1Dir, "index.md"), `---
Title: Test Ticket 123
Ticket: TST-123
Status: active
DocType: index
Topics: [docmgr, test]
Owners: [manuel]
Intent: long-term
LastUpdated: 2026-01-05T00:00:00Z
---

# Index
Hello world
`)
	mustWriteFile(t, filepath.Join(ticket1Dir, "design", "01-design.md"), `---
Title: Design 123
Ticket: TST-123
Status: active
DocType: design-doc
Topics: [docmgr]
Owners: [manuel]
Intent: long-term
LastUpdated: 2026-01-05T00:01:00Z
---

# Design
`)

	ticket2Dir := filepath.Join(root, "2026", "01", "04", "TST-999--done")
	mustMkdirAll(t, filepath.Join(ticket2Dir, "reference"))
	mustWriteFile(t, filepath.Join(ticket2Dir, "index.md"), `---
Title: Test Ticket 999
Ticket: TST-999
Status: complete
DocType: index
Topics: [other]
Owners: [alex]
Intent: short-term
LastUpdated: 2026-01-04T00:00:00Z
---

# Index
`)
	mustWriteFile(t, filepath.Join(ticket2Dir, "reference", "01-ref.md"), `---
Title: Ref 999
Ticket: TST-999
Status: complete
DocType: reference
Topics: [other]
Owners: [alex]
Intent: short-term
LastUpdated: 2026-01-04T00:10:00Z
---

# Ref
`)

	mgr := NewIndexManager(root)
	if _, err := mgr.Refresh(context.Background()); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	s := NewServer(mgr, ServerOptions{})

	// /workspace/summary
	{
		req := httptest.NewRequest(http.MethodGet, "/api/v1/workspace/summary", nil)
		rr := httptest.NewRecorder()
		s.Handler().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("workspace/summary: expected %d, got %d (%s)", http.StatusOK, rr.Code, rr.Body.String())
		}
		var body map[string]any
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("workspace/summary json: %v", err)
		}
		stats, _ := body["stats"].(map[string]any)
		if stats == nil {
			t.Fatalf("workspace/summary: missing stats")
		}
		if stats["ticketsTotal"] == nil {
			t.Fatalf("workspace/summary: missing ticketsTotal")
		}
	}

	// /workspace/tickets filter by status
	{
		req := httptest.NewRequest(http.MethodGet, "/api/v1/workspace/tickets?status=active&pageSize=200", nil)
		rr := httptest.NewRecorder()
		s.Handler().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("workspace/tickets: expected %d, got %d (%s)", http.StatusOK, rr.Code, rr.Body.String())
		}
		var body struct {
			Total   int `json:"total"`
			Results []struct {
				Ticket string `json:"ticket"`
			} `json:"results"`
		}
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("workspace/tickets json: %v", err)
		}
		if body.Total != 1 || len(body.Results) != 1 || body.Results[0].Ticket != "TST-123" {
			t.Fatalf("workspace/tickets status=active: unexpected results: %+v", body)
		}
	}

	// /workspace/tickets filter by owners + intent
	{
		req := httptest.NewRequest(http.MethodGet, "/api/v1/workspace/tickets?owners=alex&intent=short-term&pageSize=200", nil)
		rr := httptest.NewRecorder()
		s.Handler().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("workspace/tickets owners+intent: expected %d, got %d (%s)", http.StatusOK, rr.Code, rr.Body.String())
		}
		var body struct {
			Total   int `json:"total"`
			Results []struct {
				Ticket string `json:"ticket"`
			} `json:"results"`
		}
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("workspace/tickets owners+intent json: %v", err)
		}
		if body.Total != 1 || len(body.Results) != 1 || body.Results[0].Ticket != "TST-999" {
			t.Fatalf("workspace/tickets owners+intent: unexpected results: %+v", body)
		}
	}

	// /workspace/facets
	{
		req := httptest.NewRequest(http.MethodGet, "/api/v1/workspace/facets", nil)
		rr := httptest.NewRecorder()
		s.Handler().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("workspace/facets: expected %d, got %d (%s)", http.StatusOK, rr.Code, rr.Body.String())
		}
		var body struct {
			Owners []string `json:"owners"`
		}
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("workspace/facets json: %v", err)
		}
		if len(body.Owners) == 0 {
			t.Fatalf("workspace/facets: expected owners")
		}
	}

	// /workspace/recent
	{
		req := httptest.NewRequest(http.MethodGet, "/api/v1/workspace/recent?ticketsLimit=5&docsLimit=5", nil)
		rr := httptest.NewRecorder()
		s.Handler().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("workspace/recent: expected %d, got %d (%s)", http.StatusOK, rr.Code, rr.Body.String())
		}
		var body struct {
			Tickets []any `json:"tickets"`
			Docs    []any `json:"docs"`
		}
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("workspace/recent json: %v", err)
		}
		if len(body.Tickets) == 0 || len(body.Docs) == 0 {
			t.Fatalf("workspace/recent: expected non-empty tickets and docs")
		}
	}

	// /workspace/topics and /workspace/topics/get
	{
		req := httptest.NewRequest(http.MethodGet, "/api/v1/workspace/topics", nil)
		rr := httptest.NewRecorder()
		s.Handler().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("workspace/topics: expected %d, got %d (%s)", http.StatusOK, rr.Code, rr.Body.String())
		}
		var body struct {
			Total int `json:"total"`
		}
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("workspace/topics json: %v", err)
		}
		if body.Total == 0 {
			t.Fatalf("workspace/topics: expected topics")
		}

		req2 := httptest.NewRequest(http.MethodGet, "/api/v1/workspace/topics/get?topic=docmgr&docsLimit=5", nil)
		rr2 := httptest.NewRecorder()
		s.Handler().ServeHTTP(rr2, req2)
		if rr2.Code != http.StatusOK {
			t.Fatalf("workspace/topics/get: expected %d, got %d (%s)", http.StatusOK, rr2.Code, rr2.Body.String())
		}
	}
}
