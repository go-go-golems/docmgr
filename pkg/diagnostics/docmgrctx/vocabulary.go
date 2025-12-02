package docmgrctx

import (
	"fmt"

	"github.com/go-go-golems/docmgr/pkg/diagnostics/core"
)

const (
	StageVocabulary          core.StageCode   = "docmgr.vocabulary"
	SymptomUnknownValue      core.SymptomCode = "unknown_value"
	SymptomInvalidVocabulary core.SymptomCode = "invalid_vocabulary"
)

// VocabularyContext captures vocabulary validation details.
type VocabularyContext struct {
	File  string   // markdown file path
	Field string   // frontmatter field name
	Value string   // offending value
	Known []string // known values in vocabulary
}

func (c *VocabularyContext) Stage() core.StageCode { return StageVocabulary }
func (c *VocabularyContext) Summary() string {
	return fmt.Sprintf("%s: field %s unknown value %q", c.File, c.Field, c.Value)
}

// NewVocabularyUnknownTaxonomy builds a taxonomy for unknown vocabulary values.
func NewVocabularyUnknownTaxonomy(file, field, value string, known []string) *core.Taxonomy {
	return &core.Taxonomy{
		Tool:     "docmgr",
		Stage:    StageVocabulary,
		Symptom:  SymptomUnknownValue,
		Path:     field,
		Severity: core.SeverityWarning,
		Context: &VocabularyContext{
			File:  file,
			Field: field,
			Value: value,
			Known: known,
		},
	}
}
