package docmgrctx

import (
	"fmt"

	"github.com/go-go-golems/docmgr/pkg/diagnostics/core"
)

const (
	StageFrontmatterParse  core.StageCode   = "docmgr.frontmatter.parse"
	SymptomYAMLSyntax      core.SymptomCode = "yaml_syntax"
	SymptomSchemaViolation core.SymptomCode = "schema_violation"
)

// FrontmatterParseContext carries YAML parse failure details.
type FrontmatterParseContext struct {
	File    string
	Line    int
	Column  int
	Snippet string
	Problem string
	Fixes   []string
}

func (c *FrontmatterParseContext) Stage() core.StageCode { return StageFrontmatterParse }
func (c *FrontmatterParseContext) Summary() string {
	loc := ""
	if c.Line > 0 {
		loc = fmt.Sprintf(":%d", c.Line)
		if c.Column > 0 {
			loc += fmt.Sprintf(":%d", c.Column)
		}
	}
	return fmt.Sprintf("%s%s %s", c.File, loc, c.Problem)
}

// NewFrontmatterParseTaxonomy builds a taxonomy for YAML/frontmatter syntax errors.
func NewFrontmatterParseTaxonomy(file string, line, col int, snippet, problem string, cause error) *core.Taxonomy {
	return &core.Taxonomy{
		Tool:     "docmgr",
		Stage:    StageFrontmatterParse,
		Symptom:  SymptomYAMLSyntax,
		Path:     file,
		Severity: core.SeverityError,
		Context: &FrontmatterParseContext{
			File:    file,
			Line:    line,
			Column:  col,
			Snippet: snippet,
			Problem: problem,
		},
		Cause: cause,
	}
}

// FrontmatterSchemaContext carries schema validation warnings/errors.
type FrontmatterSchemaContext struct {
	File   string
	Field  string
	Detail string
}

func (c *FrontmatterSchemaContext) Stage() core.StageCode { return StageFrontmatterParse }
func (c *FrontmatterSchemaContext) Summary() string {
	return fmt.Sprintf("%s: %s (%s)", c.File, c.Field, c.Detail)
}

// NewFrontmatterSchemaTaxonomy builds a taxonomy for schema validation issues.
func NewFrontmatterSchemaTaxonomy(file, field, detail string, severity core.Severity) *core.Taxonomy {
	return &core.Taxonomy{
		Tool:     "docmgr",
		Stage:    StageFrontmatterParse,
		Symptom:  SymptomSchemaViolation,
		Path:     field,
		Severity: severity,
		Context: &FrontmatterSchemaContext{
			File:   file,
			Field:  field,
			Detail: detail,
		},
	}
}
