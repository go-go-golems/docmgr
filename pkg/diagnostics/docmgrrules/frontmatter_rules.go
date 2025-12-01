package docmgrrules

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-go-golems/docmgr/pkg/diagnostics/core"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/docmgrctx"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/rules"
)

// FrontmatterSyntaxRule renders YAML/frontmatter syntax guidance.
type FrontmatterSyntaxRule struct{}

func (r *FrontmatterSyntaxRule) Match(t *core.Taxonomy) (bool, int) {
	return t != nil && t.Stage == docmgrctx.StageFrontmatterParse && t.Symptom == docmgrctx.SymptomYAMLSyntax, 100
}

func (r *FrontmatterSyntaxRule) Render(ctx context.Context, t *core.Taxonomy) (*rules.RuleResult, error) {
	payload, ok := t.Context.(*docmgrctx.FrontmatterParseContext)
	if !ok || payload == nil {
		return nil, fmt.Errorf("frontmatter syntax rule: unexpected context type")
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("File: %s\n", payload.File))
	if payload.Line > 0 {
		b.WriteString(fmt.Sprintf("Line: %d Col: %d\n", payload.Line, payload.Column))
	}
	if payload.Problem != "" {
		b.WriteString(fmt.Sprintf("Problem: %s\n", payload.Problem))
	}
	if payload.Snippet != "" {
		b.WriteString("\nSnippet:\n")
		b.WriteString(payload.Snippet + "\n")
	}
	if len(payload.Fixes) > 0 {
		b.WriteString("\nSuggested fixes:\n")
		for i, fix := range payload.Fixes {
			b.WriteString(fmt.Sprintf("  %d. %s\n", i+1, fix))
		}
	}
	actions := []rules.Action{
		{Label: "Validate frontmatter", Command: "docmgr", Args: []string{"validate", "frontmatter", "--doc", payload.File}},
	}
	return &rules.RuleResult{
		Headline: "YAML/frontmatter syntax error",
		Body:     b.String(),
		Severity: core.SeverityError,
		Actions:  actions,
	}, nil
}

// FrontmatterSchemaRule renders schema validation guidance.
type FrontmatterSchemaRule struct{}

func (r *FrontmatterSchemaRule) Match(t *core.Taxonomy) (bool, int) {
	return t != nil && t.Stage == docmgrctx.StageFrontmatterParse && t.Symptom == docmgrctx.SymptomSchemaViolation, 80
}

func (r *FrontmatterSchemaRule) Render(ctx context.Context, t *core.Taxonomy) (*rules.RuleResult, error) {
	payload, ok := t.Context.(*docmgrctx.FrontmatterSchemaContext)
	if !ok || payload == nil {
		return nil, fmt.Errorf("frontmatter schema rule: unexpected context type")
	}
	body := fmt.Sprintf("File: %s\nField: %s\nIssue: %s\n", payload.File, payload.Field, payload.Detail)
	actions := []rules.Action{
		{
			Label:   "Update field",
			Command: "docmgr",
			Args:    []string{"meta", "update", "--doc", payload.File, "--field", payload.Field, "--value", "<value>"},
		},
	}
	return &rules.RuleResult{
		Headline: fmt.Sprintf("Frontmatter validation: %s", payload.Field),
		Body:     body,
		Severity: t.Severity,
		Actions:  actions,
	}, nil
}
