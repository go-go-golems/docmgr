package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

// StatusCommand prints a summary status of the docs root
type StatusCommand struct {
	*cmds.CommandDescription
}

type StatusSettings struct {
	Root           string `glazed.parameter:"root"`
	Ticket         string `glazed.parameter:"ticket"`
	StaleAfterDays int    `glazed.parameter:"stale-after"`
	SummaryOnly    bool   `glazed.parameter:"summary-only"`
}

func NewStatusCommand() (*StatusCommand, error) {
	return &StatusCommand{
		CommandDescription: cmds.NewCommandDescription(
			"status",
			cmds.WithShort("Show overall status of the documentation root"),
			cmds.WithLong(`Prints a summary of ticket workspaces and documents, including staleness.

Examples:
  docmgr status
  docmgr status --stale-after 30
  docmgr status --ticket MEN-4242
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
					parameters.WithHelp("Limit to a specific ticket"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"stale-after",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Days after which a ticket is considered stale (default 30)"),
					parameters.WithDefault(30),
				),
				parameters.NewParameterDefinition(
					"summary-only",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Print only the summary row, without per-ticket rows"),
					parameters.WithDefault(false),
				),
			),
		),
	}, nil
}

func (c *StatusCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &StatusSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Apply config root if present
	settings.Root = ResolveRoot(settings.Root)

	if _, err := os.Stat(settings.Root); os.IsNotExist(err) {
		return fmt.Errorf("root directory does not exist: %s", settings.Root)
	}

	ticketsTotal := 0
	ticketsStale := 0
	docsTotal := 0
	designDocs := 0
	referenceDocs := 0
	playbooks := 0

	workspaces, err := collectTicketWorkspaces(settings.Root, nil)
	if err != nil {
		return fmt.Errorf("failed to discover ticket workspaces: %w", err)
	}

	for _, ws := range workspaces {
		doc := ws.Doc
		if doc == nil {
			continue
		}
		if settings.Ticket != "" && doc.Ticket != settings.Ticket {
			continue
		}

		ticketsTotal++
		ticketPath := ws.Path

		stale := false
		if !doc.LastUpdated.IsZero() {
			days := time.Since(doc.LastUpdated).Hours() / 24
			if int(days) > settings.StaleAfterDays {
				stale = true
			}
		}
		if stale {
			ticketsStale++
		}

		ticketDocs := 0
		dd := 0
		rd := 0
		pb := 0
		_ = filepath.Walk(ticketPath, func(path string, info os.FileInfo, err error) error {
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
			d, err := readDocumentFrontmatter(path)
			if err != nil {
				return nil
			}
			ticketDocs++
			switch d.DocType {
			case "design-doc":
				dd++
			case "reference":
				rd++
			case "playbook":
				pb++
			}
			return nil
		})

		docsTotal += ticketDocs
		designDocs += dd
		referenceDocs += rd
		playbooks += pb

		if !settings.SummaryOnly {
			row := types.NewRow(
				types.MRP("ticket", doc.Ticket),
				types.MRP("title", doc.Title),
				types.MRP("status", doc.Status),
				types.MRP("last_updated", doc.LastUpdated.Format("2006-01-02")),
				types.MRP("stale", stale),
				types.MRP("docs", ticketDocs),
				types.MRP("design_docs", dd),
				types.MRP("reference_docs", rd),
				types.MRP("playbooks", pb),
				types.MRP("path", ticketPath),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}
	}

	// Resolve config and vocabulary paths for summary
	cfgPath, _ := FindTTMPConfigPath()
	vocabPath, _ := ResolveVocabularyPath()

	// Emit warnings
	cwd, _ := os.Getwd()
	fallbackCandidate := filepath.Join(cwd, "ttmp")
	if cfgPath == "" {
		if _, err := FindGitRoot(); err != nil {
			// No config and no git; if using CWD fallback, warn
			if filepath.Clean(settings.Root) == filepath.Clean(fallbackCandidate) {
				_ = gp.AddRow(ctx, types.NewRow(
					types.MRP("level", "warning"),
					types.MRP("message", "No .ttmp.yaml found; using <cwd>/ttmp fallback"),
					types.MRP("root", settings.Root),
				))
			}
		}
	}
	if roots, err := DetectMultipleTTMPRoots(); err == nil && len(roots) > 1 {
		_ = gp.AddRow(ctx, types.NewRow(
			types.MRP("level", "warning"),
			types.MRP("message", fmt.Sprintf("Multiple ttmp/ roots detected: %s", strings.Join(roots, ", "))),
		))
	}

	// Summary row
	sum := types.NewRow(
		types.MRP("root", settings.Root),
		types.MRP("config_path", cfgPath),
		types.MRP("vocabulary_path", vocabPath),
		types.MRP("tickets_total", ticketsTotal),
		types.MRP("tickets_stale", ticketsStale),
		types.MRP("docs_total", docsTotal),
		types.MRP("design_docs", designDocs),
		types.MRP("reference_docs", referenceDocs),
		types.MRP("playbooks", playbooks),
		types.MRP("stale_after_days", settings.StaleAfterDays),
		types.MRP("status", "ok"),
	)
	return gp.AddRow(ctx, sum)
}

var _ cmds.GlazeCommand = &StatusCommand{}

// Implement BareCommand for human-friendly output
func (c *StatusCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,

) error {
	settings := &StatusSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	settings.Root = ResolveRoot(settings.Root)
	if _, err := os.Stat(settings.Root); os.IsNotExist(err) {
		return fmt.Errorf("root directory does not exist: %s", settings.Root)
	}

	ticketsTotal := 0
	ticketsStale := 0
	docsTotal := 0
	designDocs := 0
	referenceDocs := 0
	playbooks := 0

	workspaces, err := collectTicketWorkspaces(settings.Root, nil)
	if err != nil {
		return fmt.Errorf("failed to discover ticket workspaces: %w", err)
	}

	for _, ws := range workspaces {
		doc := ws.Doc
		if doc == nil {
			continue
		}
		if settings.Ticket != "" && doc.Ticket != settings.Ticket {
			continue
		}

		ticketsTotal++
		ticketPath := ws.Path

		stale := false
		if !doc.LastUpdated.IsZero() {
			days := time.Since(doc.LastUpdated).Hours() / 24
			if int(days) > settings.StaleAfterDays {
				stale = true
			}
		}
		if stale {
			ticketsStale++
		}

		ticketDocs := 0
		dd, rd, pb := 0, 0, 0
		_ = filepath.Walk(ticketPath, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
				return nil
			}
			if info.Name() == "index.md" {
				return nil
			}
			d, err := readDocumentFrontmatter(path)
			if err != nil {
				return nil
			}
			ticketDocs++
			switch d.DocType {
			case "design-doc":
				dd++
			case "reference":
				rd++
			case "playbook":
				pb++
			}
			return nil
		})

		docsTotal += ticketDocs
		designDocs += dd
		referenceDocs += rd
		playbooks += pb

		if !settings.SummaryOnly {
			fmt.Printf("%s ‘%s’ status=%s stale=%t docs=%d path=%s\n",
				doc.Ticket, doc.Title, doc.Status, stale, ticketDocs, ticketPath,
			)
		}
	}

	cfgPath, _ := FindTTMPConfigPath()
	vocabPath, _ := ResolveVocabularyPath()
	fmt.Printf(
		"root=%s config=%s vocabulary=%s tickets=%d stale=%d docs=%d (design %d / reference %d / playbooks %d) stale-after=%d\n",
		settings.Root, cfgPath, vocabPath, ticketsTotal, ticketsStale, docsTotal, designDocs, referenceDocs, playbooks, settings.StaleAfterDays,
	)
	return nil
}

var _ cmds.BareCommand = &StatusCommand{}
