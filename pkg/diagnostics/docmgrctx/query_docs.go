package docmgrctx

import (
	"fmt"

	"github.com/go-go-golems/docmgr/pkg/diagnostics/core"
)

const (
	StageWorkspaceQuery core.StageCode = "docmgr.workspace.query"

	SymptomQuerySkippedParse          core.SymptomCode = "query_skipped_due_to_parse"
	SymptomQueryNormalizationFallback core.SymptomCode = "query_normalization_fallback"
)

// WorkspaceQuerySkipContext captures a document skipped (from results) during QueryDocs due to parse issues.
type WorkspaceQuerySkipContext struct {
	File   string
	Reason string
}

func (c *WorkspaceQuerySkipContext) Stage() core.StageCode { return StageWorkspaceQuery }
func (c *WorkspaceQuerySkipContext) Summary() string {
	return fmt.Sprintf("QueryDocs skipped %s: %s", c.File, c.Reason)
}

// WorkspaceQueryNormalizationContext captures when reverse-lookup normalization falls back to a weaker key set.
type WorkspaceQueryNormalizationContext struct {
	Kind  string // "file" or "dir"
	Input string
	Note  string
}

func (c *WorkspaceQueryNormalizationContext) Stage() core.StageCode { return StageWorkspaceQuery }
func (c *WorkspaceQueryNormalizationContext) Summary() string {
	return fmt.Sprintf("QueryDocs normalization fallback (%s) for %q: %s", c.Kind, c.Input, c.Note)
}

func NewWorkspaceQuerySkippedParseTaxonomy(file, reason string, cause error) *core.Taxonomy {
	return &core.Taxonomy{
		Tool:     "docmgr",
		Stage:    StageWorkspaceQuery,
		Symptom:  SymptomQuerySkippedParse,
		Path:     file,
		Severity: core.SeverityWarning,
		Context: &WorkspaceQuerySkipContext{
			File:   file,
			Reason: reason,
		},
		Cause: cause,
	}
}

func NewWorkspaceQueryNormalizationFallbackTaxonomy(kind, input, note string) *core.Taxonomy {
	return &core.Taxonomy{
		Tool:     "docmgr",
		Stage:    StageWorkspaceQuery,
		Symptom:  SymptomQueryNormalizationFallback,
		Path:     input,
		Severity: core.SeverityWarning,
		Context: &WorkspaceQueryNormalizationContext{
			Kind:  kind,
			Input: input,
			Note:  note,
		},
	}
}
