package ticket

import (
	"github.com/go-go-golems/docmgr/cmd/docmgr/cmds/common"
	"github.com/go-go-golems/docmgr/pkg/commands"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/spf13/cobra"
)

func newResumeCommand() (*cobra.Command, error) {
	return common.BuildCommand(commands.NewTicketResumeCommand(), cli.WithDualMode(true), cli.WithGlazeToggleFlag("with-glaze-output"))
}
