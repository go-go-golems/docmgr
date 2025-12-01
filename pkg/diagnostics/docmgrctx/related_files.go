package docmgrctx

import (
	"fmt"

	"github.com/go-go-golems/docmgr/pkg/diagnostics/core"
)

const (
	StageDocLink           core.StageCode   = "docmgr.related_files"
	SymptomMissingFile     core.SymptomCode = "missing_file"
	SymptomMissingNote     core.SymptomCode = "missing_note"
	SymptomInvalidFilePath core.SymptomCode = "invalid_file_path"
)

// RelatedFileContext captures issues with RelatedFiles entries.
type RelatedFileContext struct {
	DocPath  string // markdown document containing RelatedFiles
	FilePath string // related file path entry
	Note     string // optional note attached to the entry
	Exists   bool   // whether file exists on disk
}

func (c *RelatedFileContext) Stage() core.StageCode { return StageDocLink }
func (c *RelatedFileContext) Summary() string {
	status := "missing"
	if c.Exists {
		status = "present"
	}
	return fmt.Sprintf("%s: related file %q (%s)", c.DocPath, c.FilePath, status)
}

// NewRelatedFileMissingTaxonomy builds a taxonomy for missing related files.
func NewRelatedFileMissingTaxonomy(docPath, filePath, note string) *core.Taxonomy {
	return &core.Taxonomy{
		Tool:     "docmgr",
		Stage:    StageDocLink,
		Symptom:  SymptomMissingFile,
		Path:     filePath,
		Severity: core.SeverityWarning,
		Context: &RelatedFileContext{
			DocPath:  docPath,
			FilePath: filePath,
			Note:     note,
			Exists:   false,
		},
	}
}
