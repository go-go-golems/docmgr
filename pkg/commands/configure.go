package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
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
	Root       string   `glazed:"root"`
	Owners     []string `glazed:"owners"`
	Intent     string   `glazed:"intent"`
	Vocabulary string   `glazed:"vocabulary"`
	Force      bool     `glazed:"force"`
}

type ConfigureResult struct {
	ConfigPath string
	Root       string
	Vocabulary string
	Status     string
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

  # Overwrite an existing .ttmp.yaml
  docmgr configure --force --root ttmp
`),
			cmds.WithFlags(
				fields.New(
					"root",
					fields.TypeString,
					fields.WithHelp("Docs root path (relative to repo root unless absolute)"),
					fields.WithDefault("ttmp"),
				),
				fields.New(
					"owners",
					fields.TypeStringList,
					fields.WithHelp("Default owners (comma-separated)"),
					fields.WithDefault([]string{}),
				),
				fields.New(
					"intent",
					fields.TypeString,
					fields.WithHelp("Default intent for new tickets"),
					fields.WithDefault("long-term"),
				),
				fields.New(
					"vocabulary",
					fields.TypeString,
					fields.WithHelp("Vocabulary path (defaults to <root>/vocabulary.yaml)"),
					fields.WithDefault(""),
				),
				fields.New(
					"force",
					fields.TypeBool,
					fields.WithHelp("Overwrite existing .ttmp.yaml if present"),
					fields.WithDefault(false),
				),
			),
		),
	}, nil
}

func (c *ConfigureCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedValues *values.Values,
	gp middlewares.Processor,
) error {
	settings := &ConfigureSettings{}
	if err := parsedValues.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	result, err := c.writeConfig(settings)
	if err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("config", result.ConfigPath),
		types.MRP("root", result.Root),
		types.MRP("vocabulary", result.Vocabulary),
		types.MRP("status", result.Status),
	)
	return gp.AddRow(ctx, row)
}

var _ cmds.GlazeCommand = &ConfigureCommand{}

func (c *ConfigureCommand) writeConfig(settings *ConfigureSettings) (*ConfigureResult, error) {
	repoRoot, err := workspace.FindRepositoryRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find repository root: %w", err)
	}

	cfgPath := filepath.Join(repoRoot, ".ttmp.yaml")

	if _, err := os.Stat(cfgPath); err == nil && !settings.Force {
		return &ConfigureResult{
			ConfigPath: cfgPath,
			Root:       settings.Root,
			Vocabulary: settings.Vocabulary,
			Status:     "exists",
		}, nil
	}

	cfg := workspace.WorkspaceConfig{
		Root: settings.Root,
		Vocabulary: func() string {
			if strings.TrimSpace(settings.Vocabulary) != "" {
				return settings.Vocabulary
			}
			return filepath.ToSlash(filepath.Join(settings.Root, "vocabulary.yaml"))
		}(),
	}
	cfg.Defaults.Owners = append([]string{}, settings.Owners...)
	cfg.Defaults.Intent = settings.Intent

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(cfgPath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write %s: %w", cfgPath, err)
	}

	return &ConfigureResult{
		ConfigPath: cfgPath,
		Root:       cfg.Root,
		Vocabulary: cfg.Vocabulary,
		Status:     "written",
	}, nil
}

func (c *ConfigureCommand) Run(
	ctx context.Context,
	parsedValues *values.Values,
) error {
	settings := &ConfigureSettings{}
	if err := parsedValues.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	result, err := c.writeConfig(settings)
	if err != nil {
		return err
	}

	fmt.Printf("Configuration file: %s\n", result.ConfigPath)
	fmt.Printf("- Root: %s\n", result.Root)
	fmt.Printf("- Vocabulary: %s\n", result.Vocabulary)
	fmt.Printf("- Status: %s\n", result.Status)

	return nil
}

var _ cmds.BareCommand = &ConfigureCommand{}
