package commands

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-go-golems/docmgr/internal/paths"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

type TicketGraphCommand struct {
	*cmds.CommandDescription
}

type TicketGraphSettings struct {
	Ticket             string `glazed.parameter:"ticket"`
	Root               string `glazed.parameter:"root"`
	Format             string `glazed.parameter:"format"`
	Direction          string `glazed.parameter:"direction"`
	Label              string `glazed.parameter:"label"`
	EdgeNotes          string `glazed.parameter:"edge-notes"`
	IncludeControlDocs bool   `glazed.parameter:"include-control-docs"`
	IncludeArchived    bool   `glazed.parameter:"include-archived"`
	IncludeScriptsPath bool   `glazed.parameter:"include-scripts-path"`
}

func NewTicketGraphCommand() (*TicketGraphCommand, error) {
	return &TicketGraphCommand{
		CommandDescription: cmds.NewCommandDescription(
			"graph",
			cmds.WithShort("Render a Mermaid graph for a ticket (docs ↔ related files)"),
			cmds.WithLong(`Render a Mermaid graph for a ticket showing:
- all markdown docs in the ticket workspace, and
- the code files referenced via frontmatter RelatedFiles.

This is the "depth=0" ticket graph: it does not yet expand transitively to other tickets.

Examples:
  # Pasteable Markdown with a mermaid code block (default)
  docmgr ticket graph --ticket MEN-4242

  # Raw Mermaid DSL
  docmgr ticket graph --ticket MEN-4242 --format mermaid

  # Structured edge list (for scripts)
  docmgr ticket graph --ticket MEN-4242 --with-glaze-output --output table
`),
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
				parameters.NewParameterDefinition(
					"format",
					parameters.ParameterTypeString,
					parameters.WithHelp("Output format: markdown|mermaid"),
					parameters.WithDefault("markdown"),
				),
				parameters.NewParameterDefinition(
					"direction",
					parameters.ParameterTypeString,
					parameters.WithHelp("Mermaid direction: TD|LR"),
					parameters.WithDefault("TD"),
				),
				parameters.NewParameterDefinition(
					"label",
					parameters.ParameterTypeString,
					parameters.WithHelp("Doc node label: title|path|both"),
					parameters.WithDefault("both"),
				),
				parameters.NewParameterDefinition(
					"edge-notes",
					parameters.ParameterTypeString,
					parameters.WithHelp("Include RelatedFiles.Note as edge label: none|short"),
					parameters.WithDefault("short"),
				),
				parameters.NewParameterDefinition(
					"include-control-docs",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Include control docs (README.md, tasks.md, changelog.md)"),
					parameters.WithDefault(true),
				),
				parameters.NewParameterDefinition(
					"include-archived",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Include documents under archive/"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"include-scripts-path",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Include documents under scripts/"),
					parameters.WithDefault(false),
				),
			),
		),
	}, nil
}

func (c *TicketGraphCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &TicketGraphSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	graph, err := buildTicketGraphDepth0(ctx, settings)
	if err != nil {
		return err
	}

	for _, e := range graph.edges {
		row := types.NewRow(
			types.MRP("from_type", "doc"),
			types.MRP("from_ticket", e.fromTicket),
			types.MRP("from_path", e.fromDocPath),
			types.MRP("from_title", e.fromTitle),
			types.MRP("to_type", "file"),
			types.MRP("to_path", e.toFileKey),
			types.MRP("label", e.label),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return fmt.Errorf("failed to emit edge row: %w", err)
		}
	}
	return nil
}

var _ cmds.GlazeCommand = &TicketGraphCommand{}

func (c *TicketGraphCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
) error {
	settings := &TicketGraphSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	graph, err := buildTicketGraphDepth0(ctx, settings)
	if err != nil {
		return err
	}

	out, err := renderMermaidTicketGraph(graph, settings)
	if err != nil {
		return err
	}
	fmt.Print(out)
	return nil
}

var _ cmds.BareCommand = &TicketGraphCommand{}

type ticketGraph struct {
	direction string
	docNodes  map[string]ticketGraphDocNode // key: abs doc path (slash-cleaned)
	fileNodes map[string]struct{}           // key: canonical file key
	edges     []ticketGraphEdge
}

type ticketGraphDocNode struct {
	pathRel string
	ticket  string
	title   string
	docType string
}

type ticketGraphEdge struct {
	fromDocPath string
	fromTicket  string
	fromTitle   string
	toFileKey   string
	label       string
}

