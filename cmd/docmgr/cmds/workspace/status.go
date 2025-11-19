package workspace

import (
	"github.com/go-go-golems/docmgr/cmd/docmgr/cmds/common"
	"github.com/go-go-golems/docmgr/pkg/commands"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/spf13/cobra"
)

func newStatusCommand() (*cobra.Command, error) {
	cmd, err := commands.NewStatusCommand()
	if err != nil {
		return nil, err
	}
	return common.BuildCommand(
		cmd,
		cli.WithDualMode(true),
		cli.WithGlazeToggleFlag("with-glaze-output"),
	)
}
