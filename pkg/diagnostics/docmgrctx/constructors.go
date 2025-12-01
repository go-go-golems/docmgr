package docmgrctx

import "github.com/go-go-golems/docmgr/pkg/diagnostics/core"

// Helper constructors for docmgr taxonomies per domain.

func NewVocabularyUnknown(file, field, value string, known []string) *core.Taxonomy {
	return NewVocabularyUnknownTaxonomy(file, field, value, known)
}

func NewRelatedFileMissing(docPath, filePath, note string) *core.Taxonomy {
	return NewRelatedFileMissingTaxonomy(docPath, filePath, note)
}

func NewFrontmatterParse(file string, line, col int, snippet, problem string, cause error) *core.Taxonomy {
	return NewFrontmatterParseTaxonomy(file, line, col, snippet, problem, cause)
}

func NewFrontmatterSchema(file, field, detail string, severity core.Severity) *core.Taxonomy {
	return NewFrontmatterSchemaTaxonomy(file, field, detail, severity)
}

func NewTemplateParse(file, problem string, cause error) *core.Taxonomy {
	return NewTemplateParseTaxonomy(file, problem, cause)
}

func NewListingSkip(command, file, reason string, cause error) *core.Taxonomy {
	return NewListingSkipTaxonomy(command, file, reason, cause)
}

func NewWorkspaceMissingIndex(path string) *core.Taxonomy {
	return NewMissingIndexTaxonomy(path)
}

func NewWorkspaceStale(file string, lastUpdatedTime interface{}, threshold int) *core.Taxonomy {
	// lastUpdatedTime kept as interface{} to allow easy calling from various contexts.
	switch v := lastUpdatedTime.(type) {
	case core.Taxonomy:
		return NewStaleDocTaxonomy(file, v.Context.(*StalenessContext).LastUpdated, threshold)
	case *core.Taxonomy:
		return NewStaleDocTaxonomy(file, v.Context.(*StalenessContext).LastUpdated, threshold)
	case StalenessContext:
		return NewStaleDocTaxonomy(file, v.LastUpdated, threshold)
	case *StalenessContext:
		return NewStaleDocTaxonomy(file, v.LastUpdated, threshold)
	case interface{ Time() int64 }:
		// fallback
	}
	return NewStaleDocTaxonomy(file, StalenessContext{}.LastUpdated, threshold)
}
