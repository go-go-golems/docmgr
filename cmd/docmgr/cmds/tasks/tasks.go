package tasks

import "github.com/spf13/cobra"

// Attach registers the tasks command tree under the root (docmgr tasks ...).
func Attach(root *cobra.Command) error {
	tasksCmd := &cobra.Command{
		Use:     "task",
		Aliases: []string{"tasks"},
		Short:   "Manage ticket task lists",
		Long: `Manage checkbox tasks stored in a ticket's tasks.md.

Examples:
  # List tasks for a ticket
  docmgr task list --ticket MEN-4242

  # Add and complete a task
  docmgr task add --ticket MEN-4242 --text "Write design doc"

  docmgr task check --ticket MEN-4242 --id 1
`,
	}

	listCmd, err := newListCommand()
	if err != nil {
		return err
	}
	addCmd, err := newAddCommand()
	if err != nil {
		return err
	}
	checkCmd, err := newCheckCommand()
	if err != nil {
		return err
	}
	uncheckCmd, err := newUncheckCommand()
	if err != nil {
		return err
	}
	editCmd, err := newEditCommand()
	if err != nil {
		return err
	}
	removeCmd, err := newRemoveCommand()
	if err != nil {
		return err
	}

	tasksCmd.AddCommand(
		listCmd,
		addCmd,
		checkCmd,
		uncheckCmd,
		editCmd,
		removeCmd,
	)
	root.AddCommand(tasksCmd)
	return nil
}
