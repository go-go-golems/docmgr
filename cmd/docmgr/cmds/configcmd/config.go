package configcmd

import "github.com/spf13/cobra"

// Attach registers configuration commands such as docmgr config show.
func Attach(root *cobra.Command) error {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Inspect docmgr configuration",
	}

	showCmd, err := newShowCommand()
	if err != nil {
		return err
	}
	configCmd.AddCommand(showCmd)
	root.AddCommand(configCmd)
	return nil
}
