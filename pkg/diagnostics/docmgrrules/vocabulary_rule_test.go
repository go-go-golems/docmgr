package docmgrrules

import (
	"context"
	"strings"
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

func TestVocabularySuggestionRule_RemediationUsesValidCategories(t *testing.T) {
	rule := &VocabularySuggestionRule{}
	fieldToCategory := map[string]string{
		"Topics":  "topics",
		"DocType": "docTypes",
		"Intent":  "intent",
		"Status":  "status",
	}

	for field, wantCategory := range fieldToCategory {
		tax := &core.Taxonomy{
			Tool:     "docmgr",
			Stage:    docmgrctx.StageVocabulary,
			Symptom:  docmgrctx.SymptomUnknownValue,
			Path:     field,
			Severity: core.SeverityWarning,
			Context: &docmgrctx.VocabularyContext{
				File:  "doc.md",
				Field: field,
				Value: "custom",
			},
		}

		res, err := rule.Render(context.Background(), tax)
		if err != nil {
			t.Fatalf("render failed for %s: %v", field, err)
		}
		args := res.Actions[0].Args
		joined := strings.Join(args, " ")
		if !strings.Contains(joined, "--category "+wantCategory+" ") {
			t.Fatalf("expected category %q for field %s, got args %v", wantCategory, field, args)
		}
		if !strings.Contains(joined, "--description") {
			t.Fatalf("expected --description in remediation for field %s, got args %v", field, args)
		}
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
