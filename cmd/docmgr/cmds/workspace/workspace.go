package workspace

import "github.com/spf13/cobra"

// Attach registers workspace-wide commands (init/configure/status/doctor) both at the root level
// and under a namespaced "workspace" command to match documentation references.
func Attach(root *cobra.Command) error {
	workspaceCmd := &cobra.Command{
		Use:   "workspace",
		Short: "Workspace initialization and configuration commands",
		Long: `Workspace-wide commands, also available at the root for convenience (init/status/doctor/configure/export-sqlite).

Examples:
  # Initialize the docs root
  docmgr workspace init

  # Show workspace status (namespaced form)
  docmgr workspace status --summary-only
`,
	}

	builders := []func() (*cobra.Command, error){
		newInitCommand,
		newConfigureCommand,
		newStatusCommand,
		newDoctorCommand,
		newExportSQLiteCommand,
	}

	for _, builder := range builders {
		cmd, err := builder()
		if err != nil {
			return err
		}
		root.AddCommand(cmd)

		nestedCmd, err := builder()
		if err != nil {
			return err
		}
		workspaceCmd.AddCommand(nestedCmd)
	}

	root.AddCommand(workspaceCmd)
	return nil
}
