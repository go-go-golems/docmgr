package documents

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-go-golems/docmgr/pkg/diagnostics/core"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/docmgrctx"
	"github.com/go-go-golems/docmgr/pkg/models"
)

func TestReadDocumentWithFrontmatter_Valid(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "doc.md")
	content := `---
Title: Hello
Ticket: TEST-1
DocType: design-doc
Summary: simple
---
Body line
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	doc, body, err := ReadDocumentWithFrontmatter(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Title != "Hello" || doc.Ticket != "TEST-1" || doc.DocType != "design-doc" {
		t.Fatalf("parsed doc mismatch: %+v", doc)
	}
	if strings.TrimSpace(body) != "Body line" {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestReadDocumentWithFrontmatter_InvalidReportsLine(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "bad.md")
	content := `---
Title: Broken
Ticket: TEST-2
DocType: design-doc
Topics: [chat
---
Body
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	_, _, err := ReadDocumentWithFrontmatter(path)
	if err == nil {
		t.Fatalf("expected error")
	}
	tax, ok := core.AsTaxonomy(err)
	if !ok || tax == nil {
		t.Fatalf("expected taxonomy in error chain, got %v", err)
	}
	ctx, ok := tax.Context.(*docmgrctx.FrontmatterParseContext)
	if !ok {
		t.Fatalf("unexpected context type: %T", tax.Context)
	}
	if ctx.Line == 0 {
		t.Fatalf("expected line number in context")
	}
	if ctx.Problem == "" {
		t.Fatalf("expected problem message")
	}
	if ctx.Snippet == "" || !strings.Contains(ctx.Snippet, "Topics: [chat") {
		t.Fatalf("expected snippet with topics line, got: %q", ctx.Snippet)
	}
}

func TestWriteDocumentWithFrontmatter_QuotesUnsafeScalars(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "out.md")
	doc := &models.Document{
		Title:       "Hello",
		Ticket:      "TEST-3",
		DocType:     "design-doc",
		Summary:     "colon: here",
		LastUpdated: time.Now(),
	}
	if err := WriteDocumentWithFrontmatter(path, doc, "body", true); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	text := string(b)
	if !strings.Contains(text, "Summary: 'colon: here'") {
		t.Fatalf("expected quoted summary, got:\n%s", text)
	}
	// Ensure it remains parseable.
	if _, _, err := ReadDocumentWithFrontmatter(path); err != nil {
		t.Fatalf("round-trip parse failed: %v", err)
	}
}
