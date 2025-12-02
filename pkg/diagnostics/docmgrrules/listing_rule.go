package docmgrrules

import (
	"context"
	"fmt"

	"github.com/go-go-golems/docmgr/pkg/diagnostics/core"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/docmgrctx"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/rules"
)

// ListingSkipRule surfaces documents skipped due to parse issues during listing/search.
type ListingSkipRule struct{}

func (r *ListingSkipRule) Match(t *core.Taxonomy) (bool, int) {
	return t != nil && t.Stage == docmgrctx.StageListing && t.Symptom == docmgrctx.SymptomSkippedParse, 60
}

func (r *ListingSkipRule) Render(ctx context.Context, t *core.Taxonomy) (*rules.RuleResult, error) {
	payload, ok := t.Context.(*docmgrctx.ListingSkipContext)
	if !ok || payload == nil {
		return nil, fmt.Errorf("listing skip rule: unexpected context type")
	}
	body := fmt.Sprintf("Command: %s\nFile: %s\nReason: %s\n", payload.Command, payload.File, payload.Reason)
	actions := []rules.Action{
		{Label: "Validate frontmatter", Command: "docmgr", Args: []string{"validate", "frontmatter", "--doc", payload.File}},
	}
	return &rules.RuleResult{
		Headline: "Document skipped during listing",
		Body:     body,
		Severity: t.Severity,
		Actions:  actions,
	}, nil
}
