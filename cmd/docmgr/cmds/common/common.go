package common

import (
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/spf13/cobra"
)

// BuildCommand wraps cli.BuildCobraCommand with the default parser configuration
// and middleware wiring used across docmgr commands.
func BuildCommand(glazedCmd cmds.Command, opts ...cli.CobraOption) (*cobra.Command, error) {
	// Ensure the glazed section defaults to JSON output in glaze mode.
	// This only affects structured output; classic mode (BareCommand) is unchanged.
	if _, isGlaze := glazedCmd.(cmds.GlazeCommand); isGlaze {
		desc := glazedCmd.Description()
		glazedSection, err := settings.NewGlazedSection(
			settings.WithOutputSectionOptions(
				schema.WithDefaults(map[string]interface{}{"output": "json"}),
			),
		)
		if err != nil {
			return nil, err
		}
		desc.Schema.Set(settings.GlazedSlug, glazedSection)
	}

	defaultOpts := []cli.CobraOption{
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpSections: []string{schema.DefaultSlug},
			MiddlewaresFunc:   cli.CobraCommandDefaultMiddlewares,
		}),
		cli.WithCobraMiddlewaresFunc(cli.CobraCommandDefaultMiddlewares),
		cli.WithCobraShortHelpSections(schema.DefaultSlug),
	}
	defaultOpts = append(defaultOpts, opts...)
	return cli.BuildCobraCommand(glazedCmd, defaultOpts...)
}
