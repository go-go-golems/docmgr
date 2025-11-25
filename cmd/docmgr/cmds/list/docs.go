package list

import (
	"github.com/carapace-sh/carapace"
	"github.com/go-go-golems/docmgr/cmd/docmgr/cmds/common"
	"github.com/go-go-golems/docmgr/pkg/commands"
	"github.com/go-go-golems/docmgr/pkg/completion"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/spf13/cobra"
)

func newDocsCommand() (*cobra.Command, error) {
	cmd, err := commands.NewListDocsCommand()
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
	cobraCmd.Use = "docs"
	cobraCmd.Short = "List documents"
	carapace.Gen(cobraCmd).FlagCompletion(carapace.ActionMap{
		"root":     completion.ActionDirectories(),
		"ticket":   completion.ActionTickets(),
		"status":   completion.ActionStatus(),
		"doc-type": completion.ActionDocTypes(),
		"topics":   completion.ActionTopics(),
	})
	return cobraCmd, nil
}
