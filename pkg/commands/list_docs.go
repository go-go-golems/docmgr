package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

// ListDocsCommand lists individual documents
type ListDocsCommand struct {
	*cmds.CommandDescription
}

// ListDocsSettings holds the parameters for the list docs command
type ListDocsSettings struct {
	Root    string   `glazed.parameter:"root"`
	Ticket  string   `glazed.parameter:"ticket"`
	Status  string   `glazed.parameter:"status"`
	DocType string   `glazed.parameter:"doc-type"`
	Topics  []string `glazed.parameter:"topics"`
}

func NewListDocsCommand() (*ListDocsCommand, error) {
	return &ListDocsCommand{
		CommandDescription: cmds.NewCommandDescription(
			"docs",
			cmds.WithShort("List individual documents"),
			cmds.WithLong(`Lists all individual documents across all workspaces.

Columns:
  ticket,doc_type,title,status,topics,path,last_updated

Examples:
  # Human output
  docmgr list docs
  docmgr list docs --ticket MEN-3475
  docmgr list docs --doc-type design-doc
  docmgr list docs --topics chat,backend

  # Scriptable (paths only)
  docmgr list docs --ticket MEN-3475 --with-glaze-output --select path

  # TSV subset
  docmgr list docs --ticket MEN-3475 --with-glaze-output --output tsv --fields doc_type,title,path
`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"root",
					parameters.ParameterTypeString,
					parameters.WithHelp("Root directory for docs"),
					parameters.WithDefault("ttmp"),
				),
				parameters.NewParameterDefinition(
					"ticket",
					parameters.ParameterTypeString,
					parameters.WithHelp("Filter by ticket identifier"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"status",
					parameters.ParameterTypeString,
					parameters.WithHelp("Filter by status"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"doc-type",
					parameters.ParameterTypeString,
					parameters.WithHelp("Filter by document type"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"topics",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Filter by topics (comma-separated, matches any)"),
					parameters.WithDefault([]string{}),
				),
			),
		),
	}, nil
}

func (c *ListDocsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &ListDocsSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Apply config root if present
	settings.Root = ResolveRoot(settings.Root)

	if _, err := os.Stat(settings.Root); os.IsNotExist(err) {
		return fmt.Errorf("root directory does not exist: %s", settings.Root)
	}

	// Find all markdown files recursively
	err := filepath.Walk(settings.Root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Skip index.md files (those are tickets, use list tickets for those)
		if info.Name() == "index.md" {
			return nil
		}

		doc, err := readDocumentFrontmatter(path)
		if err != nil {
			return nil // Skip files with invalid frontmatter
		}

		// Apply filters
		if settings.Ticket != "" && doc.Ticket != settings.Ticket {
			return nil
		}
		if settings.Status != "" && doc.Status != settings.Status {
			return nil
		}
		if settings.DocType != "" && doc.DocType != settings.DocType {
			return nil
		}
		if len(settings.Topics) > 0 {
			// Check if any of the filter topics match any of the document's topics
			topicMatch := false
			for _, filterTopic := range settings.Topics {
				for _, docTopic := range doc.Topics {
					if strings.EqualFold(strings.TrimSpace(filterTopic), strings.TrimSpace(docTopic)) {
						topicMatch = true
						break
					}
				}
				if topicMatch {
					break
				}
			}
			if !topicMatch {
				return nil
			}
		}

		// Get relative path from root
		relPath, err := filepath.Rel(settings.Root, path)
		if err != nil {
			relPath = path
		}

		row := types.NewRow(
			types.MRP(ColTicket, doc.Ticket),
			types.MRP(ColDocType, doc.DocType),
			types.MRP(ColTitle, doc.Title),
			types.MRP(ColStatus, doc.Status),
			types.MRP(ColTopics, strings.Join(doc.Topics, ", ")),
			types.MRP(ColPath, relPath),
			types.MRP(ColLastUpdated, doc.LastUpdated.Format("2006-01-02 15:04")),
		)

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}

		return nil
	})

	return err
}

var _ cmds.GlazeCommand = &ListDocsCommand{}

// Implement BareCommand for human-friendly output
func (c *ListDocsCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
) error {
	settings := &ListDocsSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Apply config root if present
	settings.Root = ResolveRoot(settings.Root)

	if _, err := os.Stat(settings.Root); os.IsNotExist(err) {
		return fmt.Errorf("root directory does not exist: %s", settings.Root)
	}

	err := filepath.Walk(settings.Root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}
		if info.Name() == "index.md" {
			return nil
		}

		doc, err := readDocumentFrontmatter(path)
		if err != nil {
			return nil
		}

		if settings.Ticket != "" && doc.Ticket != settings.Ticket {
			return nil
		}
		if settings.Status != "" && doc.Status != settings.Status {
			return nil
		}
		if settings.DocType != "" && doc.DocType != settings.DocType {
			return nil
		}
		if len(settings.Topics) > 0 {
			topicMatch := false
			for _, filterTopic := range settings.Topics {
				for _, docTopic := range doc.Topics {
					if strings.EqualFold(strings.TrimSpace(filterTopic), strings.TrimSpace(docTopic)) {
						topicMatch = true
						break
					}
				}
				if topicMatch {
					break
				}
			}
			if !topicMatch {
				return nil
			}
		}

		relPath, err := filepath.Rel(settings.Root, path)
		if err != nil {
			relPath = path
		}

		fmt.Printf("%s %s ‘%s’ status=%s topics=%s updated=%s path=%s\n",
			doc.Ticket,
			doc.DocType,
			doc.Title,
			doc.Status,
			strings.Join(doc.Topics, ", "),
			doc.LastUpdated.Format("2006-01-02 15:04"),
			relPath,
		)
		return nil
	})

	return err
}

var _ cmds.BareCommand = &ListDocsCommand{}
