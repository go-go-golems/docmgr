package docmgrrules

import (
	"context"
	"strings"
	"testing"

	"github.com/go-go-golems/docmgr/pkg/diagnostics/core"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/docmgrctx"
)

func TestFrontmatterSchemaRule_Render(t *testing.T) {
	rule := &FrontmatterSchemaRule{}
	tax := docmgrctx.NewFrontmatterSchema("doc.md", "Summary", "missing Summary", core.SeverityWarning)

	ok, score := rule.Match(tax)
	if !ok || score <= 0 {
		t.Fatalf("expected match with positive score, got ok=%v score=%d", ok, score)
	}

	res, err := rule.Render(context.Background(), tax)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if res == nil {
		t.Fatalf("nil result")
	}
	if res.Severity != core.SeverityWarning {
		t.Fatalf("expected severity %s, got %s", core.SeverityWarning, res.Severity)
	}
	if !strings.Contains(res.Headline, "Summary") {
		t.Fatalf("headline missing field: %s", res.Headline)
	}
	if !strings.Contains(res.Body, "doc.md") {
		t.Fatalf("body missing file: %s", res.Body)
	}
	if len(res.Actions) == 0 {
		t.Fatalf("expected suggested actions")
	}
}
