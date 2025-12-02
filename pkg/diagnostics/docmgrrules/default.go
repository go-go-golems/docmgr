package docmgrrules

import (
	"github.com/go-go-golems/docmgr/pkg/diagnostics/rules"
)

// DefaultRegistry returns a registry seeded with docmgr rules.
func DefaultRegistry() *rules.Registry {
	reg := rules.NewRegistry()
	reg.Register(&VocabularySuggestionRule{})
	reg.Register(&RelatedFileMissingRule{})
	reg.Register(&FrontmatterSyntaxRule{})
	reg.Register(&FrontmatterSchemaRule{})
	reg.Register(&ListingSkipRule{})
	reg.Register(&WorkspaceRule{})
	return reg
}
