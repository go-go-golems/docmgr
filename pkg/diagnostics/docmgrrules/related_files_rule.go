package docmgrrules

import (
	"context"
	"fmt"

	"github.com/go-go-golems/docmgr/pkg/diagnostics/core"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/docmgrctx"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/rules"
)

// RelatedFileMissingRule suggests actions for missing related file entries.
type RelatedFileMissingRule struct{}

func (r *RelatedFileMissingRule) Match(t *core.Taxonomy) (bool, int) {
	if t == nil {
		return false, 0
	}
	return t.Stage == docmgrctx.StageDocLink && t.Symptom == docmgrctx.SymptomMissingFile, 70
}

func (r *RelatedFileMissingRule) Render(ctx context.Context, t *core.Taxonomy) (*rules.RuleResult, error) {
	payload, ok := t.Context.(*docmgrctx.RelatedFileContext)
	if !ok || payload == nil {
		return nil, fmt.Errorf("related-file rule: unexpected context type")
	}

	body := fmt.Sprintf("Doc: %s\nRelated file: %s\nNote: %s\nStatus: missing on disk\n", payload.DocPath, payload.FilePath, payload.Note)
	actions := []rules.Action{
		{
			Label:   "Remove invalid entry",
			Command: "docmgr",
			Args:    []string{"doc", "relate", "--doc", payload.DocPath, "--remove-files", payload.FilePath},
		},
		{
			Label:   "Fix path",
			Command: "docmgr",
			Args:    []string{"doc", "relate", "--doc", payload.DocPath, "--file-note", fmt.Sprintf("%s:<reason>", payload.FilePath)},
		},
	}

	return &rules.RuleResult{
		Headline: "Missing related file entry",
		Body:     body,
		Severity: core.SeverityWarning,
		Actions:  actions,
	}, nil
}
