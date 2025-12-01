package docmgrrules

import (
	"context"
	"fmt"
	"time"

	"github.com/go-go-golems/docmgr/pkg/diagnostics/core"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/docmgrctx"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/rules"
)

// WorkspaceRule renders workspace-related guidance (missing index, staleness).
type WorkspaceRule struct{}

func (r *WorkspaceRule) Match(t *core.Taxonomy) (bool, int) {
	if t == nil || t.Stage != docmgrctx.StageWorkspace {
		return false, 0
	}
	switch t.Symptom {
	case docmgrctx.SymptomMissingIndex:
		return true, 90
	case docmgrctx.SymptomStale:
		return true, 50
	default:
		return false, 0
	}
}

func (r *WorkspaceRule) Render(ctx context.Context, t *core.Taxonomy) (*rules.RuleResult, error) {
	switch t.Symptom {
	case docmgrctx.SymptomMissingIndex:
		payload, ok := t.Context.(*docmgrctx.WorkspaceContext)
		if !ok || payload == nil {
			return nil, fmt.Errorf("workspace rule: expected WorkspaceContext")
		}
		body := fmt.Sprintf("Path: %s\nIssue: %s\n", payload.Path, payload.Note)
		return &rules.RuleResult{
			Headline: "Missing index.md",
			Body:     body,
			Severity: t.Severity,
			Actions: []rules.Action{
				{Label: "Recreate ticket", Command: "docmgr", Args: []string{"ticket", "create-ticket", "--ticket", "<ID>"}},
			},
		}, nil

	case docmgrctx.SymptomStale:
		payload, ok := t.Context.(*docmgrctx.StalenessContext)
		if !ok || payload == nil {
			return nil, fmt.Errorf("workspace rule: expected StalenessContext")
		}
		body := fmt.Sprintf("File: %s\nLastUpdated: %s\nThreshold: %d days\n", payload.File, payload.LastUpdated.Format(time.RFC3339), payload.ThresholdDays)
		return &rules.RuleResult{
			Headline: "Stale document",
			Body:     body,
			Severity: t.Severity,
			Actions: []rules.Action{
				{Label: "Refresh doc", Command: "docmgr", Args: []string{"meta", "update", "--doc", payload.File, "--field", "LastUpdated", "--value", time.Now().Format(time.RFC3339)}},
			},
		}, nil
	}
	return nil, fmt.Errorf("workspace rule: unsupported symptom %s", t.Symptom)
}
