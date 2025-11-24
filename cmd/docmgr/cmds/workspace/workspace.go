package workspace

import "github.com/spf13/cobra"

// Attach registers workspace-wide commands (init/configure/status/doctor) both at the root level
// and under a namespaced "workspace" command to match documentation references.
func Attach(root *cobra.Command) error {
	workspaceCmd := &cobra.Command{
		Use:   "workspace",
		Short: "Workspace initialization and configuration commands",
	}

	builders := []func() (*cobra.Command, error){
		newInitCommand,
		newConfigureCommand,
		newStatusCommand,
		newDoctorCommand,
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
