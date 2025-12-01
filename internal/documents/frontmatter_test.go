package documents

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-go-golems/docmgr/pkg/diagnostics/core"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/docmgrctx"
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
Summary: unquoted: colon
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
	if ctx.Snippet == "" || !strings.Contains(ctx.Snippet, "Summary: unquoted: colon") {
		t.Fatalf("expected snippet with summary line, got: %q", ctx.Snippet)
	}
}
