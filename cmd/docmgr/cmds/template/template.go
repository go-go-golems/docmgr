package template

import "github.com/spf13/cobra"

// Attach registers template commands (validate) as docmgr template ...
func Attach(root *cobra.Command) error {
	templateCmd := &cobra.Command{
		Use:   "template",
		Short: "Manage and validate templates",
		Long: `Validate and debug docmgr output templates (used for rich human-mode rendering).

Examples:
  # Validate all templates under <root>/templates/
  docmgr template validate

  # Validate one template file
  docmgr template validate --path /tmp/example.templ
`,
	}

	validateCmd, err := newValidateCommand()
	if err != nil {
		return err
	}

	templateCmd.AddCommand(validateCmd)
	root.AddCommand(templateCmd)
	return nil
}
