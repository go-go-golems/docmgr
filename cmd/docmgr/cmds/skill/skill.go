package skill

import "github.com/spf13/cobra"

// Attach registers skill commands (list/show) as docmgr skill ...
func Attach(root *cobra.Command) error {
	skillCmd := &cobra.Command{
		Use:   "skill",
		Short: "List and show skills",
		Long: `Manage skills documentation. Skills are documents with DocType=skill
that provide structured information about what a skill is for and when to use it.

Examples:
 
    docmgr skill list --topics tooling

    docmgr skill show "Test-Driven Development"
`,
	}

	listCmd, err := newListCommand()
	if err != nil {
		return err
	}
	showCmd, err := newShowCommand()
	if err != nil {
		return err
	}

	skillCmd.AddCommand(listCmd, showCmd)
	root.AddCommand(skillCmd)
	return nil
}
