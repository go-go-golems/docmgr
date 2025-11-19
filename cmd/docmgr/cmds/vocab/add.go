package vocab

import (
	"github.com/go-go-golems/docmgr/cmd/docmgr/cmds/common"
	"github.com/go-go-golems/docmgr/pkg/commands"
	"github.com/spf13/cobra"
)

func newAddCommand() (*cobra.Command, error) {
	cmd, err := commands.NewVocabAddCommand()
	if err != nil {
		return nil, err
	}
	return common.BuildCommand(cmd)
}