func buildTicketGraphDepth0(ctx context.Context, settings *TicketGraphSettings) (*ticketGraph, error) {
	if strings.TrimSpace(settings.Ticket) == "" {
		return nil, fmt.Errorf("missing --ticket")
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

	res, err := ws.QueryDocs(ctx, workspace.DocQuery{
		Scope: workspace.Scope{Kind: workspace.ScopeTicket, TicketID: strings.TrimSpace(settings.Ticket)},
		Options: workspace.DocQueryOptions{
			IncludeErrors:       false,
			IncludeDiagnostics:  false,
			IncludeBody:         false,
			IncludeControlDocs:  settings.IncludeControlDocs,
			IncludeArchivedPath: settings.IncludeArchived,
			IncludeScriptsPath:  settings.IncludeScriptsPath,
			OrderBy:             workspace.OrderByPath,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query docs: %w", err)
	}

	g := &ticketGraph{
		direction: strings.TrimSpace(settings.Direction),
		docNodes:  map[string]ticketGraphDocNode{},
		fileNodes: map[string]struct{}{},
	}

	for _, h := range res.Docs {
		if h.Doc == nil {
			continue
		}

		docPathAbs := filepath.ToSlash(filepath.Clean(h.Path))
		docPathRel := docPathAbs
		if rel, err := filepath.Rel(ws.Context().Root, filepath.FromSlash(docPathAbs)); err == nil {
			docPathRel = filepath.ToSlash(rel)
		}

		g.docNodes[docPathAbs] = ticketGraphDocNode{
			pathRel: docPathRel,
			ticket:  h.Doc.Ticket,
			title:   h.Doc.Title,
			docType: h.Doc.DocType,
		}

		docResolver := paths.NewResolver(paths.ResolverOptions{
			DocsRoot:  ws.Context().Root,
			DocPath:   filepath.FromSlash(docPathAbs),
			ConfigDir: ws.Context().ConfigDir,
			RepoRoot:  ws.Context().RepoRoot,
		})

		for _, rf := range h.Doc.RelatedFiles {
			fileKey := canonicalizeForGraph(docResolver, rf.Path)
			if strings.TrimSpace(fileKey) == "" {
				continue
			}
			g.fileNodes[fileKey] = struct{}{}
			g.edges = append(g.edges, ticketGraphEdge{
				fromDocPath: docPathAbs,
				fromTicket:  h.Doc.Ticket,
				fromTitle:   h.Doc.Title,
				toFileKey:   fileKey,
				label:       edgeLabel(settings.EdgeNotes, rf.Note),
			})
		}
	}

	// Stable output ordering.
	sort.Slice(g.edges, func(i, j int) bool {
		if g.edges[i].fromDocPath != g.edges[j].fromDocPath {
			return g.edges[i].fromDocPath < g.edges[j].fromDocPath
		}
		if g.edges[i].toFileKey != g.edges[j].toFileKey {
			return g.edges[i].toFileKey < g.edges[j].toFileKey
		}
		return g.edges[i].label < g.edges[j].label
	})

	return g, nil
}

func canonicalizeForGraph(resolver *paths.Resolver, raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if resolver == nil {
		return filepath.ToSlash(raw)
	}
	n := resolver.Normalize(raw)
	switch {
	case strings.TrimSpace(n.Canonical) != "":
		return filepath.ToSlash(strings.TrimSpace(n.Canonical))
	case strings.TrimSpace(n.Abs) != "":
		return filepath.ToSlash(strings.TrimSpace(n.Abs))
	case strings.TrimSpace(n.OriginalClean) != "":
		return filepath.ToSlash(strings.TrimSpace(n.OriginalClean))
	default:
		return filepath.ToSlash(raw)
	}
}

func edgeLabel(mode string, note string) string {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "" || mode == "short" {
		return sanitizeMermaidLabel(note, 80)
	}
	if mode == "none" {
		return ""
	}
	return sanitizeMermaidLabel(note, 80)
}

func renderMermaidTicketGraph(g *ticketGraph, settings *TicketGraphSettings) (string, error) {
	if g == nil {
		return "", fmt.Errorf("nil graph")
	}

	direction := strings.ToUpper(strings.TrimSpace(g.direction))
	if direction == "" {
		direction = "TD"
	}
	if direction != "TD" && direction != "LR" {
		return "", fmt.Errorf("invalid --direction %q (expected TD or LR)", direction)
	}

	format := strings.ToLower(strings.TrimSpace(settings.Format))
	if format == "" {
		format = "markdown"
	}
	if format != "markdown" && format != "mermaid" {
		return "", fmt.Errorf("invalid --format %q (expected markdown or mermaid)", format)
	}

	labelMode := strings.ToLower(strings.TrimSpace(settings.Label))
	if labelMode == "" {
		labelMode = "both"
	}
	if labelMode != "title" && labelMode != "path" && labelMode != "both" {
		return "", fmt.Errorf("invalid --label %q (expected title, path, or both)", labelMode)
	}

	type node struct {
		id    string
		label string
		class string
	}

	docKeys := make([]string, 0, len(g.docNodes))
	for k := range g.docNodes {
		docKeys = append(docKeys, k)
	}
	sort.Strings(docKeys)

	fileKeys := make([]string, 0, len(g.fileNodes))
	for k := range g.fileNodes {
		fileKeys = append(fileKeys, k)
	}
	sort.Strings(fileKeys)

	nodes := make([]node, 0, len(docKeys)+len(fileKeys))
	mermaidIDByDoc := map[string]string{}
	mermaidIDByFile := map[string]string{}

	for _, k := range docKeys {
		n := g.docNodes[k]
		id := "D_" + shortHash(k)
		mermaidIDByDoc[k] = id
		lbl := buildDocLabel(n, labelMode)
		nodes = append(nodes, node{id: id, label: sanitizeMermaidLabel(lbl, 180), class: "doc"})
	}
	for _, k := range fileKeys {
		id := "F_" + shortHash(k)
		mermaidIDByFile[k] = id
		nodes = append(nodes, node{id: id, label: sanitizeMermaidLabel(k, 180), class: "file"})
	}

	var b strings.Builder
	if format == "markdown" {
		b.WriteString("```mermaid\n")
	}
	b.WriteString("graph ")
	b.WriteString(direction)
	b.WriteString("\n")

	for _, n := range nodes {
		b.WriteString("  ")
		b.WriteString(n.id)
		b.WriteString("[\"")
		b.WriteString(n.label)
		b.WriteString("\"]\n")
	}

	for _, e := range g.edges {
		fromID := mermaidIDByDoc[e.fromDocPath]
		toID := mermaidIDByFile[e.toFileKey]
		if fromID == "" || toID == "" {
			continue
		}
		b.WriteString("  ")
		b.WriteString(fromID)
		if strings.TrimSpace(e.label) != "" {
			b.WriteString(" -->|")
			b.WriteString(sanitizeMermaidEdgeLabel(e.label, 80))
			b.WriteString("| ")
			b.WriteString(toID)
			b.WriteString("\n")
			continue
		}
		b.WriteString(" --> ")
		b.WriteString(toID)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString("  classDef doc fill:#eef,stroke:#446,stroke-width:1px;\n")
	b.WriteString("  classDef file fill:#efe,stroke:#464,stroke-width:1px;\n")

	// Apply classes (Mermaid supports comma-separated node lists).
	if len(docKeys) > 0 {
		b.WriteString("  class ")
		for i, k := range docKeys {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString(mermaidIDByDoc[k])
		}
		b.WriteString(" doc;\n")
	}
	if len(fileKeys) > 0 {
		b.WriteString("  class ")
		for i, k := range fileKeys {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString(mermaidIDByFile[k])
		}
		b.WriteString(" file;\n")
	}

	if format == "markdown" {
		b.WriteString("```\n")
	}

	return b.String(), nil
}

func buildDocLabel(n ticketGraphDocNode, labelMode string) string {
	title := strings.TrimSpace(n.title)
	if title == "" {
		title = "(untitled)"
	}
	path := strings.TrimSpace(n.pathRel)
	switch labelMode {
	case "title":
		return title
	case "path":
		return path
	default:
		if path == "" {
			return title
		}
		if strings.TrimSpace(n.docType) != "" {
			return fmt.Sprintf("%s: %s\n(%s)", n.docType, title, path)
		}
		return fmt.Sprintf("%s\n(%s)", title, path)
	}
}

func shortHash(s string) string {
	sum := sha1.Sum([]byte(s))
	return hex.EncodeToString(sum[:])[:10]
}

func sanitizeMermaidLabel(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\"", "'")
	s = strings.ReplaceAll(s, "[", "(")
	s = strings.ReplaceAll(s, "]", ")")
	s = strings.ReplaceAll(s, "|", "/")
	if maxLen > 0 && len(s) > maxLen {
		s = s[:maxLen] + "…"
	}
	return s
}

func sanitizeMermaidEdgeLabel(s string, maxLen int) string {
	// Edge labels are surrounded by |...|, so be stricter.
	s = sanitizeMermaidLabel(s, maxLen)
	s = strings.ReplaceAll(s, "|", "/")
	return s
}
