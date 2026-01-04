package changelog

import "github.com/spf13/cobra"

// Attach registers changelog update command under docmgr changelog update.
func Attach(root *cobra.Command) error {
	changelogCmd := &cobra.Command{
		Use:   "changelog",
		Short: "Manage changelog entries",
		Long: `Append dated entries to a ticket changelog.md.

Examples:
  # Add a short entry
  docmgr changelog update --ticket MEN-4242 --entry "Normalized chat API paths"

  # Include file notes
  docmgr changelog update --ticket MEN-4242 --entry "Refactor" --file-note "pkg/foo.go:reason"
`,
	}

	updateCmd, err := newUpdateCommand()
	if err != nil {
		return err
	}
	changelogCmd.AddCommand(updateCmd)
	root.AddCommand(changelogCmd)
	return nil
}
