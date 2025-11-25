package doc

import (
	"github.com/carapace-sh/carapace"
	"github.com/go-go-golems/docmgr/cmd/docmgr/cmds/common"
	"github.com/go-go-golems/docmgr/pkg/commands"
	"github.com/go-go-golems/docmgr/pkg/completion"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/spf13/cobra"
)

func newAddCommand() (*cobra.Command, error) {
	cmd, err := commands.NewAddCommand()
	if err != nil {
		return nil, err
	}
	cmd2, err := common.BuildCommand(
		cmd,
		cli.WithDualMode(true),
		cli.WithGlazeToggleFlag("with-glaze-output"),
	)
	if err != nil {
		return nil, err
	}

	// Register dynamic flag completions
	carapace.Gen(cmd2).FlagCompletion(carapace.ActionMap{
		"ticket":           completion.ActionTickets(),
		"doc-type":         completion.ActionDocTypes(),
		"topics":           completion.ActionTopics(),
		"status":           completion.ActionStatus(),
		"intent":           completion.ActionIntent(),
		"related-files":    completion.ActionFiles().MultiParts(","),
		"external-sources": carapace.ActionValues(), // no-op placeholder for now
	})

	return cmd2, nil
}
