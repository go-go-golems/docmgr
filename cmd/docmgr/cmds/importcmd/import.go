package importcmd

import "github.com/spf13/cobra"

// Attach registers the import command tree (currently only file import).
func Attach(root *cobra.Command) error {
	importCmd := &cobra.Command{
		Use:   "import",
		Short: "Import external sources",
		Long: `Import external source artifacts into a ticket workspace (under sources/).

Examples:
  # Import a local file into sources/
  docmgr import file --ticket MEN-4242 --file /path/to/spec.pdf --name "API Spec"

  # Import and emit JSON (for scripts)
  docmgr import file --ticket MEN-4242 --file /path/to/spec.pdf --with-glaze-output --output json
`,
	}

	fileCmd, err := newFileCommand()
	if err != nil {
		return err
	}
	importCmd.AddCommand(fileCmd)
	root.AddCommand(importCmd)
	return nil
}
