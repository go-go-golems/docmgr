package ignorecmd

import "github.com/spf13/cobra"

func Attach(root *cobra.Command) error {
	ignoreCmd := &cobra.Command{
		Use:   "ignore",
		Short: "Inspect docmgr ignore decisions",
		Long: `Inspect the workspace-owned docmgr ignore policy.

Examples:
  docmgr ignore explain ttmp/2026/06/08/TICKET--slug/scripts/node_modules/pkg/README.md
  docmgr ignore explain --trace ttmp/2026/06/08/TICKET--slug/reference/01-plan.md
`,
	}

	explainCmd, err := newExplainCommand()
	if err != nil {
		return err
	}
	ignoreCmd.AddCommand(explainCmd)
	root.AddCommand(ignoreCmd)
	return nil
}
