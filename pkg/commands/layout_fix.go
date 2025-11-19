package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/docmgr/internal/documents"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

// LayoutFixCommand moves documents into subdirectories named after their doc-type
type LayoutFixCommand struct {
	*cmds.CommandDescription
}

type LayoutFixSettings struct {
	Root   string `glazed.parameter:"root"`
	Ticket string `glazed.parameter:"ticket"`
	DryRun bool   `glazed.parameter:"dry-run"`
}

func NewLayoutFixCommand() (*LayoutFixCommand, error) {
	return &LayoutFixCommand{
		CommandDescription: cmds.NewCommandDescription(
			"layout-fix",
			cmds.WithShort("Move docs into <doc-type>/ subdirectories and update links"),
			cmds.WithLong(`Scans ticket workspaces and moves markdown documents into a subdirectory
named exactly after their DocType frontmatter (e.g., design-doc/, reference/, playbook/, or custom).
Skips root-level control files (index.md, tasks.md, changelog.md, README.md).`),
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
					parameters.WithHelp("Limit to a specific ticket"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"dry-run",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Show planned moves without changing files"),
					parameters.WithDefault(false),
				),
			),
		),
	}, nil
}

func (c *LayoutFixCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &LayoutFixSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	settings.Root = workspace.ResolveRoot(settings.Root)
	// Collect tickets to process
	var ticketDirs []string
	if settings.Ticket != "" {
		td, err := findTicketDirectory(settings.Root, settings.Ticket)
		if err != nil {
			return fmt.Errorf("failed to find ticket directory: %w", err)
		}
		ticketDirs = append(ticketDirs, td)
	} else {
		entries, err := os.ReadDir(settings.Root)
		if err != nil {
			return fmt.Errorf("failed to read root: %w", err)
		}
		for _, e := range entries {
			if e.IsDir() {
				// consider directories with an index.md
				idx := filepath.Join(settings.Root, e.Name(), "index.md")
				if _, err := os.Stat(idx); err == nil {
					ticketDirs = append(ticketDirs, filepath.Join(settings.Root, e.Name()))
				}
			}
		}
	}

	for _, ticketDir := range ticketDirs {
		// Walk ticketDir for markdown files
		renameMap := map[string]string{}
		err := filepath.WalkDir(ticketDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				// Skip scaffolding dirs
				name := d.Name()
				if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") || name == ".meta" || name == "scripts" || name == "sources" || name == "archive" {
					return nil
				}
				return nil
			}
			if !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
				return nil
			}
			// Skip root-level control files
			dir := filepath.Dir(path)
			if filepath.Clean(dir) == filepath.Clean(ticketDir) {
				bn := d.Name()
				if bn == "index.md" || bn == "README.md" || bn == "tasks.md" || bn == "changelog.md" {
					return nil
				}
			}

			// Determine current first-level directory relative to ticketDir
			rel, _ := filepath.Rel(ticketDir, path)
			parts := strings.Split(rel, string(os.PathSeparator))
			if len(parts) < 1 {
				return nil
			}

			// Read frontmatter to get DocType
			doc, _, err := documents.ReadDocumentWithFrontmatter(path)
			if err != nil || doc.DocType == "" {
				return nil
			}

			expected := doc.DocType
			currentTop := parts[0]
			if currentTop == expected { // already in the right place
				return nil
			}

			// Plan move: keep basename, move under expected/
			newRel := filepath.Join(expected, filepath.Base(path))
			newAbs := filepath.Join(ticketDir, newRel)
			if settings.DryRun {
				row := types.NewRow(
					types.MRP("ticket", filepath.Base(ticketDir)),
					types.MRP("from", filepath.ToSlash(rel)),
					types.MRP("to", filepath.ToSlash(newRel)),
					types.MRP("status", "would-move"),
				)
				if err := gp.AddRow(ctx, row); err != nil {
					return fmt.Errorf("failed to add layout-fix dry-run row for %s: %w", rel, err)
				}
				return nil
			}

			if err := os.MkdirAll(filepath.Dir(newAbs), 0755); err != nil {
				return fmt.Errorf("failed to ensure target directory for %s: %w", newAbs, err)
			}
			if err := os.Rename(path, newAbs); err != nil {
				return fmt.Errorf("rename %s -> %s failed: %w", path, newAbs, err)
			}
			oldRel := filepath.ToSlash(rel)
			renameMap[oldRel] = filepath.ToSlash(newRel)

			row := types.NewRow(
				types.MRP("ticket", filepath.Base(ticketDir)),
				types.MRP("from", oldRel),
				types.MRP("to", filepath.ToSlash(newRel)),
				types.MRP("status", "moved"),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return fmt.Errorf("failed to add layout-fix row for %s: %w", oldRel, err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to walk ticket directory %s: %w", ticketDir, err)
		}
		if !settings.DryRun && len(renameMap) > 0 {
			if err := updateTicketReferences(ticketDir, renameMap); err != nil {
				return fmt.Errorf("failed to update references in %s: %w", ticketDir, err)
			}
		}
	}

	return nil
}

var _ cmds.GlazeCommand = &LayoutFixCommand{}
