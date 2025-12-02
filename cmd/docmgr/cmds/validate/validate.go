package validate

import (
	"github.com/go-go-golems/docmgr/cmd/docmgr/cmds/validate/validator"
	"github.com/spf13/cobra"
)

// Attach registers validation-related commands under the root.
func Attach(root *cobra.Command) error {
	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validation utilities",
	}

	frontmatterCmd, err := validator.NewFrontmatterCommand()
	if err != nil {
		return err
	}

	validateCmd.AddCommand(frontmatterCmd)
	root.AddCommand(validateCmd)
	return nil
}
