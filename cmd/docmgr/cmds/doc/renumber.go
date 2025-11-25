package doc

import (
	"github.com/carapace-sh/carapace"
	"github.com/go-go-golems/docmgr/cmd/docmgr/cmds/common"
	"github.com/go-go-golems/docmgr/pkg/commands"
	"github.com/go-go-golems/docmgr/pkg/completion"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/spf13/cobra"
)

func newRenumberCommand() (*cobra.Command, error) {
	cmd, err := commands.NewRenumberCommand()
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
		"ticket": completion.ActionTickets(),
		"root":   completion.ActionDirectories(),
	})
	return cobraCmd, nil
}
