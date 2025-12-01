package docmgrctx

import (
	"fmt"

	"github.com/go-go-golems/docmgr/pkg/diagnostics/core"
)

const (
	StageTemplateParse   core.StageCode   = "docmgr.template.parse"
	SymptomTemplateError core.SymptomCode = "template_parse_error"
)

// TemplateParseContext captures template parse failures.
type TemplateParseContext struct {
	File    string
	Problem string
}

func (c *TemplateParseContext) Stage() core.StageCode { return StageTemplateParse }
func (c *TemplateParseContext) Summary() string {
	return fmt.Sprintf("%s: %s", c.File, c.Problem)
}

// NewTemplateParseTaxonomy builds a taxonomy for template parse errors.
func NewTemplateParseTaxonomy(file, problem string, cause error) *core.Taxonomy {
	return &core.Taxonomy{
		Tool:     "docmgr",
		Stage:    StageTemplateParse,
		Symptom:  SymptomTemplateError,
		Path:     file,
		Severity: core.SeverityError,
		Context: &TemplateParseContext{
			File:    file,
			Problem: problem,
		},
		Cause: cause,
	}
}
