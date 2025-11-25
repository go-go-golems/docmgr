package completion

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

// Attach enables carapace on the provided root command.
// This registers the hidden `_carapace` subcommand and bridges completions.
func Attach(root *cobra.Command) {
	carapace.Gen(root)
}
