package docmgrrules

import (
	"context"
	"testing"

	"github.com/go-go-golems/docmgr/pkg/diagnostics/core"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/docmgrctx"
)

func TestRelatedFileMissingRule_MatchAndRender(t *testing.T) {
	rule := &RelatedFileMissingRule{}
	tax := &core.Taxonomy{
		Tool:     "docmgr",
		Stage:    docmgrctx.StageDocLink,
		Symptom:  docmgrctx.SymptomMissingFile,
		Path:     "a/b.md",
		Severity: core.SeverityWarning,
		Context: &docmgrctx.RelatedFileContext{
			DocPath:  "docs/index.md",
			FilePath: "missing.go",
			Note:     "example note",
			Exists:   false,
		},
	}

	ok, score := rule.Match(tax)
	if !ok {
		t.Fatalf("expected match")
	}
	if score <= 0 {
		t.Fatalf("expected positive score, got %d", score)
	}

	res, err := rule.Render(context.Background(), tax)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if res == nil || res.Headline == "" {
		t.Fatalf("expected headline, got %+v", res)
	}
	if len(res.Actions) < 1 {
		t.Fatalf("expected at least one action, got %+v", res.Actions)
	}
}

func TestRelatedFileMissingRule_NoMatch(t *testing.T) {
	rule := &RelatedFileMissingRule{}
	tax := &core.Taxonomy{
		Stage:   docmgrctx.StageVocabulary,
		Symptom: docmgrctx.SymptomUnknownValue,
	}
	if ok, _ := rule.Match(tax); ok {
		t.Fatalf("expected no match")
	}
}
