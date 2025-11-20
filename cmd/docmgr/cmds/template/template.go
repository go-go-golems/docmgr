package template

import "github.com/spf13/cobra"

// Attach registers template commands (validate) as docmgr template ...
func Attach(root *cobra.Command) error {
	templateCmd := &cobra.Command{
		Use:   "template",
		Short: "Manage and validate templates",
	}

	validateCmd, err := newValidateCommand()
	if err != nil {
		return err
	}

	templateCmd.AddCommand(validateCmd)
	root.AddCommand(templateCmd)
	return nil
}

