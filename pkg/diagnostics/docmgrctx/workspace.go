package docmgrctx

import (
	"fmt"
	"time"

	"github.com/go-go-golems/docmgr/pkg/diagnostics/core"
)

const (
	StageWorkspace      core.StageCode   = "docmgr.workspace"
	SymptomMissingIndex core.SymptomCode = "missing_index"
	SymptomStale        core.SymptomCode = "stale_document"
)

// WorkspaceContext captures workspace structural issues.
type WorkspaceContext struct {
	Path string
	Note string
}

func (c *WorkspaceContext) Stage() core.StageCode { return StageWorkspace }
func (c *WorkspaceContext) Summary() string {
	return fmt.Sprintf("%s: %s", c.Path, c.Note)
}

// StalenessContext captures staleness warnings.
type StalenessContext struct {
	File          string
	LastUpdated   time.Time
	ThresholdDays int
}

func (c *StalenessContext) Stage() core.StageCode { return StageWorkspace }
func (c *StalenessContext) Summary() string {
	return fmt.Sprintf("%s stale (%d days threshold)", c.File, c.ThresholdDays)
}

// Constructors
func NewMissingIndexTaxonomy(path string) *core.Taxonomy {
	return &core.Taxonomy{
		Tool:     "docmgr",
		Stage:    StageWorkspace,
		Symptom:  SymptomMissingIndex,
		Path:     path,
		Severity: core.SeverityError,
		Context: &WorkspaceContext{
			Path: path,
			Note: "index.md not found",
		},
	}
}

func NewStaleDocTaxonomy(file string, lastUpdated time.Time, threshold int) *core.Taxonomy {
	return &core.Taxonomy{
		Tool:     "docmgr",
		Stage:    StageWorkspace,
		Symptom:  SymptomStale,
		Path:     file,
		Severity: core.SeverityWarning,
		Context: &StalenessContext{
			File:          file,
			LastUpdated:   lastUpdated,
			ThresholdDays: threshold,
		},
	}
}
