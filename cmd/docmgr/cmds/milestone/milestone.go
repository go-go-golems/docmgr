package milestone

import (
	"github.com/go-go-golems/docmgr/cmd/docmgr/cmds/common"
	"github.com/go-go-golems/docmgr/pkg/commands"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/spf13/cobra"
)

func Attach(root *cobra.Command) error {
	group := &cobra.Command{Use: "milestone", Short: "Recoverable ticket milestones"}
	command, err := common.BuildCommand(commands.NewMilestoneCommand(), cli.WithDualMode(true), cli.WithGlazeToggleFlag("with-glaze-output"))
	if err != nil {
		return err
	}
	group.AddCommand(command)
	root.AddCommand(group)
	return nil
}
