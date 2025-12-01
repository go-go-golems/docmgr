package docmgrrules

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-go-golems/docmgr/pkg/diagnostics/core"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/docmgrctx"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/rules"
)

// VocabularySuggestionRule renders guidance for unknown vocabulary entries.
type VocabularySuggestionRule struct{}

func (r *VocabularySuggestionRule) Match(t *core.Taxonomy) (bool, int) {
	if t == nil {
		return false, 0
	}
	return t.Stage == docmgrctx.StageVocabulary && t.Symptom == docmgrctx.SymptomUnknownValue, 80
}

func (r *VocabularySuggestionRule) Render(ctx context.Context, t *core.Taxonomy) (*rules.RuleResult, error) {
	payload, ok := t.Context.(*docmgrctx.VocabularyContext)
	if !ok || payload == nil {
		return nil, fmt.Errorf("vocabulary rule: unexpected context type")
	}

	body := fmt.Sprintf("File: %s\nField: %s\nValue: %q\n", payload.File, payload.Field, payload.Value)
	if len(payload.Known) > 0 {
		body += fmt.Sprintf("Known values: %s\n", strings.Join(payload.Known, ", "))
	}

	actions := []rules.Action{
		{
			Label:   "Add to vocabulary",
			Command: "docmgr",
			Args:    []string{"vocab", "add", "--category", strings.ToLower(payload.Field), "--slug", payload.Value},
		},
		{
			Label:   "List vocabulary",
			Command: "docmgr",
			Args:    []string{"vocab", "list"},
		},
	}

	return &rules.RuleResult{
		Headline: fmt.Sprintf("Unknown vocabulary value for %s", payload.Field),
		Body:     body,
		Severity: t.Severity,
		Actions:  actions,
	}, nil
}
