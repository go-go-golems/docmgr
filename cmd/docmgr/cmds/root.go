package cmds

import (
	"github.com/go-go-golems/docmgr/cmd/docmgr/cmds/changelog"
	"github.com/go-go-golems/docmgr/cmd/docmgr/cmds/configcmd"
	"github.com/go-go-golems/docmgr/cmd/docmgr/cmds/doc"
	"github.com/go-go-golems/docmgr/cmd/docmgr/cmds/importcmd"
	"github.com/go-go-golems/docmgr/cmd/docmgr/cmds/list"
	"github.com/go-go-golems/docmgr/cmd/docmgr/cmds/meta"
	"github.com/go-go-golems/docmgr/cmd/docmgr/cmds/tasks"
	"github.com/go-go-golems/docmgr/cmd/docmgr/cmds/template"
	"github.com/go-go-golems/docmgr/cmd/docmgr/cmds/ticket"
	"github.com/go-go-golems/docmgr/cmd/docmgr/cmds/vocab"
	"github.com/go-go-golems/docmgr/cmd/docmgr/cmds/workspace"
	"github.com/go-go-golems/docmgr/pkg/completion"
	"github.com/go-go-golems/glazed/pkg/help"
	help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
	"github.com/spf13/cobra"
)

// NewRootCommand builds the Cobra root command and wires the hierarchical command tree.
func NewRootCommand(helpSystem *help.HelpSystem) (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:   "docmgr",
		Short: "Document Manager for LLM Workflows",
		Long: `docmgr is a structured document manager for managing documentation
in LLM-assisted workflows. It provides commands to create, organize,
and validate documentation workspaces with metadata and external source support.

Helpful docs (built-in):

  - Quick usage:          docmgr help how-to-use
  - Initial setup guide:  docmgr help how-to-setup
  - List all embedded docs: docmgr help --all`,
	}

	help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)

	// Enable carapace dynamic completion (adds hidden `_carapace` and bridges cobra)
	completion.Attach(rootCmd)

	if err := workspace.Attach(rootCmd); err != nil {
		return nil, err
	}
	if err := ticket.Attach(rootCmd); err != nil {
		return nil, err
	}
	if err := doc.Attach(rootCmd); err != nil {
		return nil, err
	}
	// Add alias: docmgr search -> docmgr doc search
	searchCmd, err := doc.NewSearchCommand()
	if err != nil {
		return nil, err
	}
	searchCmd.Use = "search"
	rootCmd.AddCommand(searchCmd)
	if err := tasks.Attach(rootCmd); err != nil {
		return nil, err
	}
	if err := vocab.Attach(rootCmd); err != nil {
		return nil, err
	}
	if err := meta.Attach(rootCmd); err != nil {
		return nil, err
	}
	if err := importcmd.Attach(rootCmd); err != nil {
		return nil, err
	}
	if err := configcmd.Attach(rootCmd); err != nil {
		return nil, err
	}
	if err := changelog.Attach(rootCmd); err != nil {
		return nil, err
	}
	if err := list.Attach(rootCmd); err != nil {
		return nil, err
	}
	if err := template.Attach(rootCmd); err != nil {
		return nil, err
	}

	return rootCmd, nil
}
