package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-go-golems/docmgr/internal/documents"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/docmgr/pkg/models"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"gopkg.in/yaml.v3"
)

// ImportFileCommand imports a file into the document workspace
type ImportFileCommand struct {
	*cmds.CommandDescription
}

// ImportFileSettings holds the parameters for the import file command
type ImportFileSettings struct {
	Ticket   string `glazed.parameter:"ticket"`
	FilePath string `glazed.parameter:"file"`
	Root     string `glazed.parameter:"root"`
	Name     string `glazed.parameter:"name"`
}

func NewImportFileCommand() (*ImportFileCommand, error) {
	return &ImportFileCommand{
		CommandDescription: cmds.NewCommandDescription(
			"file",
			cmds.WithShort("Import a file into the document workspace"),
			cmds.WithLong(`Imports a local file into the sources directory of a document workspace.

Example:
  docmgr import file --ticket MEN-3475 --file /path/to/doc.md
  docmgr import file --ticket MEN-3475 --file /path/to/spec.pdf --name "API Spec"
`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"ticket",
					parameters.ParameterTypeString,
					parameters.WithHelp("Ticket identifier"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"file",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path to file to import"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"root",
					parameters.ParameterTypeString,
					parameters.WithHelp("Root directory for docs"),
					parameters.WithDefault("ttmp"),
				),
				parameters.NewParameterDefinition(
					"name",
					parameters.ParameterTypeString,
					parameters.WithHelp("Optional name for the imported file"),
					parameters.WithDefault(""),
				),
			),
		),
	}, nil
}

func (c *ImportFileCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &ImportFileSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Apply config root if present
	settings.Root = workspace.ResolveRoot(settings.Root)

	// Find the ticket directory
	ticketDir, err := findTicketDirectory(settings.Root, settings.Ticket)
	if err != nil {
		return fmt.Errorf("failed to find ticket directory: %w", err)
	}

	// Check if source file exists
	if _, err := os.Stat(settings.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("source file does not exist: %s", settings.FilePath)
	}

	// Create sources directory if it doesn't exist
	sourcesDir := filepath.Join(ticketDir, "sources", "local")
	if err := os.MkdirAll(sourcesDir, 0755); err != nil {
		return fmt.Errorf("failed to create sources directory: %w", err)
	}

	// Determine destination filename
	destName := filepath.Base(settings.FilePath)
	if settings.Name != "" {
		ext := filepath.Ext(settings.FilePath)
		destName = settings.Name + ext
	}
	destPath := filepath.Join(sourcesDir, destName)

	// Copy the file
	input, err := os.ReadFile(settings.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	if err := os.WriteFile(destPath, input, 0644); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	// Create metadata file
	source := models.ExternalSource{
		Type:        "local",
		Path:        settings.FilePath,
		LastFetched: time.Now(),
	}

	metaPath := filepath.Join(ticketDir, ".meta", "sources.yaml")
	if err := appendSourceMetadata(metaPath, &source); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	// Update index.md to add external source reference
	indexPath := filepath.Join(ticketDir, "index.md")
	doc, body, err := documents.ReadDocumentWithFrontmatter(indexPath)
	if err != nil {
		return fmt.Errorf("failed to read index.md: %w", err)
	}

	sourceRef := fmt.Sprintf("local:%s", destName)
	if !contains(doc.ExternalSources, sourceRef) {
		doc.ExternalSources = append(doc.ExternalSources, sourceRef)
		doc.LastUpdated = time.Now()

		if err := documents.WriteDocumentWithFrontmatter(indexPath, doc, body, true); err != nil {
			return fmt.Errorf("failed to update index.md: %w", err)
		}
	}

	row := types.NewRow(
		types.MRP("ticket", settings.Ticket),
		types.MRP("source_file", settings.FilePath),
		types.MRP("destination", destPath),
		types.MRP("type", "local"),
		types.MRP("status", "imported"),
	)

	if err := gp.AddRow(ctx, row); err != nil {
		return fmt.Errorf("failed to add import row for %s: %w", settings.FilePath, err)
	}
	return nil
}

func findTicketDirectory(root, ticket string) (string, error) {
	workspaces, err := workspace.CollectTicketWorkspaces(root, nil)
	if err != nil {
		return "", err
	}
	for _, ws := range workspaces {
		if ws.Doc != nil && ws.Doc.Ticket == ticket {
			return ws.Path, nil
		}
	}
	return "", fmt.Errorf("ticket not found: %s", ticket)
}

func appendSourceMetadata(path string, source *models.ExternalSource) error {
	var sources []models.ExternalSource

	// Read existing sources if file exists
	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read external sources file %s: %w", path, err)
		}
		if err := yaml.Unmarshal(data, &sources); err != nil {
			return fmt.Errorf("failed to parse external sources file %s: %w", path, err)
		}
	}

	sources = append(sources, *source)

	data, err := yaml.Marshal(sources)
	if err != nil {
		return fmt.Errorf("failed to encode external sources for %s: %w", path, err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write external sources file %s: %w", path, err)
	}
	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

var _ cmds.GlazeCommand = &ImportFileCommand{}
