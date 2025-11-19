package list

import "github.com/spf13/cobra"

// Attach registers the compatibility list command (docmgr list docs|tickets).
func Attach(root *cobra.Command) error {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List tickets or documents",
	}

	docsCmd, err := newDocsCommand()
	if err != nil {
		return err
	}
	ticketsCmd, err := newTicketsCommand()
	if err != nil {
		return err
	}

	listCmd.AddCommand(ticketsCmd, docsCmd)
	root.AddCommand(listCmd)
	return nil
}
