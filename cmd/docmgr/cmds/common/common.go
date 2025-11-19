package common

import (
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/spf13/cobra"
)

// BuildCommand wraps cli.BuildCobraCommand with the default parser configuration
// and middleware wiring used across docmgr commands.
func BuildCommand(glazedCmd cmds.Command, opts ...cli.CobraOption) (*cobra.Command, error) {
	defaultOpts := []cli.CobraOption{
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpLayers: []string{layers.DefaultSlug},
			MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
		}),
		cli.WithCobraMiddlewaresFunc(cli.CobraCommandDefaultMiddlewares),
		cli.WithCobraShortHelpLayers(layers.DefaultSlug),
	}
	defaultOpts = append(defaultOpts, opts...)
	return cli.BuildCobraCommand(glazedCmd, defaultOpts...)
}
