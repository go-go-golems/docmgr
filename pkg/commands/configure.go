package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"gopkg.in/yaml.v3"
)

// ConfigureCommand writes a .ttmp.yaml configuration file at the repository root
type ConfigureCommand struct {
	*cmds.CommandDescription
}

// ConfigureSettings holds parameters for writing .ttmp.yaml
type ConfigureSettings struct {
	Root       string   `glazed.parameter:"root"`
	Owners     []string `glazed.parameter:"owners"`
	Intent     string   `glazed.parameter:"intent"`
	Vocabulary string   `glazed.parameter:"vocabulary"`
	Force      bool     `glazed.parameter:"force"`
}

func NewConfigureCommand() (*ConfigureCommand, error) {
	return &ConfigureCommand{
		CommandDescription: cmds.NewCommandDescription(
			"configure",
			cmds.WithShort("Create or update a .ttmp.yaml at the repository root"),
			cmds.WithLong(`Writes a .ttmp.yaml configuration file to the nearest repository root.

Examples:
  # Write default config pointing to ttmp/
  docmgr configure

  # Explicit values
  docmgr configure --root ttmp --owners manuel,alice --intent long-term --vocabulary ttmp/vocabulary.yaml
`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"root",
					parameters.ParameterTypeString,
					parameters.WithHelp("Docs root path (relative to repo root unless absolute)"),
					parameters.WithDefault("ttmp"),
				),
				parameters.NewParameterDefinition(
					"owners",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Default owners (comma-separated)"),
					parameters.WithDefault([]string{}),
				),
				parameters.NewParameterDefinition(
					"intent",
					parameters.ParameterTypeString,
					parameters.WithHelp("Default intent for new tickets"),
					parameters.WithDefault("long-term"),
				),
				parameters.NewParameterDefinition(
					"vocabulary",
					parameters.ParameterTypeString,
					parameters.WithHelp("Vocabulary path (defaults to <root>/vocabulary.yaml)"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"force",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Overwrite existing .ttmp.yaml if present"),
					parameters.WithDefault(false),
				),
			),
		),
	}, nil
}

func (c *ConfigureCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &ConfigureSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Find repository root (shared helper from vocab_add)
	repoRoot, err := workspace.FindRepositoryRoot()
	if err != nil {
		return fmt.Errorf("failed to find repository root: %w", err)
	}

	// Target path for config
	cfgPath := filepath.Join(repoRoot, ".ttmp.yaml")

	if _, err := os.Stat(cfgPath); err == nil && !settings.Force {
		// Do not overwrite existing config unless --force
		row := types.NewRow(
			types.MRP("config", cfgPath),
			types.MRP("root", settings.Root),
			types.MRP("vocabulary", settings.Vocabulary),
			types.MRP("status", "exists"),
		)
		return gp.AddRow(ctx, row)
	}

	// Build config structure; keep relative paths relative to cfg file directory
	cfg := workspace.WorkspaceConfig{
		Root: settings.Root,
		Vocabulary: func() string {
			if strings.TrimSpace(settings.Vocabulary) != "" {
				return settings.Vocabulary
			}
			// default to <root>/vocabulary.yaml
			return filepath.ToSlash(filepath.Join(settings.Root, "vocabulary.yaml"))
		}(),
	}
	cfg.Defaults.Owners = append([]string{}, settings.Owners...)
	cfg.Defaults.Intent = settings.Intent

	// Serialize YAML
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Ensure parent dir exists (repo root should exist)
	if err := os.WriteFile(cfgPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", cfgPath, err)
	}

	// Echo resolved context prior to write-like behavior
	absRoot := settings.Root
	if !filepath.IsAbs(absRoot) {
		absRoot = filepath.Join(repoRoot, settings.Root)
	}
	fmt.Printf("root=%s config=%s vocabulary=%s\n", absRoot, cfgPath, cfg.Vocabulary)

	row := types.NewRow(
		types.MRP("config", cfgPath),
		types.MRP("root", cfg.Root),
		types.MRP("vocabulary", cfg.Vocabulary),
		types.MRP("status", "written"),
	)
	return gp.AddRow(ctx, row)
}

var _ cmds.GlazeCommand = &ConfigureCommand{}
