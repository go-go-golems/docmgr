package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
	Root                string `glazed.parameter:"root"`
	Ticket              string `glazed.parameter:"ticket"`
	StaleAfterDays      int    `glazed.parameter:"stale-after"`
	SummaryOnly         bool   `glazed.parameter:"summary-only"`
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
  docmgr status --with-glaze-output --output json
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
		templateData := map[string]interface{}{
			"TicketsTotal":   0,
			"TicketsStale":   0,
			"DocsTotal":      0,
			"DesignDocs":     0,
			"ReferenceDocs":  0,
			"Playbooks":      0,
			"StaleAfterDays": 30,
			"Root":           "",
			"ConfigPath":     "",
			"VocabularyPath": "",
			"Tickets": []TicketInfo{
				{
					Ticket:        "",
					Title:         "",
					Status:        "",
					Stale:         false,
					Docs:          0,
					DesignDocs:    0,
					ReferenceDocs: 0,
					Playbooks:     0,
					Path:          "",
					LastUpdated:   "",
				},
			},
		}
		_ = templates.PrintSchema(os.Stdout, templateData, settings.SchemaFormat)
		return nil
	}

	if _, err := os.Stat(settings.Root); os.IsNotExist(err) {
		return fmt.Errorf("root directory does not exist: %s", settings.Root)
	}

	tickets, summary, err := computeStatusTickets(ctx, settings.Root, settings.Ticket, settings.StaleAfterDays)
	if err != nil {
		return err
	}

	if !settings.SummaryOnly {
		for _, t := range tickets {
			row := types.NewRow(
				types.MRP("ticket", t.Ticket),
				types.MRP("title", t.Title),
				types.MRP("status", t.Status),
				types.MRP("last_updated", formatDate(t.LastUpdated)),
				types.MRP("stale", t.Stale),
				types.MRP("docs", t.Docs),
				types.MRP("design_docs", t.DesignDocs),
				types.MRP("reference_docs", t.ReferenceDocs),
				types.MRP("playbooks", t.Playbooks),
				types.MRP("path", t.Path),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return fmt.Errorf("failed to add status row for %s: %w", t.Ticket, err)
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
		types.MRP("tickets_total", summary.TicketsTotal),
		types.MRP("tickets_stale", summary.TicketsStale),
		types.MRP("docs_total", summary.DocsTotal),
		types.MRP("design_docs", summary.DesignDocs),
		types.MRP("reference_docs", summary.ReferenceDocs),
		types.MRP("playbooks", summary.Playbooks),
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
		templateData := map[string]interface{}{
			"TicketsTotal":   0,
			"TicketsStale":   0,
			"DocsTotal":      0,
			"DesignDocs":     0,
			"ReferenceDocs":  0,
			"Playbooks":      0,
			"StaleAfterDays": 30,
			"Root":           "",
			"ConfigPath":     "",
			"VocabularyPath": "",
			"Tickets": []TicketInfo{
				{
					Ticket:        "",
					Title:         "",
					Status:        "",
					Stale:         false,
					Docs:          0,
					DesignDocs:    0,
					ReferenceDocs: 0,
					Playbooks:     0,
					Path:          "",
					LastUpdated:   "",
				},
			},
		}
		_ = templates.PrintSchema(os.Stdout, templateData, settings.SchemaFormat)
		return nil
	}

	if _, err := os.Stat(settings.Root); os.IsNotExist(err) {
		return fmt.Errorf("root directory does not exist: %s", settings.Root)
	}

	tickets, summary, err := computeStatusTickets(ctx, settings.Root, settings.Ticket, settings.StaleAfterDays)
	if err != nil {
		return err
	}

	if !settings.SummaryOnly {
		for _, t := range tickets {
			fmt.Printf("%s ‘%s’ status=%s stale=%t docs=%d path=%s\n",
				t.Ticket, t.Title, t.Status, t.Stale, t.Docs, t.Path,
			)
		}
	}

	cfgPath, _ := workspace.FindTTMPConfigPath()
	vocabPath, _ := workspace.ResolveVocabularyPath()
	fmt.Printf(
		"root=%s config=%s vocabulary=%s tickets=%d stale=%d docs=%d (design %d / reference %d / playbooks %d) stale-after=%d\n",
		settings.Root, cfgPath, vocabPath, summary.TicketsTotal, summary.TicketsStale, summary.DocsTotal, summary.DesignDocs, summary.ReferenceDocs, summary.Playbooks, settings.StaleAfterDays,
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

	ticketInfos := make([]TicketInfo, 0, len(tickets))
	for _, t := range tickets {
		lastUpdated := ""
		if !t.LastUpdated.IsZero() {
			lastUpdated = t.LastUpdated.Format("2006-01-02 15:04")
		}

		ticketInfos = append(ticketInfos, TicketInfo{
			Ticket:        t.Ticket,
			Title:         t.Title,
			Status:        t.Status,
			Stale:         t.Stale,
			Docs:          t.Docs,
			DesignDocs:    t.DesignDocs,
			ReferenceDocs: t.ReferenceDocs,
			Playbooks:     t.Playbooks,
			Path:          t.Path,
			LastUpdated:   lastUpdated,
		})
	}

	templateData := map[string]interface{}{
		"TicketsTotal":   summary.TicketsTotal,
		"TicketsStale":   summary.TicketsStale,
		"DocsTotal":      summary.DocsTotal,
		"DesignDocs":     summary.DesignDocs,
		"ReferenceDocs":  summary.ReferenceDocs,
		"Playbooks":      summary.Playbooks,
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

type statusTicket struct {
	Ticket        string
	Title         string
	Status        string
	Stale         bool
	Docs          int
	DesignDocs    int
	ReferenceDocs int
	Playbooks     int
	Path          string
	LastUpdated   time.Time
}

type statusSummary struct {
	TicketsTotal  int
	TicketsStale  int
	DocsTotal     int
	DesignDocs    int
	ReferenceDocs int
	Playbooks     int
}

func computeStatusTickets(ctx context.Context, root string, ticketFilter string, staleAfterDays int) ([]statusTicket, statusSummary, error) {
	ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{RootOverride: root})
	if err != nil {
		return nil, statusSummary{}, fmt.Errorf("failed to discover workspace: %w", err)
	}
	if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{}); err != nil {
		return nil, statusSummary{}, fmt.Errorf("failed to initialize workspace index: %w", err)
	}

	res, err := ws.QueryDocs(ctx, workspace.DocQuery{
		Scope: workspace.Scope{Kind: workspace.ScopeRepo},
		Options: workspace.DocQueryOptions{
			OrderBy:             workspace.OrderByPath,
			IncludeArchivedPath: true,
			IncludeScriptsPath:  true,
			IncludeControlDocs:  true,
		},
	})
	if err != nil {
		return nil, statusSummary{}, fmt.Errorf("failed to query docs: %w", err)
	}

	type agg struct {
		ticketID     string
		title        string
		status       string
		lastUpdated  time.Time
		ticketDir    string
		hasIndex     bool
		docs         int
		designDocs   int
		referenceDoc int
		playbooks    int
	}

	aggs := map[string]*agg{}
	for _, h := range res.Docs {
		if h.Doc == nil {
			continue
		}
		ticketID := strings.TrimSpace(h.Doc.Ticket)
		if ticketID == "" {
			continue
		}
		a, ok := aggs[ticketID]
		if !ok {
			a = &agg{ticketID: ticketID}
			aggs[ticketID] = a
		}

		docPathOS := filepath.Clean(filepath.FromSlash(h.Path))
		if filepath.Base(docPathOS) == "index.md" || strings.TrimSpace(h.Doc.DocType) == "index" {
			a.title = h.Doc.Title
			a.status = h.Doc.Status
			a.lastUpdated = h.Doc.LastUpdated
			a.ticketDir = filepath.Dir(docPathOS)
			a.hasIndex = true
			continue
		}

		a.docs++
		switch strings.TrimSpace(h.Doc.DocType) {
		case "design-doc":
			a.designDocs++
		case "reference":
			a.referenceDoc++
		case "playbook":
			a.playbooks++
		}
	}

	out := make([]statusTicket, 0, len(aggs))
	sum := statusSummary{}
	for _, a := range aggs {
		if !a.hasIndex {
			continue
		}
		if ticketFilter != "" && strings.TrimSpace(a.ticketID) != strings.TrimSpace(ticketFilter) {
			continue
		}

		stale := false
		if !a.lastUpdated.IsZero() {
			days := time.Since(a.lastUpdated).Hours() / 24
			if int(days) > staleAfterDays {
				stale = true
			}
		}

		out = append(out, statusTicket{
			Ticket:        a.ticketID,
			Title:         a.title,
			Status:        a.status,
			Stale:         stale,
			Docs:          a.docs,
			DesignDocs:    a.designDocs,
			ReferenceDocs: a.referenceDoc,
			Playbooks:     a.playbooks,
			Path:          a.ticketDir,
			LastUpdated:   a.lastUpdated,
		})

		sum.TicketsTotal++
		if stale {
			sum.TicketsStale++
		}
		sum.DocsTotal += a.docs
		sum.DesignDocs += a.designDocs
		sum.ReferenceDocs += a.referenceDoc
		sum.Playbooks += a.playbooks
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Path < out[j].Path
	})

	return out, sum, nil
}

func formatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}
