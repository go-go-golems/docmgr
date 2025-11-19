package changelog

import "github.com/spf13/cobra"

// Attach registers changelog update command under docmgr changelog update.
func Attach(root *cobra.Command) error {
	changelogCmd := &cobra.Command{
		Use:   "changelog",
		Short: "Manage changelog entries",
	}

	updateCmd, err := newUpdateCommand()
	if err != nil {
		return err
	}
	changelogCmd.AddCommand(updateCmd)
	root.AddCommand(changelogCmd)
	return nil
}
