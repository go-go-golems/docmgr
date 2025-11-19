package importcmd

import (
	"github.com/go-go-golems/docmgr/cmd/docmgr/cmds/common"
	"github.com/go-go-golems/docmgr/pkg/commands"
	"github.com/spf13/cobra"
)

func newFileCommand() (*cobra.Command, error) {
	cmd, err := commands.NewImportFileCommand()
	if err != nil {
		return nil, err
	}
	return common.BuildCommand(cmd)
}
