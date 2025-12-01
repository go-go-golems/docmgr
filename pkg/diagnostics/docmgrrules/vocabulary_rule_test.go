package docmgrrules

import (
	"context"
	"testing"

	"github.com/go-go-golems/docmgr/pkg/diagnostics/core"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/docmgrctx"
)

func TestVocabularySuggestionRule_MatchAndRender(t *testing.T) {
	rule := &VocabularySuggestionRule{}
	tax := &core.Taxonomy{
		Tool:     "docmgr",
		Stage:    docmgrctx.StageVocabulary,
		Symptom:  docmgrctx.SymptomUnknownValue,
		Path:     "Topics",
		Severity: core.SeverityWarning,
		Context: &docmgrctx.VocabularyContext{
			File:  "doc.md",
			Field: "Topics",
			Value: "custom",
			Known: []string{"chat", "backend"},
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
	if res == nil {
		t.Fatalf("nil result")
	}
	if res.Headline == "" || len(res.Actions) == 0 {
		t.Fatalf("expected headline and actions, got %+v", res)
	}
}

func TestVocabularySuggestionRule_NoMatch(t *testing.T) {
	rule := &VocabularySuggestionRule{}
	tax := &core.Taxonomy{
		Stage:   docmgrctx.StageDocLink,
		Symptom: docmgrctx.SymptomMissingFile,
	}
	if ok, _ := rule.Match(tax); ok {
		t.Fatalf("expected no match")
	}
}
