package vocab

import "github.com/spf13/cobra"

// Attach registers vocabulary commands (list/add) as docmgr vocab ...
func Attach(root *cobra.Command) error {
	vocabCmd := &cobra.Command{
		Use:   "vocab",
		Short: "Manage workspace vocabulary",
		Long: `Manage vocabulary entries used to validate Topics/DocTypes/Status/Intent.

Examples:
  # List vocabulary entries
  docmgr vocab list --category topics

  # Add a new topic
  docmgr vocab add --category topics --slug observability --description "Logging and metrics"
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

	vocabCmd.AddCommand(listCmd, addCmd)
	root.AddCommand(vocabCmd)
	return nil
}
