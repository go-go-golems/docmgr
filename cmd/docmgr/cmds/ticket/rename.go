package ticket

import (
	"github.com/go-go-golems/docmgr/cmd/docmgr/cmds/common"
	"github.com/go-go-golems/docmgr/pkg/commands"
	"github.com/spf13/cobra"
)

func newRenameCommand() (*cobra.Command, error) {
	cmd, err := commands.NewRenameTicketCommand()
	if err != nil {
		return nil, err
	}
	return common.BuildCommand(cmd)
}
