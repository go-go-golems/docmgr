package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"gopkg.in/yaml.v3"
)

// ConfigShowCommand shows the current configuration resolution
type ConfigShowCommand struct {
	*cmds.CommandDescription
}

// ConfigShowSettings holds the parameters for the config show command
type ConfigShowSettings struct {
	Root string `glazed.parameter:"root"`
}

func NewConfigShowCommand() (*ConfigShowCommand, error) {
	return &ConfigShowCommand{
		CommandDescription: cmds.NewCommandDescription(
			"show",
			cmds.WithShort("Show configuration resolution and active settings"),
			cmds.WithLong(`Displays the configuration resolution process, showing which configuration
source was used and the active configuration values. This helps debug configuration
issues and understand which config file is being used.

The command shows:
- Configuration sources checked in precedence order
- Which source was actually used
- Active configuration values (root, vocabulary, defaults)`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"root",
					parameters.ParameterTypeString,
					parameters.WithHelp("Root directory for docs (for testing resolution)"),
					parameters.WithDefault("ttmp"),
				),
			),
		),
	}, nil
}

func (c *ConfigShowCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &ConfigShowSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Track resolution steps
	type resolutionStep struct {
		Source string
		Status string
		Path   string
		Used   bool
		Error  string
	}

	steps := []resolutionStep{}
	var activeConfig *WorkspaceConfig
	var activeConfigPath string
	var resolvedRoot string

	// Step 1: --root flag
	providedRoot := settings.Root
	if providedRoot != "ttmp" && providedRoot != "" {
		steps = append(steps, resolutionStep{
			Source: "--root flag",
			Status: "set",
			Path:   providedRoot,
			Used:   true,
		})
		resolvedRoot = providedRoot
		if filepath.IsAbs(providedRoot) {
			resolvedRoot = providedRoot
		} else if cwd, err := os.Getwd(); err == nil {
			resolvedRoot = filepath.Join(cwd, providedRoot)
		}
	} else {
		steps = append(steps, resolutionStep{
			Source: "--root flag",
			Status: "not set",
			Used:   false,
		})
	}

	// Step 2-4: Config file search (only if root flag wasn't set)
	if providedRoot == "ttmp" || providedRoot == "" {
		// Check DOCMGR_CONFIG env var
		if env := os.Getenv("DOCMGR_CONFIG"); env != "" {
			var envPath string
			if filepath.IsAbs(env) {
				envPath = env
			} else if cwd, err := os.Getwd(); err == nil {
				envPath = filepath.Join(cwd, env)
			} else {
				envPath = env
			}
			if _, err := os.Stat(envPath); err == nil {
				steps = append(steps, resolutionStep{
					Source: "DOCMGR_CONFIG env var",
					Status: "found",
					Path:   envPath,
					Used:   true,
				})
				activeConfigPath = envPath
			} else {
				steps = append(steps, resolutionStep{
					Source: "DOCMGR_CONFIG env var",
					Status: "not found",
					Path:   envPath,
					Used:   false,
				})
			}
		}

		// Walk up directory tree for .ttmp.yaml
		cwd, _ := os.Getwd()
		dir := cwd
		configFound := false
		for {
			cfgPath := filepath.Join(dir, ".ttmp.yaml")
			if _, err := os.Stat(cfgPath); err == nil {
				steps = append(steps, resolutionStep{
					Source: fmt.Sprintf(".ttmp.yaml (walking up from %s)", cwd),
					Status: "found",
					Path:   cfgPath,
					Used:   !configFound,
				})
				if !configFound {
					activeConfigPath = cfgPath
					configFound = true
				}
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
		if !configFound {
			steps = append(steps, resolutionStep{
				Source: ".ttmp.yaml (walking up)",
				Status: "not found",
				Used:   false,
			})
		}

		// Try to load the config file
		if activeConfigPath != "" {
			data, err := os.ReadFile(activeConfigPath)
			if err == nil {
				var cfg WorkspaceConfig
				if err := yaml.Unmarshal(data, &cfg); err == nil {
					activeConfig = &cfg
					// Normalize paths
					if cfg.Root != "" && !filepath.IsAbs(cfg.Root) {
						cfg.Root = filepath.Join(filepath.Dir(activeConfigPath), cfg.Root)
					}
					if cfg.Vocabulary != "" && !filepath.IsAbs(cfg.Vocabulary) {
						cfg.Vocabulary = filepath.Join(filepath.Dir(activeConfigPath), cfg.Vocabulary)
					}
					resolvedRoot = cfg.Root
				} else {
					steps = append(steps, resolutionStep{
						Source: "Parse config file",
						Status: "error",
						Path:   activeConfigPath,
						Error:  err.Error(),
						Used:   false,
					})
				}
			}
		}
	}

	// Step 5: Git repository root fallback
	if resolvedRoot == "" || resolvedRoot == "ttmp" {
		if gitRoot, err := FindGitRoot(); err == nil && gitRoot != "" {
			gitRootPath := filepath.Join(gitRoot, "ttmp")
			steps = append(steps, resolutionStep{
				Source: "Git repository root",
				Status: "found",
				Path:   gitRootPath,
				Used:   true,
			})
			if resolvedRoot == "" || resolvedRoot == "ttmp" {
				resolvedRoot = gitRootPath
			}
		} else {
			steps = append(steps, resolutionStep{
				Source: "Git repository root",
				Status: "not found",
				Used:   false,
			})
		}
	}

	// Step 6: Default fallback
	if resolvedRoot == "" || resolvedRoot == "ttmp" {
		if cwd, err := os.Getwd(); err == nil {
			defaultPath := filepath.Join(cwd, "ttmp")
			steps = append(steps, resolutionStep{
				Source: "Default (current directory)",
				Status: "used",
				Path:   defaultPath,
				Used:   true,
			})
			resolvedRoot = defaultPath
		} else {
			steps = append(steps, resolutionStep{
				Source: "Default",
				Status: "ttmp",
				Used:   true,
			})
			resolvedRoot = "ttmp"
		}
	}

	// Output resolution steps
	row := types.NewRow(
		types.MRP("section", "Configuration Sources (in precedence order)"),
	)
	if err := gp.AddRow(ctx, row); err != nil {
		return err
	}

	for i, step := range steps {
		status := step.Status
		if step.Used {
			status = "✓ " + status
		}
		path := step.Path
		if path == "" {
			path = "<not applicable>"
		}
		if step.Error != "" {
			path = fmt.Sprintf("%s (error: %s)", path, step.Error)
		}

		row := types.NewRow(
			types.MRP("step", fmt.Sprintf("%d", i+1)),
			types.MRP("source", step.Source),
			types.MRP("status", status),
			types.MRP("path", path),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	// Output active configuration
	row = types.NewRow(
		types.MRP("section", "Active Configuration"),
	)
	if err := gp.AddRow(ctx, row); err != nil {
		return err
	}

	row = types.NewRow(
		types.MRP("setting", "root"),
		types.MRP("value", resolvedRoot),
		types.MRP("source", func() string {
			if activeConfigPath != "" {
				return activeConfigPath
			}
			return "fallback/default"
		}()),
	)
	if err := gp.AddRow(ctx, row); err != nil {
		return err
	}

	if activeConfig != nil {
		if activeConfig.Vocabulary != "" {
			row = types.NewRow(
				types.MRP("setting", "vocabulary"),
				types.MRP("value", activeConfig.Vocabulary),
				types.MRP("source", activeConfigPath),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}

		if len(activeConfig.Defaults.Owners) > 0 {
			row = types.NewRow(
				types.MRP("setting", "defaults.owners"),
				types.MRP("value", fmt.Sprintf("%v", activeConfig.Defaults.Owners)),
				types.MRP("source", activeConfigPath),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}

		if activeConfig.Defaults.Intent != "" {
			row = types.NewRow(
				types.MRP("setting", "defaults.intent"),
				types.MRP("value", activeConfig.Defaults.Intent),
				types.MRP("source", activeConfigPath),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}

		if activeConfig.FilenamePrefixPolicy != "" {
			row = types.NewRow(
				types.MRP("setting", "filenamePrefixPolicy"),
				types.MRP("value", activeConfig.FilenamePrefixPolicy),
				types.MRP("source", activeConfigPath),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}
	}

	return nil
}

// Run provides human-friendly output for config show command
func (c *ConfigShowCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
) error {
	settings := &ConfigShowSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Track resolution steps
	type resolutionStep struct {
		Source string
		Status string
		Path   string
		Used   bool
		Error  string
	}

	steps := []resolutionStep{}
	var activeConfig *WorkspaceConfig
	var activeConfigPath string
	var resolvedRoot string

	// Step 1: --root flag
	providedRoot := settings.Root
	if providedRoot != "ttmp" && providedRoot != "" {
		steps = append(steps, resolutionStep{
			Source: "--root flag",
			Status: "set",
			Path:   providedRoot,
			Used:   true,
		})
		resolvedRoot = providedRoot
		if filepath.IsAbs(providedRoot) {
			resolvedRoot = providedRoot
		} else if cwd, err := os.Getwd(); err == nil {
			resolvedRoot = filepath.Join(cwd, providedRoot)
		}
	} else {
		steps = append(steps, resolutionStep{
			Source: "--root flag",
			Status: "<not set>",
			Path:   "",
			Used:   false,
		})
	}

	// Step 2-4: Config file search (only if root flag wasn't set)
	if providedRoot == "ttmp" || providedRoot == "" {
		// Check DOCMGR_CONFIG env var
		if env := os.Getenv("DOCMGR_CONFIG"); env != "" {
			var envPath string
			if filepath.IsAbs(env) {
				envPath = env
			} else if cwd, err := os.Getwd(); err == nil {
				envPath = filepath.Join(cwd, env)
			} else {
				envPath = env
			}
			if _, err := os.Stat(envPath); err == nil {
				steps = append(steps, resolutionStep{
					Source: "DOCMGR_CONFIG env var",
					Status: "found",
					Path:   envPath,
					Used:   true,
				})
				activeConfigPath = envPath
			} else {
				steps = append(steps, resolutionStep{
					Source: "DOCMGR_CONFIG env var",
					Status: "<not found>",
					Path:   envPath,
					Used:   false,
				})
			}
		}

		// Walk up directory tree for .ttmp.yaml
		cwd, _ := os.Getwd()
		dir := cwd
		configFound := false
		firstFoundPath := ""
		for {
			cfgPath := filepath.Join(dir, ".ttmp.yaml")
			if _, err := os.Stat(cfgPath); err == nil {
				if !configFound {
					firstFoundPath = cfgPath
					activeConfigPath = cfgPath
					configFound = true
				}
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
		if configFound {
			relPath, _ := filepath.Rel(cwd, firstFoundPath)
			if relPath == ".ttmp.yaml" {
				relPath = ".ttmp.yaml (current dir)"
			} else {
				relPath = fmt.Sprintf("%s (walking up)", relPath)
			}
			steps = append(steps, resolutionStep{
				Source: ".ttmp.yaml",
				Status: "found",
				Path:   relPath,
				Used:   true,
			})
		} else {
			steps = append(steps, resolutionStep{
				Source: ".ttmp.yaml (walking up)",
				Status: "<not found>",
				Used:   false,
			})
		}

		// Try to load the config file
		if activeConfigPath != "" {
			data, err := os.ReadFile(activeConfigPath)
			if err == nil {
				var cfg WorkspaceConfig
				if err := yaml.Unmarshal(data, &cfg); err == nil {
					activeConfig = &cfg
					// Normalize paths
					if cfg.Root != "" && !filepath.IsAbs(cfg.Root) {
						cfg.Root = filepath.Join(filepath.Dir(activeConfigPath), cfg.Root)
					}
					if cfg.Vocabulary != "" && !filepath.IsAbs(cfg.Vocabulary) {
						cfg.Vocabulary = filepath.Join(filepath.Dir(activeConfigPath), cfg.Vocabulary)
					}
					resolvedRoot = cfg.Root
				} else {
					steps = append(steps, resolutionStep{
						Source: "Parse config file",
						Status: "error",
						Path:   activeConfigPath,
						Error:  err.Error(),
						Used:   false,
					})
				}
			}
		}
	}

	// Step 5: Git repository root fallback
	if resolvedRoot == "" || resolvedRoot == "ttmp" {
		if gitRoot, err := FindGitRoot(); err == nil && gitRoot != "" {
			gitRootPath := filepath.Join(gitRoot, "ttmp")
			steps = append(steps, resolutionStep{
				Source: "Git repository root",
				Status: "found",
				Path:   gitRootPath,
				Used:   true,
			})
			if resolvedRoot == "" || resolvedRoot == "ttmp" {
				resolvedRoot = gitRootPath
			}
		} else {
			steps = append(steps, resolutionStep{
				Source: "Git repository root",
				Status: "<not found>",
				Used:   false,
			})
		}
	}

	// Step 6: Default fallback
	if resolvedRoot == "" || resolvedRoot == "ttmp" {
		if cwd, err := os.Getwd(); err == nil {
			defaultPath := filepath.Join(cwd, "ttmp")
			steps = append(steps, resolutionStep{
				Source: "Default (current directory)",
				Status: "used",
				Path:   defaultPath,
				Used:   true,
			})
			resolvedRoot = defaultPath
		} else {
			steps = append(steps, resolutionStep{
				Source: "Default",
				Status: "ttmp",
				Used:   true,
			})
			resolvedRoot = "ttmp"
		}
	}

	// Output resolution steps
	fmt.Println("Configuration sources (in precedence order):")
	for i, step := range steps {
		marker := " "
		if step.Used {
			marker = "✓"
		}
		path := step.Path
		if path == "" {
			path = "<not applicable>"
		}
		if step.Error != "" {
			path = fmt.Sprintf("%s (error: %s)", path, step.Error)
		}
		fmt.Printf("  %d. %s %s: %s\n", i+1, marker, step.Source, path)
	}

	fmt.Println("\nActive configuration:")
	fmt.Printf("  root: %s\n", resolvedRoot)
	if activeConfigPath != "" {
		cwd, _ := os.Getwd()
		relPath, err := filepath.Rel(cwd, activeConfigPath)
		if err != nil || len(relPath) > len(activeConfigPath) {
			relPath = activeConfigPath
		}
		fmt.Printf("  source: %s\n", relPath)
	} else {
		fmt.Printf("  source: fallback/default\n")
	}

	if activeConfig != nil {
		if activeConfig.Vocabulary != "" {
			fmt.Printf("  vocabulary: %s\n", activeConfig.Vocabulary)
		}
		if len(activeConfig.Defaults.Owners) > 0 {
			fmt.Printf("  defaults.owners: %v\n", activeConfig.Defaults.Owners)
		}
		if activeConfig.Defaults.Intent != "" {
			fmt.Printf("  defaults.intent: %s\n", activeConfig.Defaults.Intent)
		}
		if activeConfig.FilenamePrefixPolicy != "" {
			fmt.Printf("  filenamePrefixPolicy: %s\n", activeConfig.FilenamePrefixPolicy)
		}
	}

	return nil
}

var _ cmds.GlazeCommand = &ConfigShowCommand{}
var _ cmds.BareCommand = &ConfigShowCommand{}
