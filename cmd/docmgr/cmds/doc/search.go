package doc

import (
	"github.com/carapace-sh/carapace"
	"github.com/go-go-golems/docmgr/cmd/docmgr/cmds/common"
	"github.com/go-go-golems/docmgr/pkg/commands"
	"github.com/go-go-golems/docmgr/pkg/completion"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/spf13/cobra"
)

func newSearchCommand() (*cobra.Command, error) {
	return NewSearchCommand()
}

// NewSearchCommand creates a new search command. Exported for use as an alias.
func NewSearchCommand() (*cobra.Command, error) {
	cmd, err := commands.NewSearchCommand()
	if err != nil {
		return nil, err
	}
	cobraCmd, err := common.BuildCommand(
		cmd,
		cli.WithDualMode(true),
		cli.WithGlazeToggleFlag("with-glaze-output"),
	)
	if err != nil {
		return nil, err
	}
	carapace.Gen(cobraCmd).FlagCompletion(carapace.ActionMap{
		"root":     completion.ActionDirectories(),
		"ticket":   completion.ActionTickets(),
		"topics":   completion.ActionTopics(),
		"doc-type": completion.ActionDocTypes(),
		"status":   completion.ActionStatus(),
		"order-by": carapace.ActionValues("path", "last_updated", "rank"),
		"file":     completion.ActionFiles(),
		"dir":      completion.ActionDirectories(),
	})
	return cobraCmd, nil
}
