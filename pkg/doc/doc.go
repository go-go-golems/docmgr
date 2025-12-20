// Package doc contains embedded documentation files for docmgr's help system.
//
// Documentation files are embedded at build time and made available through
// the Glazed help system. Users can access these docs via `docmgr help` commands.
//
// The embedded files include:
//   - docmgr-how-to-use.md: Quick start and usage guide
//   - docmgr-how-to-setup.md: Initial setup instructions
//   - how-to-work-on-any-ticket.md: Checklist for taking over any ticket
//   - docmgr-cli-guide.md: Complete CLI reference
//   - templates-and-guidelines.md: Document templates and guidelines
//   - docmgr-ci-automation.md: CI/CD integration guide
//   - using-skills.md: LLM bootstrap prompt pack (docmgr + skills)
//   - how-to-write-skills.md: Guide for creating and maintaining skill documents
package doc

import (
	"embed"
	"github.com/go-go-golems/glazed/pkg/help"
)

//go:embed *
var docFS embed.FS

func AddDocToHelpSystem(helpSystem *help.HelpSystem) error {
	return helpSystem.LoadSectionsFromFS(docFS, ".")
}
