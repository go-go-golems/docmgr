package meta

import (
	"github.com/carapace-sh/carapace"
	"github.com/go-go-golems/docmgr/cmd/docmgr/cmds/common"
	"github.com/go-go-golems/docmgr/pkg/commands"
	"github.com/go-go-golems/docmgr/pkg/completion"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/spf13/cobra"
)

func newUpdateCommand() (*cobra.Command, error) {
	cmd, err := commands.NewMetaUpdateCommand()
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
		"doc":      completion.ActionFiles(),
		"ticket":   completion.ActionTickets(),
		"doc-type": completion.ActionDocTypes(),
		"field":    completion.ActionMetaFields(),
		"value":    completion.ActionMetaValue(),
		"root":     completion.ActionDirectories(),
	})
	return cobraCmd, nil
}
