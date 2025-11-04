package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/go-go-golems/docmgr/pkg/models"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

// MetaUpdateCommand updates document frontmatter
type MetaUpdateCommand struct {
	*cmds.CommandDescription
}

// MetaUpdateSettings holds the parameters for the meta update command
type MetaUpdateSettings struct {
	Doc     string `glazed.parameter:"doc"`
	Ticket  string `glazed.parameter:"ticket"`
	DocType string `glazed.parameter:"doc-type"`
	Field   string `glazed.parameter:"field"`
	Value   string `glazed.parameter:"value"`
	Root    string `glazed.parameter:"root"`
}

func NewMetaUpdateCommand() (*MetaUpdateCommand, error) {
	return &MetaUpdateCommand{
		CommandDescription: cmds.NewCommandDescription(
			"update",
			cmds.WithShort("Update document metadata"),
			cmds.WithLong(`Updates frontmatter fields in document files.

Example:
  docmgr meta update --doc ttmp/MEN-1234-slug/index.md --field Status --value review
  docmgr meta update --ticket MEN-1234 --field Status --value active
  docmgr meta update --ticket MEN-1234 --doc-type design-doc --field Topics --value chat,backend
`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"doc",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path to specific document file"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"ticket",
					parameters.ParameterTypeString,
					parameters.WithHelp("Ticket identifier (updates all docs for this ticket)"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"doc-type",
					parameters.ParameterTypeString,
					parameters.WithHelp("Filter by document type (used with --ticket)"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"field",
					parameters.ParameterTypeString,
					parameters.WithHelp("Field name to update (Title, Ticket, Status, Topics, DocType, Intent, Owners, RelatedFiles, ExternalSources, Summary)"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"value",
					parameters.ParameterTypeString,
					parameters.WithHelp("New value for the field (for lists, use comma-separated values)"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"root",
					parameters.ParameterTypeString,
					parameters.WithHelp("Root directory for docs"),
					parameters.WithDefault("ttmp"),
				),
			),
		),
	}, nil
}

func (c *MetaUpdateCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &MetaUpdateSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

    // Apply config root if present
    settings.Root = ResolveRoot(settings.Root)
    // Echo resolved context prior to write
    cfgPath, _ := FindTTMPConfigPath()
    vocabPath, _ := ResolveVocabularyPath()
    absRoot := settings.Root
    if !filepath.IsAbs(absRoot) {
        if cwd, err := os.Getwd(); err == nil {
            absRoot = filepath.Join(cwd, absRoot)
        }
    }
    fmt.Printf("root=%s config=%s vocabulary=%s\n", absRoot, cfgPath, vocabPath)

	var filesToUpdate []string

	if settings.Doc != "" {
		// Update specific file
		filesToUpdate = []string{settings.Doc}
	} else if settings.Ticket != "" {
		// Find all files for ticket
		ticketDir, err := findTicketDirectory(settings.Root, settings.Ticket)
		if err != nil {
			return fmt.Errorf("failed to find ticket directory: %w", err)
		}

		// Find all markdown files in ticket directory
		files, err := findMarkdownFiles(ticketDir, settings.DocType)
		if err != nil {
			return fmt.Errorf("failed to find files: %w", err)
		}
		filesToUpdate = files
	} else {
		return fmt.Errorf("must specify either --doc or --ticket")
	}

	// Update each file
	for _, filePath := range filesToUpdate {
		if err := updateDocumentField(filePath, settings.Field, settings.Value); err != nil {
			row := types.NewRow(
				types.MRP("doc", filePath),
				types.MRP("field", settings.Field),
				types.MRP("status", "error"),
				types.MRP("error", err.Error()),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
			continue
		}

		row := types.NewRow(
			types.MRP("doc", filePath),
			types.MRP("field", settings.Field),
			types.MRP("value", settings.Value),
			types.MRP("status", "updated"),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

// updateDocumentField updates a specific field in a document's frontmatter
func updateDocumentField(filePath string, fieldName string, value string) error {
	// Read file
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = f.Close() }()

	// Parse frontmatter
	var doc models.Document
	rest, err := frontmatter.Parse(f, &doc)
	if err != nil {
		return fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Read rest of content
	restBytes := rest
	content := string(restBytes)

	// Update field based on field name
	// Map field names to struct fields (case-insensitive)
	fieldNameLower := strings.ToLower(fieldName)
	switch fieldNameLower {
	case "title":
		doc.Title = value
	case "ticket":
		doc.Ticket = value
	case "status":
		doc.Status = value
	case "topics":
		// Parse comma-separated values
		topics := []string{}
		for _, topic := range strings.Split(value, ",") {
			topic = strings.TrimSpace(topic)
			if topic != "" {
				topics = append(topics, topic)
			}
		}
		doc.Topics = topics
	case "doctype":
		doc.DocType = value
	case "intent":
		doc.Intent = value
	case "owners":
		// Parse comma-separated values
		owners := []string{}
		for _, owner := range strings.Split(value, ",") {
			owner = strings.TrimSpace(owner)
			if owner != "" {
				owners = append(owners, owner)
			}
		}
		doc.Owners = owners
	case "relatedfiles":
		// Parse comma-separated values into structured entries with empty notes
		var rfs models.RelatedFiles
		for _, file := range strings.Split(value, ",") {
			file = strings.TrimSpace(file)
			if file != "" {
				rfs = append(rfs, models.RelatedFile{Path: file})
			}
		}
		doc.RelatedFiles = rfs
	case "externalsources":
		// Parse comma-separated values
		sources := []string{}
		for _, source := range strings.Split(value, ",") {
			source = strings.TrimSpace(source)
			if source != "" {
				sources = append(sources, source)
			}
		}
		doc.ExternalSources = sources
	case "summary":
		doc.Summary = value
	default:
		return fmt.Errorf("unknown field: %s", fieldName)
	}

	// Update LastUpdated
	doc.LastUpdated = time.Now()

	// Write back to file
	return writeDocumentWithFrontmatter(filePath, &doc, content, true)
}

// findMarkdownFiles finds all markdown files in a directory, optionally filtered by doc type
func findMarkdownFiles(rootDir string, docTypeFilter string) ([]string, error) {
	var files []string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		// If docType filter specified, check frontmatter
		if docTypeFilter != "" {
			doc, err := readDocumentFrontmatter(path)
			if err != nil {
				return nil // Skip files with invalid frontmatter
			}
			if doc.DocType != docTypeFilter {
				return nil
			}
		}

		files = append(files, path)
		return nil
	})

	return files, err
}

var _ cmds.GlazeCommand = &MetaUpdateCommand{}
