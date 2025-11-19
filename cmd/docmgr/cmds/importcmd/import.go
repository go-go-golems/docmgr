package importcmd

import "github.com/spf13/cobra"

// Attach registers the import command tree (currently only file import).
func Attach(root *cobra.Command) error {
	importCmd := &cobra.Command{
		Use:   "import",
		Short: "Import external sources",
	}

	fileCmd, err := newFileCommand()
	if err != nil {
		return err
	}
	importCmd.AddCommand(fileCmd)
	root.AddCommand(importCmd)
	return nil
}
