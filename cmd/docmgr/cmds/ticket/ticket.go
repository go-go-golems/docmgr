package ticket

import "github.com/spf13/cobra"

// Attach registers ticket-related commands (create/list/rename) under the root.
func Attach(root *cobra.Command) error {
	ticketCmd := &cobra.Command{
		Use:   "ticket",
		Short: "Ticket workspace management",
	}

	createCmd, err := newCreateCommand()
	if err != nil {
		return err
	}
	listCmd, err := newListCommand()
	if err != nil {
		return err
	}
	renameCmd, err := newRenameCommand()
	if err != nil {
		return err
	}
	closeCmd, err := newCloseCommand()
	if err != nil {
		return err
	}
	moveCmd, err := newMoveCommand()
	if err != nil {
		return err
	}

	ticketCmd.AddCommand(createCmd, listCmd, renameCmd, closeCmd, moveCmd)
	root.AddCommand(ticketCmd)
	return nil
}
