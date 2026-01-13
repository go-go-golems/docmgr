package skill

import "github.com/spf13/cobra"

// Attach registers skill commands (list/show) as docmgr skill ...
func Attach(root *cobra.Command) error {
	skillCmd := &cobra.Command{
		Use:   "skill",
		Short: "List and show skills",
		Long: `Manage skill plans. Skills are skill.yaml plans
that package references and help output into Agent Skills format.

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
	exportCmd, err := newExportCommand()
	if err != nil {
		return err
	}
	importCmd, err := newImportCommand()
	if err != nil {
		return err
	}

	skillCmd.AddCommand(listCmd, showCmd, exportCmd, importCmd)
	root.AddCommand(skillCmd)
	return nil
}
