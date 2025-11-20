package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/go-go-golems/docmgr/internal/templates"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
)

// TemplateValidateCommand validates template files
type TemplateValidateCommand struct {
	*cmds.CommandDescription
}

// TemplateValidateSettings holds the parameters for the template validate command
type TemplateValidateSettings struct {
	Root    string `glazed.parameter:"root"`
	Path    string `glazed.parameter:"path"`
	Verbose bool   `glazed.parameter:"verbose"`
}

func NewTemplateValidateCommand() (*TemplateValidateCommand, error) {
	return &TemplateValidateCommand{
		CommandDescription: cmds.NewCommandDescription(
			"validate",
			cmds.WithShort("Validate template syntax"),
			cmds.WithLong(`Validates template files for syntax errors before runtime.

Validates one or more template files by parsing them with Go's text/template engine.
Reports syntax errors, undefined functions, and other template issues.

If --path is specified, validates only that template file.
Otherwise, scans all templates in <root>/templates/ directory.

Examples:
  # Validate all templates
  docmgr template validate

  # Validate a specific template
  docmgr template validate --path templates/status.templ

  # Verbose output showing all validated templates
  docmgr template validate --verbose
`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"root",
					parameters.ParameterTypeString,
					parameters.WithHelp("Root directory for docs (used to find templates directory)"),
					parameters.WithDefault("ttmp"),
				),
				parameters.NewParameterDefinition(
					"path",
					parameters.ParameterTypeString,
					parameters.WithHelp("Specific template file to validate (relative to root or absolute). If not specified, validates all templates."),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"verbose",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Show all validated templates, not just errors"),
					parameters.WithDefault(false),
				),
			),
		),
	}, nil
}

func (c *TemplateValidateCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &TemplateValidateSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Resolve root to absolute path
	root := workspace.ResolveRoot(settings.Root)
	templatesDir := filepath.Join(root, "templates")

	// Get FuncMap for validation (same as used in rendering)
	funcMap := getTemplateFuncMap()

	var templatePaths []string

	if settings.Path != "" {
		// Validate specific template
		var templatePath string
		if filepath.IsAbs(settings.Path) {
			templatePath = settings.Path
		} else {
			// Try relative to root first
			templatePath = filepath.Join(root, settings.Path)
			if _, err := os.Stat(templatePath); os.IsNotExist(err) {
				// Try relative to templates directory
				templatePath = filepath.Join(templatesDir, settings.Path)
			}
		}
		templatePaths = []string{templatePath}
	} else {
		// Scan all templates
		if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
			return fmt.Errorf("templates directory does not exist: %s", templatesDir)
		}
		err := filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(path, ".templ") {
				templatePaths = append(templatePaths, path)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to scan templates directory: %w", err)
		}
	}

	if len(templatePaths) == 0 {
		if settings.Path != "" {
			return fmt.Errorf("template file not found: %s", settings.Path)
		}
		fmt.Fprintf(os.Stderr, "No templates found in %s\n", templatesDir)
		return nil
	}

	// Validate each template
	errors := 0
	for _, templatePath := range templatePaths {
		err := validateTemplate(templatePath, funcMap, settings.Verbose)
		if err != nil {
			errors++
			fmt.Fprintf(os.Stderr, "ERROR: %s: %v\n", templatePath, err)
		} else if settings.Verbose {
			fmt.Fprintf(os.Stdout, "OK: %s\n", templatePath)
		}
	}

	if errors > 0 {
		return fmt.Errorf("validation failed: %d template(s) had errors", errors)
	}

	if !settings.Verbose && len(templatePaths) > 0 {
		fmt.Fprintf(os.Stdout, "All %d template(s) validated successfully\n", len(templatePaths))
	}

	return nil
}

func (c *TemplateValidateCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Template validate doesn't support Glaze output mode
	return c.Run(ctx, parsedLayers, gp)
}

// validateTemplate validates a single template file
func validateTemplate(templatePath string, funcMap template.FuncMap, verbose bool) error {
	// Read template file
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Create template with FuncMap
	tmpl := template.New(filepath.Base(templatePath)).Funcs(funcMap)

	// Parse template (this will catch syntax errors)
	_, err = tmpl.Parse(string(content))
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	return nil
}

// getTemplateFuncMap returns the same FuncMap used in template rendering
// We use the exported function from templates package
func getTemplateFuncMap() template.FuncMap {
	return templates.GetTemplateFuncMap()
}

