package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-go-golems/docmgr/internal/templates"
	"github.com/go-go-golems/docmgr/internal/workspace"
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
	Root               string `glazed.parameter:"root"`
	Ticket             string `glazed.parameter:"ticket"`
	StaleAfterDays     int    `glazed.parameter:"stale-after"`
	SummaryOnly        bool   `glazed.parameter:"summary-only"`
	PrintTemplateSchema bool   `glazed.parameter:"print-template-schema"`
	SchemaFormat        string `glazed.parameter:"schema-format"`
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
				parameters.NewParameterDefinition(
					"print-template-schema",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Print template schema after output (human mode only)"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"schema-format",
					parameters.ParameterTypeString,
					parameters.WithHelp("Template schema output format: json|yaml"),
					parameters.WithDefault("json"),
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
	settings.Root = workspace.ResolveRoot(settings.Root)

	// If only printing template schema, skip all other processing and output
	if settings.PrintTemplateSchema {
		type TicketInfo struct {
			Ticket      string
			Title       string
			Status      string
			Stale       bool
			Docs        int
			DesignDocs  int
			ReferenceDocs int
			Playbooks   int
			Path        string
			LastUpdated string
		}
		templateData := map[string]interface{}{
			"TicketsTotal":  0,
			"TicketsStale":  0,
			"DocsTotal":     0,
			"DesignDocs":    0,
			"ReferenceDocs": 0,
			"Playbooks":     0,
			"StaleAfterDays": 30,
			"Root":          "",
			"ConfigPath":    "",
			"VocabularyPath": "",
			"Tickets": []TicketInfo{
				{
					Ticket:       "",
					Title:        "",
					Status:       "",
					Stale:        false,
					Docs:         0,
					DesignDocs:   0,
					ReferenceDocs: 0,
					Playbooks:    0,
					Path:         "",
					LastUpdated:  "",
				},
			},
		}
		_ = templates.PrintSchema(os.Stdout, templateData, settings.SchemaFormat)
		return nil
	}

	if _, err := os.Stat(settings.Root); os.IsNotExist(err) {
		return fmt.Errorf("root directory does not exist: %s", settings.Root)
	}

	ticketsTotal := 0
	ticketsStale := 0
	docsTotal := 0
	designDocs := 0
	referenceDocs := 0
	playbooks := 0

	workspaces, err := workspace.CollectTicketWorkspaces(settings.Root, nil)
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
				return fmt.Errorf("failed to add status row for %s: %w", doc.Ticket, err)
			}
		}
	}

	// Resolve config and vocabulary paths for summary
	cfgPath, _ := workspace.FindTTMPConfigPath()
	vocabPath, _ := workspace.ResolveVocabularyPath()

	// Emit warnings
	cwd, _ := os.Getwd()
	fallbackCandidate := filepath.Join(cwd, "ttmp")
	if cfgPath == "" {
		if _, err := workspace.FindGitRoot(); err != nil {
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
	if roots, err := workspace.DetectMultipleTTMPRoots(); err == nil && len(roots) > 1 {
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
	if err := gp.AddRow(ctx, sum); err != nil {
		return fmt.Errorf("failed to add status summary row: %w", err)
	}
	return nil
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

	settings.Root = workspace.ResolveRoot(settings.Root)

	// If only printing template schema, skip all other processing and output
	if settings.PrintTemplateSchema {
		type TicketInfo struct {
			Ticket      string
			Title       string
			Status      string
			Stale       bool
			Docs        int
			DesignDocs  int
			ReferenceDocs int
			Playbooks   int
			Path        string
			LastUpdated string
		}
		templateData := map[string]interface{}{
			"TicketsTotal":  0,
			"TicketsStale":  0,
			"DocsTotal":     0,
			"DesignDocs":    0,
			"ReferenceDocs": 0,
			"Playbooks":     0,
			"StaleAfterDays": 30,
			"Root":          "",
			"ConfigPath":    "",
			"VocabularyPath": "",
			"Tickets": []TicketInfo{
				{
					Ticket:       "",
					Title:        "",
					Status:       "",
					Stale:        false,
					Docs:         0,
					DesignDocs:   0,
					ReferenceDocs: 0,
					Playbooks:    0,
					Path:         "",
					LastUpdated:  "",
				},
			},
		}
		_ = templates.PrintSchema(os.Stdout, templateData, settings.SchemaFormat)
		return nil
	}

	if _, err := os.Stat(settings.Root); os.IsNotExist(err) {
		return fmt.Errorf("root directory does not exist: %s", settings.Root)
	}

	ticketsTotal := 0
	ticketsStale := 0
	docsTotal := 0
	designDocs := 0
	referenceDocs := 0
	playbooks := 0

	workspaces, err := workspace.CollectTicketWorkspaces(settings.Root, nil)
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

	cfgPath, _ := workspace.FindTTMPConfigPath()
	vocabPath, _ := workspace.ResolveVocabularyPath()
	fmt.Printf(
		"root=%s config=%s vocabulary=%s tickets=%d stale=%d docs=%d (design %d / reference %d / playbooks %d) stale-after=%d\n",
		settings.Root, cfgPath, vocabPath, ticketsTotal, ticketsStale, docsTotal, designDocs, referenceDocs, playbooks, settings.StaleAfterDays,
	)

	// Render postfix template if it exists
	// Build template data struct
	type TicketInfo struct {
		Ticket        string
		Title         string
		Status        string
		Stale         bool
		Docs          int
		DesignDocs    int
		ReferenceDocs int
		Playbooks     int
		Path          string
		LastUpdated   string
	}

	ticketInfos := make([]TicketInfo, 0)
	for _, ws := range workspaces {
		doc := ws.Doc
		if doc == nil {
			continue
		}
		if settings.Ticket != "" && doc.Ticket != settings.Ticket {
			continue
		}

		ticketPath := ws.Path
		stale := false
		if !doc.LastUpdated.IsZero() {
			days := time.Since(doc.LastUpdated).Hours() / 24
			if int(days) > settings.StaleAfterDays {
				stale = true
			}
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

		lastUpdated := ""
		if !doc.LastUpdated.IsZero() {
			lastUpdated = doc.LastUpdated.Format("2006-01-02 15:04")
		}

		ticketInfos = append(ticketInfos, TicketInfo{
			Ticket:        doc.Ticket,
			Title:         doc.Title,
			Status:        doc.Status,
			Stale:         stale,
			Docs:          ticketDocs,
			DesignDocs:    dd,
			ReferenceDocs: rd,
			Playbooks:     pb,
			Path:          ticketPath,
			LastUpdated:   lastUpdated,
		})
	}

	templateData := map[string]interface{}{
		"TicketsTotal":   ticketsTotal,
		"TicketsStale":   ticketsStale,
		"DocsTotal":      docsTotal,
		"DesignDocs":     designDocs,
		"ReferenceDocs":  referenceDocs,
		"Playbooks":      playbooks,
		"StaleAfterDays": settings.StaleAfterDays,
		"Root":           settings.Root,
		"ConfigPath":     cfgPath,
		"VocabularyPath": vocabPath,
		"Tickets":        ticketInfos,
	}

	// Try verb path: ["status"]
	verbCandidates := [][]string{
		{"status"},
	}
	settingsMap := map[string]interface{}{
		"root":           settings.Root,
		"ticket":         settings.Ticket,
		"staleAfterDays": settings.StaleAfterDays,
		"summaryOnly":    settings.SummaryOnly,
	}
	absRoot := settings.Root
	if abs, err := filepath.Abs(settings.Root); err == nil {
		absRoot = abs
	}
	_ = templates.RenderVerbTemplate(verbCandidates, absRoot, settingsMap, templateData)

	return nil
}

var _ cmds.BareCommand = &StatusCommand{}
