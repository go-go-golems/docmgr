package ticket

import (
	"github.com/carapace-sh/carapace"
	"github.com/go-go-golems/docmgr/cmd/docmgr/cmds/common"
	"github.com/go-go-golems/docmgr/pkg/commands"
	"github.com/go-go-golems/docmgr/pkg/completion"
	"github.com/spf13/cobra"
)

func newRenameCommand() (*cobra.Command, error) {
	cmd, err := commands.NewRenameTicketCommand()
	if err != nil {
		return nil, err
	}
	cobraCmd, err := common.BuildCommand(cmd)
	if err != nil {
		return nil, err
	}
	carapace.Gen(cobraCmd).FlagCompletion(carapace.ActionMap{
		"ticket": completion.ActionTickets(),
		"root":   completion.ActionDirectories(),
	})
	return cobraCmd, nil
}
