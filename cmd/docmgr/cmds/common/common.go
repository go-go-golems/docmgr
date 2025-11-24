package common

import (
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/spf13/cobra"
)

// BuildCommand wraps cli.BuildCobraCommand with the default parser configuration
// and middleware wiring used across docmgr commands.
func BuildCommand(glazedCmd cmds.Command, opts ...cli.CobraOption) (*cobra.Command, error) {
	// Ensure glazed layer defaults to JSON output in Glaze mode.
	// This only affects structured output; classic mode (BareCommand) is unchanged.
	if _, isGlaze := glazedCmd.(cmds.GlazeCommand); isGlaze {
		desc := glazedCmd.Description()
		gpl, err := settings.NewGlazedParameterLayers(
			settings.WithOutputParameterLayerOptions(
				layers.WithDefaults(map[string]interface{}{"output": "json"}),
			),
		)
		if err != nil {
			return nil, err
		}
		desc.Layers.Set(settings.GlazedSlug, gpl)
	}

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
