package docmgrctx

import (
	"fmt"

	"github.com/go-go-golems/docmgr/pkg/diagnostics/core"
)

const (
	StageListing        core.StageCode   = "docmgr.listing"
	SymptomSkippedParse core.SymptomCode = "skipped_due_to_parse"
)

// ListingSkipContext captures a document skipped during listing/search due to parse issues.
type ListingSkipContext struct {
	Command string
	File    string
	Reason  string
}

func (c *ListingSkipContext) Stage() core.StageCode { return StageListing }
func (c *ListingSkipContext) Summary() string {
	return fmt.Sprintf("%s skipped %s: %s", c.Command, c.File, c.Reason)
}

// NewListingSkipTaxonomy builds a taxonomy for skipped entries.
func NewListingSkipTaxonomy(command, file, reason string, cause error) *core.Taxonomy {
	return &core.Taxonomy{
		Tool:     "docmgr",
		Stage:    StageListing,
		Symptom:  SymptomSkippedParse,
		Path:     file,
		Severity: core.SeverityWarning,
		Context: &ListingSkipContext{
			Command: command,
			File:    file,
			Reason:  reason,
		},
		Cause: cause,
	}
}
