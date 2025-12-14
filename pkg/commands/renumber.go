package commands

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

// RenumberCommand resequences numeric prefixes within a ticket and updates intra-ticket references.
type RenumberCommand struct {
	*cmds.CommandDescription
}

type RenumberSettings struct {
	Root   string `glazed.parameter:"root"`
	Ticket string `glazed.parameter:"ticket"`
}

type RenumberResult struct {
	Ticket      string
	Renamed     int
	Path        string
	CompletedAt time.Time
}

func NewRenumberCommand() (*RenumberCommand, error) {
	return &RenumberCommand{
		CommandDescription: cmds.NewCommandDescription(
			"renumber",
			cmds.WithShort("Resequence numeric prefixes within a ticket and update references"),
			cmds.WithLong(`Renames .md files in all subdirectories of a ticket to enforce
sequential 2-digit prefixes (01-, 02-, ...; switches to 3 digits past 99) and updates
links within the ticket to reflect new paths.`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"ticket",
					parameters.ParameterTypeString,
					parameters.WithHelp("Ticket identifier"),
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

func (c *RenumberCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &RenumberSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	result, err := c.applyRenumber(ctx, settings)
	if err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("ticket", result.Ticket),
		types.MRP("renamed", result.Renamed),
		types.MRP("status", "completed"),
		types.MRP("path", result.Path),
		types.MRP("time", result.CompletedAt.Format(time.RFC3339)),
	)
	return gp.AddRow(ctx, row)
}

func updateTicketReferences(ticketDir string, renameMap map[string]string) error {
	// Walk all .md files under ticketDir and replace oldRel with newRel
	return filepath.WalkDir(ticketDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		content := string(b)
		updated := content
		for oldRel, newRel := range renameMap {
			// Use forward slashes for markdown links
			updated = strings.ReplaceAll(updated, oldRel, newRel)
		}
		if updated != content {
			if err := os.WriteFile(path, []byte(updated), 0644); err != nil {
				return err
			}
		}
		return nil
	})
}

var _ cmds.GlazeCommand = &RenumberCommand{}
var _ cmds.BareCommand = &RenumberCommand{}

func (c *RenumberCommand) applyRenumber(ctx context.Context, settings *RenumberSettings) (*RenumberResult, error) {
	if ctx == nil {
		return nil, fmt.Errorf("nil context")
	}
	settings.Root = workspace.ResolveRoot(settings.Root)

	ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{RootOverride: settings.Root})
	if err != nil {
		return nil, fmt.Errorf("failed to discover workspace: %w", err)
	}
	settings.Root = ws.Context().Root
	if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{IncludeBody: false}); err != nil {
		return nil, fmt.Errorf("failed to initialize workspace index: %w", err)
	}

	ticketDir, err := resolveTicketDirViaWorkspace(ctx, ws, settings.Ticket)
	if err != nil {
		return nil, fmt.Errorf("failed to find ticket directory: %w", err)
	}

	entries, err := os.ReadDir(ticketDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read ticket dir: %w", err)
	}

	renameMap := map[string]string{}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, "_") || name == ".meta" {
			continue
		}

		subdir := filepath.Join(ticketDir, name)
		var files []string
		err = filepath.WalkDir(subdir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				return nil
			}
			if strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("walk failed: %w", err)
		}

		sort.Slice(files, func(i, j int) bool {
			bi, _, _ := stripNumericPrefix(filepath.Base(files[i]))
			bj, _, _ := stripNumericPrefix(filepath.Base(files[j]))
			return bi < bj
		})

		next := 1
		for _, oldPath := range files {
			base := filepath.Base(oldPath)
			stripped, _, _ := stripNumericPrefix(base)
			width := 2
			if next >= 100 {
				width = 3
			}
			newBase := fmt.Sprintf("%0*d-%s", width, next, stripped)
			next++
			if base == newBase {
				continue
			}
			newPath := filepath.Join(filepath.Dir(oldPath), newBase)
			if err := os.Rename(oldPath, newPath); err != nil {
				return nil, fmt.Errorf("failed to rename %s -> %s: %w", oldPath, newPath, err)
			}
			oldRel, _ := filepath.Rel(ticketDir, oldPath)
			newRel, _ := filepath.Rel(ticketDir, newPath)
			renameMap[filepath.ToSlash(oldRel)] = filepath.ToSlash(newRel)
		}
	}

	if len(renameMap) > 0 {
		if err := updateTicketReferences(ticketDir, renameMap); err != nil {
			return nil, fmt.Errorf("failed to update references: %w", err)
		}
	}

	return &RenumberResult{
		Ticket:      settings.Ticket,
		Renamed:     len(renameMap),
		Path:        ticketDir,
		CompletedAt: time.Now(),
	}, nil
}

func (c *RenumberCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
) error {
	settings := &RenumberSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	result, err := c.applyRenumber(ctx, settings)
	if err != nil {
		return err
	}

	fmt.Printf("Ticket %s renumbered.\n", result.Ticket)
	fmt.Printf("- Files renamed: %d\n", result.Renamed)
	fmt.Printf("- Path: %s\n", result.Path)
	fmt.Printf("- Completed: %s\n", result.CompletedAt.Format(time.RFC3339))

	return nil
}
