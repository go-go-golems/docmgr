package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-go-golems/docmgr/internal/documents"
	"github.com/go-go-golems/docmgr/internal/workspace"
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

type MetaUpdateContext struct {
	Root           string
	ConfigPath     string
	VocabularyPath string
}

type MetaUpdateRow struct {
	Doc    string
	Field  string
	Value  string
	Status string
	Error  string
}

type MetaUpdateExecutionResult struct {
	Context MetaUpdateContext
	Updates []MetaUpdateRow
}

func NewMetaUpdateCommand() (*MetaUpdateCommand, error) {
	return &MetaUpdateCommand{
		CommandDescription: cmds.NewCommandDescription(
			"update",
			cmds.WithShort("Update document metadata"),
			cmds.WithLong(`Updates frontmatter fields in document files.

Behavior:
  • If --doc is provided: update that file.
  • If --ticket is provided without --doc-type: update the ticket's index.md only (default).
  • If --ticket and --doc-type are provided: update all docs of that type under the ticket.

Examples:
  # Update a specific file
  docmgr meta update --doc ttmp/YYYY/MM/DD/MEN-1234-slug/index.md --field Status --value review

  # Update ticket index.md (default when only --ticket is specified)
  docmgr meta update --ticket MEN-1234 --field Status --value active

  # Update all design-docs under a ticket
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

	result, err := c.applyMetaUpdate(ctx, settings)
	if err != nil {
		return err
	}

	for _, update := range result.Updates {
		if update.Status == "error" {
			row := types.NewRow(
				types.MRP("doc", update.Doc),
				types.MRP("field", update.Field),
				types.MRP("status", update.Status),
				types.MRP("error", update.Error),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return fmt.Errorf("failed to report meta update error for %s: %w", update.Doc, err)
			}
			continue
		}

		row := types.NewRow(
			types.MRP("doc", update.Doc),
			types.MRP("field", update.Field),
			types.MRP("value", update.Value),
			types.MRP("status", update.Status),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return fmt.Errorf("failed to add meta update row for %s: %w", update.Doc, err)
		}
	}

	return nil
}

func (c *MetaUpdateCommand) applyMetaUpdate(ctx context.Context, settings *MetaUpdateSettings) (*MetaUpdateExecutionResult, error) {
	if ctx == nil {
		return nil, fmt.Errorf("nil context")
	}
	settings.Root = workspace.ResolveRoot(settings.Root)
	cfgPath, _ := workspace.FindTTMPConfigPath()
	vocabPath, _ := workspace.ResolveVocabularyPath()
	absRoot := settings.Root
	if !filepath.IsAbs(absRoot) {
		if cwd, err := os.Getwd(); err == nil {
			absRoot = filepath.Join(cwd, absRoot)
		}
	}

	muCtx := MetaUpdateContext{
		Root:           absRoot,
		ConfigPath:     cfgPath,
		VocabularyPath: vocabPath,
	}

	var filesToUpdate []string

	if settings.Doc != "" {
		filesToUpdate = []string{settings.Doc}
	} else if settings.Ticket != "" {
		// Workspace+QueryDocs-backed ticket discovery and document enumeration.
		ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{RootOverride: settings.Root})
		if err != nil {
			return nil, fmt.Errorf("failed to discover workspace: %w", err)
		}
		settings.Root = ws.Context().Root
		muCtx.Root = ws.Context().Root
		if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{IncludeBody: false}); err != nil {
			return nil, fmt.Errorf("failed to initialize workspace index: %w", err)
		}

		if settings.DocType == "" {
			// Default: update ticket index.md only.
			res, err := ws.QueryDocs(ctx, workspace.DocQuery{
				Scope:   workspace.Scope{Kind: workspace.ScopeTicket, TicketID: strings.TrimSpace(settings.Ticket)},
				Filters: workspace.DocFilters{DocType: "index"},
				Options: workspace.DocQueryOptions{
					IncludeErrors:       false,
					IncludeArchivedPath: true,
					IncludeScriptsPath:  true,
					IncludeControlDocs:  true,
					OrderBy:             workspace.OrderByPath,
				},
			})
			if err != nil {
				return nil, fmt.Errorf("failed to find ticket index doc: %w", err)
			}
			if len(res.Docs) != 1 || res.Docs[0].Path == "" {
				return nil, fmt.Errorf("ticket not found or ambiguous: %s", strings.TrimSpace(settings.Ticket))
			}
			filesToUpdate = []string{filepath.FromSlash(res.Docs[0].Path)}
		} else {
			// Update all docs of the requested doc type within the ticket.
			res, err := ws.QueryDocs(ctx, workspace.DocQuery{
				Scope: workspace.Scope{Kind: workspace.ScopeTicket, TicketID: strings.TrimSpace(settings.Ticket)},
				Filters: workspace.DocFilters{
					DocType: strings.TrimSpace(settings.DocType),
				},
				Options: workspace.DocQueryOptions{
					IncludeErrors:       false,
					IncludeArchivedPath: true,
					IncludeScriptsPath:  true,
					IncludeControlDocs:  true,
					OrderBy:             workspace.OrderByPath,
				},
			})
			if err != nil {
				return nil, fmt.Errorf("failed to query docs: %w", err)
			}
			for _, h := range res.Docs {
				if strings.TrimSpace(h.Path) == "" {
					continue
				}
				filesToUpdate = append(filesToUpdate, filepath.FromSlash(h.Path))
			}
		}
	} else {
		return nil, fmt.Errorf("must specify either --doc or --ticket")
	}

	updates := make([]MetaUpdateRow, 0, len(filesToUpdate))
	for _, filePath := range filesToUpdate {
		if err := updateDocumentField(filePath, settings.Field, settings.Value); err != nil {
			updates = append(updates, MetaUpdateRow{
				Doc:    filePath,
				Field:  settings.Field,
				Status: "error",
				Error:  err.Error(),
			})
			continue
		}

		updates = append(updates, MetaUpdateRow{
			Doc:    filePath,
			Field:  settings.Field,
			Value:  settings.Value,
			Status: "updated",
		})
	}

	return &MetaUpdateExecutionResult{
		Context: muCtx,
		Updates: updates,
	}, nil
}

func (c *MetaUpdateCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
) error {
	settings := &MetaUpdateSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	result, err := c.applyMetaUpdate(ctx, settings)
	if err != nil {
		return err
	}

	fmt.Printf("Docs root: `%s`\n", result.Context.Root)
	if result.Context.ConfigPath != "" {
		fmt.Printf("Config: `%s`\n", result.Context.ConfigPath)
	}
	if result.Context.VocabularyPath != "" {
		fmt.Printf("Vocabulary: `%s`\n", result.Context.VocabularyPath)
	}

	fmt.Printf("\n## Metadata Updates\n\n")
	for _, update := range result.Updates {
		if update.Status == "error" {
			fmt.Printf("- `%s`: error updating %s (%s)\n", update.Doc, update.Field, update.Error)
			continue
		}
		fmt.Printf("- `%s`: %s → %s\n", update.Doc, update.Field, update.Value)
	}

	return nil
}

// updateDocumentField updates a specific field in a document's frontmatter
func updateDocumentField(filePath string, fieldName string, value string) error {
	doc, content, err := documents.ReadDocumentWithFrontmatter(filePath)
	if err != nil {
		return err
	}

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
	return documents.WriteDocumentWithFrontmatter(filePath, doc, content, true)
}

var _ cmds.GlazeCommand = &MetaUpdateCommand{}
var _ cmds.BareCommand = &MetaUpdateCommand{}
