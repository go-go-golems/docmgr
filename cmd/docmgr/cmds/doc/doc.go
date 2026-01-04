package doc

import (
	"github.com/spf13/cobra"
)

// Attach registers the document-focused command tree under the provided root command.
func Attach(root *cobra.Command) error {
	docCmd := &cobra.Command{
		Use:   "doc",
		Short: "Document workspace operations",
		Long: `Document workspace operations: create docs, list/search, and maintain relationships to code.

Examples:
  # Create a doc in a ticket
  docmgr doc add --ticket MEN-4242 --doc-type reference --title "Diary"

  # Search docs by content
  docmgr doc search --query "WebSocket"
`,
	}

	addCmd, err := newAddCommand()
	if err != nil {
		return err
	}
	listCmd, err := newListCommand()
	if err != nil {
		return err
	}
	searchCmd, err := newSearchCommand()
	if err != nil {
		return err
	}
	guidelinesCmd, err := newGuidelinesCommand()
	if err != nil {
		return err
	}
	relateCmd, err := newRelateCommand()
	if err != nil {
		return err
	}
	layoutFixCmd, err := newLayoutFixCommand()
	if err != nil {
		return err
	}
	renumberCmd, err := newRenumberCommand()
	if err != nil {
		return err
	}
	moveCmd, err := newMoveCommand()
	if err != nil {
		return err
	}

	docCmd.AddCommand(
		addCmd,
		listCmd,
		searchCmd,
		guidelinesCmd,
		relateCmd,
		layoutFixCmd,
		renumberCmd,
		moveCmd,
	)
	root.AddCommand(docCmd)
	return nil
}
