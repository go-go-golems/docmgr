package ticket

import "github.com/spf13/cobra"

// Attach registers ticket-related commands (create/list/rename) under the root.
func Attach(root *cobra.Command) error {
	ticketCmd := &cobra.Command{
		Use:   "ticket",
		Short: "Ticket workspace management",
		Long: `Ticket workspace management: create tickets, list and show them, and manage lifecycle operations.

Examples:
  # Create a ticket workspace
  docmgr ticket create --ticket MEN-4242 --title "Normalize chat API paths" --topics chat,backend

  # Show one ticket's detail
  docmgr ticket show MEN-4242

  # Close a ticket and record a changelog entry
  docmgr ticket close --ticket MEN-4242 --changelog-entry "Implementation complete"
`,
	}

	createCmd, err := newCreateCommand()
	if err != nil {
		return err
	}
	listCmd, err := newListCommand()
	if err != nil {
		return err
	}
	showCmd, err := newShowCommand()
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
	graphCmd, err := newGraphCommand()
	if err != nil {
		return err
	}

	ticketCmd.AddCommand(createCmd, listCmd, showCmd, renameCmd, closeCmd, moveCmd, graphCmd)
	root.AddCommand(ticketCmd)
	return nil
}
