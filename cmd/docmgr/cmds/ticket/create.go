package ticket

import (
	"github.com/go-go-golems/docmgr/cmd/docmgr/cmds/common"
	"github.com/go-go-golems/docmgr/pkg/commands"
	"github.com/spf13/cobra"
)

func newCreateCommand() (*cobra.Command, error) {
	cmd, err := commands.NewCreateTicketCommand()
	if err != nil {
		return nil, err
	}
	return common.BuildCommand(cmd)
}
