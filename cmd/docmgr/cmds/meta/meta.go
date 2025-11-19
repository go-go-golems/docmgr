package meta

import "github.com/spf13/cobra"

// Attach registers metadata commands (currently only update) under docmgr meta.
func Attach(root *cobra.Command) error {
	metaCmd := &cobra.Command{
		Use:   "meta",
		Short: "Manage document metadata",
	}

	updateCmd, err := newUpdateCommand()
	if err != nil {
		return err
	}
	metaCmd.AddCommand(updateCmd)
	root.AddCommand(metaCmd)
	return nil
}
