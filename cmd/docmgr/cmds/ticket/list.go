package ticket

import (
	"github.com/go-go-golems/docmgr/cmd/docmgr/cmds/common"
	"github.com/go-go-golems/docmgr/pkg/commands"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/spf13/cobra"
)

func newListCommand() (*cobra.Command, error) {
	cmd, err := commands.NewListTicketsCommand()
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
	// Allow `docmgr ticket list` as an alias to `docmgr ticket tickets`
	cobraCmd.Aliases = append(cobraCmd.Aliases, "list")
	return cobraCmd, nil
}
