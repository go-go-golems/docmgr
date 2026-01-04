package meta

import "github.com/spf13/cobra"

// Attach registers metadata commands (currently only update) under docmgr meta.
func Attach(root *cobra.Command) error {
	metaCmd := &cobra.Command{
		Use:   "meta",
		Short: "Manage document metadata",
		Long: `Update YAML frontmatter fields in documents and ticket index files.

Examples:
  # Update the ticket index.md status
  docmgr meta update --ticket MEN-4242 --field Status --value active

  # Update a specific doc by path
  docmgr meta update --doc ttmp/YYYY/MM/DD/MEN-4242--.../reference/01-diary.md --field Summary --value "Short summary"
`,
	}

	updateCmd, err := newUpdateCommand()
	if err != nil {
		return err
	}
	metaCmd.AddCommand(updateCmd)
	root.AddCommand(metaCmd)
	return nil
}
