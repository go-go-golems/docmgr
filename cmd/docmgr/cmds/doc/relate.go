package doc

import (
	"github.com/carapace-sh/carapace"
	"github.com/go-go-golems/docmgr/cmd/docmgr/cmds/common"
	"github.com/go-go-golems/docmgr/pkg/commands"
	"github.com/go-go-golems/docmgr/pkg/completion"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/spf13/cobra"
)

func newRelateCommand() (*cobra.Command, error) {
	cmd, err := commands.NewRelateCommand()
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
		"ticket":       completion.ActionTickets(),
		"doc":          completion.ActionFiles(),
		"file-note":    completion.ActionFiles().MultiParts(":", ":"), // left is file path, right freeform
		"remove-files": completion.ActionFiles().MultiParts(","),
		"topics":       completion.ActionTopics(),
		"root":         completion.ActionDirectories(),
	})
	return cobraCmd, nil
}
