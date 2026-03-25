package commands

import (
	"context"
	"fmt"
	"hash/fnv"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-go-golems/docmgr/internal/paths"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/docmgr/pkg/models"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

type TicketGraphCommand struct {
	*cmds.CommandDescription
}

type TicketGraphSettings struct {
	Ticket             string `glazed:"ticket"`
	Root               string `glazed:"root"`
	Format             string `glazed:"format"`
	Direction          string `glazed:"direction"`
	Label              string `glazed:"label"`
	EdgeNotes          string `glazed:"edge-notes"`
	Depth              int    `glazed:"depth"`
	Scope              string `glazed:"scope"`
	ExpandFiles        bool   `glazed:"expand-files"`
	MaxNodes           int    `glazed:"max-nodes"`
	MaxEdges           int    `glazed:"max-edges"`
	BatchSize          int    `glazed:"batch-size"`
	IncludeControlDocs bool   `glazed:"include-control-docs"`
	IncludeArchived    bool   `glazed:"include-archived"`
	IncludeScriptsPath bool   `glazed:"include-scripts-path"`
}

func NewTicketGraphCommand() (*TicketGraphCommand, error) {
	return &TicketGraphCommand{
		CommandDescription: cmds.NewCommandDescription(
			"graph",
			cmds.WithShort("Render a Mermaid graph for a ticket (docs ↔ related files)"),
			cmds.WithLong(`Render a Mermaid graph for a ticket showing:
- all markdown docs in the ticket workspace, and
- the code files referenced via frontmatter RelatedFiles.

With --scope repo and --depth > 0, the command expands the graph transitively:
  docs -> related files -> other docs that reference those files -> ...

Safety: transitive expansion can grow quickly; use --max-nodes/--max-edges and keep depth small.

Examples:
  # Pasteable Markdown with a mermaid code block (default)
  docmgr ticket graph --ticket MEN-4242

  # Raw Mermaid DSL
  docmgr ticket graph --ticket MEN-4242 --format mermaid

  # Structured edge list (for scripts)
  docmgr ticket graph --ticket MEN-4242 --with-glaze-output --output table

  # Repo-wide transitive expansion (depth 1), do not expand file frontier
  docmgr ticket graph --ticket MEN-4242 --scope repo --depth 1 --expand-files=false
`),
			cmds.WithFlags(
				fields.New(
					"ticket",
					fields.TypeString,
					fields.WithHelp("Ticket identifier"),
					fields.WithRequired(true),
				),
				fields.New(
					"root",
					fields.TypeString,
					fields.WithHelp("Root directory for docs"),
					fields.WithDefault("ttmp"),
				),
				fields.New(
					"format",
					fields.TypeString,
					fields.WithHelp("Output format: markdown|mermaid"),
					fields.WithDefault("markdown"),
				),
				fields.New(
					"direction",
					fields.TypeString,
					fields.WithHelp("Mermaid direction: TD|LR"),
					fields.WithDefault("TD"),
				),
				fields.New(
					"label",
					fields.TypeString,
					fields.WithHelp("Doc node label: title|path|both"),
					fields.WithDefault("both"),
				),
				fields.New(
					"edge-notes",
					fields.TypeString,
					fields.WithHelp("Include RelatedFiles.Note as edge label: none|short"),
					fields.WithDefault("short"),
				),
				fields.New(
					"depth",
					fields.TypeInteger,
					fields.WithHelp("Transitive expansion depth (0=ticket docs + their related files only)"),
					fields.WithDefault(0),
				),
				fields.New(
					"scope",
					fields.TypeString,
					fields.WithHelp("Graph expansion scope: ticket|repo (repo required for depth>0)"),
					fields.WithDefault("ticket"),
				),
				fields.New(
					"expand-files",
					fields.TypeBool,
					fields.WithHelp("When expanding to new docs, also add their RelatedFiles to the file frontier"),
					fields.WithDefault(false),
				),
				fields.New(
					"max-nodes",
					fields.TypeInteger,
					fields.WithHelp("Maximum total nodes (docs + files) before failing"),
					fields.WithDefault(500),
				),
				fields.New(
					"max-edges",
					fields.TypeInteger,
					fields.WithHelp("Maximum total edges before failing"),
					fields.WithDefault(2000),
				),
				fields.New(
					"batch-size",
					fields.TypeInteger,
					fields.WithHelp("Batch size for repo-scope reverse lookup queries during expansion"),
					fields.WithDefault(50),
				),
				fields.New(
					"include-control-docs",
					fields.TypeBool,
					fields.WithHelp("Include control docs (README.md, tasks.md, changelog.md)"),
					fields.WithDefault(true),
				),
				fields.New(
					"include-archived",
					fields.TypeBool,
					fields.WithHelp("Include documents under archive/"),
					fields.WithDefault(false),
				),
				fields.New(
					"include-scripts-path",
					fields.TypeBool,
					fields.WithHelp("Include documents under scripts/"),
					fields.WithDefault(false),
				),
			),
		),
	}, nil
}

func (c *TicketGraphCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedValues *values.Values,
	gp middlewares.Processor,
) error {
	settings := &TicketGraphSettings{}
	if err := parsedValues.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	graph, err := buildTicketGraph(ctx, settings)
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
	parsedValues *values.Values,
) error {
	settings := &TicketGraphSettings{}
	if err := parsedValues.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	graph, err := buildTicketGraph(ctx, settings)
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

type ticketGraphBuilder struct {
	ws       *workspace.Workspace
	settings *TicketGraphSettings

	graph   *ticketGraph
	edgeSet map[string]struct{}
}

func buildTicketGraph(ctx context.Context, settings *TicketGraphSettings) (*ticketGraph, error) {
	if strings.TrimSpace(settings.Ticket) == "" {
		return nil, fmt.Errorf("missing --ticket")
	}

	if settings.Depth < 0 {
		return nil, fmt.Errorf("invalid --depth %d (must be >= 0)", settings.Depth)
	}
	settings.Scope = strings.ToLower(strings.TrimSpace(settings.Scope))
	if settings.Scope == "" {
		settings.Scope = "ticket"
	}
	if settings.Scope != "ticket" && settings.Scope != "repo" {
		return nil, fmt.Errorf("invalid --scope %q (expected ticket or repo)", settings.Scope)
	}
	if settings.Depth > 0 && settings.Scope != "repo" {
		return nil, fmt.Errorf("--depth %d requires --scope repo (refusing to expand without explicit repo scope)", settings.Depth)
	}
	if settings.BatchSize <= 0 {
		return nil, fmt.Errorf("invalid --batch-size %d (must be > 0)", settings.BatchSize)
	}
	if settings.MaxNodes <= 0 {
		return nil, fmt.Errorf("invalid --max-nodes %d (must be > 0)", settings.MaxNodes)
	}
	if settings.MaxEdges <= 0 {
		return nil, fmt.Errorf("invalid --max-edges %d (must be > 0)", settings.MaxEdges)
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

	b := &ticketGraphBuilder{
		ws:       ws,
		settings: settings,
		graph: &ticketGraph{
			direction: strings.TrimSpace(settings.Direction),
			docNodes:  map[string]ticketGraphDocNode{},
			fileNodes: map[string]struct{}{},
		},
		edgeSet: map[string]struct{}{},
	}
	if err := b.addTicketDocsDepth0(ctx); err != nil {
		return nil, err
	}
	if settings.Depth > 0 {
		if err := b.expandTransitive(ctx); err != nil {
			return nil, err
		}
	}

	// Stable output ordering.
	sort.Slice(b.graph.edges, func(i, j int) bool {
		if b.graph.edges[i].fromDocPath != b.graph.edges[j].fromDocPath {
			return b.graph.edges[i].fromDocPath < b.graph.edges[j].fromDocPath
		}
		if b.graph.edges[i].toFileKey != b.graph.edges[j].toFileKey {
			return b.graph.edges[i].toFileKey < b.graph.edges[j].toFileKey
		}
		return b.graph.edges[i].label < b.graph.edges[j].label
	})

	return b.graph, nil
}

func (b *ticketGraphBuilder) addTicketDocsDepth0(ctx context.Context) error {
	res, err := b.ws.QueryDocs(ctx, workspace.DocQuery{
		Scope: workspace.Scope{Kind: workspace.ScopeTicket, TicketID: strings.TrimSpace(b.settings.Ticket)},
		Options: workspace.DocQueryOptions{
			IncludeErrors:       false,
			IncludeDiagnostics:  false,
			IncludeBody:         false,
			IncludeControlDocs:  b.settings.IncludeControlDocs,
			IncludeArchivedPath: b.settings.IncludeArchived,
			IncludeScriptsPath:  b.settings.IncludeScriptsPath,
			OrderBy:             workspace.OrderByPath,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to query ticket docs: %w", err)
	}

	for _, h := range res.Docs {
		if h.Doc == nil {
			continue
		}
		if err := b.addDocAndEdges(h.Path, h.Doc, nil); err != nil {
			return err
		}
	}
	return nil
}

func (b *ticketGraphBuilder) expandTransitive(ctx context.Context) error {
	frontier := make([]string, 0, len(b.graph.fileNodes))
	for k := range b.graph.fileNodes {
		frontier = append(frontier, k)
	}
	sort.Strings(frontier)

	for level := 1; level <= b.settings.Depth; level++ {
		if len(frontier) == 0 {
			return nil
		}

		var nextFrontier []string
		seenThisLevel := map[string]struct{}{}

		for i := 0; i < len(frontier); i += b.settings.BatchSize {
			end := i + b.settings.BatchSize
			if end > len(frontier) {
				end = len(frontier)
			}
			batch := frontier[i:end]

			res, err := b.ws.QueryDocs(ctx, workspace.DocQuery{
				Scope: workspace.Scope{Kind: workspace.ScopeRepo},
				Filters: workspace.DocFilters{
					RelatedFile: batch,
				},
				Options: workspace.DocQueryOptions{
					IncludeErrors:       false,
					IncludeDiagnostics:  false,
					IncludeBody:         false,
					IncludeControlDocs:  b.settings.IncludeControlDocs,
					IncludeArchivedPath: b.settings.IncludeArchived,
					IncludeScriptsPath:  b.settings.IncludeScriptsPath,
					OrderBy:             workspace.OrderByPath,
				},
			})
			if err != nil {
				return fmt.Errorf("failed to expand via related file batch (level=%d): %w", level, err)
			}

			for _, h := range res.Docs {
				if h.Doc == nil {
					continue
				}
				if strings.TrimSpace(h.Doc.Ticket) == "" {
					continue
				}
				// In repo scope we include all tickets; keep this check here in case we add more scopes later.
				if b.settings.Scope != "repo" && strings.TrimSpace(h.Doc.Ticket) != strings.TrimSpace(b.settings.Ticket) {
					continue
				}

				if err := b.addDocAndEdges(h.Path, h.Doc, batch); err != nil {
					return err
				}

				if b.settings.ExpandFiles {
					for _, rf := range h.Doc.RelatedFiles {
						key := canonicalizeForGraph(b.docResolver(h.Path), rf.Path)
						if strings.TrimSpace(key) == "" {
							continue
						}
						if _, ok := b.graph.fileNodes[key]; ok {
							continue
						}
						if _, ok := seenThisLevel[key]; ok {
							continue
						}
						seenThisLevel[key] = struct{}{}
						nextFrontier = append(nextFrontier, key)
					}
				}
			}
		}

		sort.Strings(nextFrontier)
		frontier = nextFrontier
	}

	return nil
}

func (b *ticketGraphBuilder) addDocAndEdges(docPath string, doc *models.Document, triggerFiles []string) error {
	if doc == nil {
		return nil
	}

	docPathAbs := filepath.ToSlash(filepath.Clean(docPath))
	if _, ok := b.graph.docNodes[docPathAbs]; !ok {
		if err := b.ensureNodeBudget(1); err != nil {
			return err
		}

		docPathRel := docPathAbs
		if rel, err := filepath.Rel(b.ws.Context().Root, filepath.FromSlash(docPathAbs)); err == nil {
			docPathRel = filepath.ToSlash(rel)
		}
		b.graph.docNodes[docPathAbs] = ticketGraphDocNode{
			pathRel: docPathRel,
			ticket:  doc.Ticket,
			title:   doc.Title,
			docType: doc.DocType,
		}
	}

	docResolver := b.docResolver(docPathAbs)
	triggerSet := map[string]struct{}{}
	triggerBasenames := map[string]struct{}{}
	if len(triggerFiles) > 0 && !b.settings.ExpandFiles {
		for _, t := range triggerFiles {
			t = strings.TrimSpace(t)
			if t == "" {
				continue
			}
			triggerSet[t] = struct{}{}
			// Mirror QueryDocs behavior: if the trigger is basename-only (no separators),
			// QueryDocs enables suffix matching ("%/basename"). When a doc is pulled in via
			// such a suffix match, its canonicalized RelatedFiles entry will typically be
			// repo-relative (e.g. "pkg/main.go") and would otherwise be dropped here.
			if strings.Contains(t, "/") || strings.Contains(t, "\\") {
				continue
			}
			base := filepath.ToSlash(filepath.Clean(t))
			if base == "" || base == "." || base == "/" {
				continue
			}
			triggerBasenames[base] = struct{}{}
		}
	}

	for _, rf := range doc.RelatedFiles {
		fileKey := canonicalizeForGraph(docResolver, rf.Path)
		if strings.TrimSpace(fileKey) == "" {
			continue
		}
		if len(triggerSet) > 0 {
			if _, ok := triggerSet[fileKey]; !ok {
				matched := false
				for base := range triggerBasenames {
					if fileKey == base || strings.HasSuffix(fileKey, "/"+base) {
						matched = true
						break
					}
				}
				if !matched {
					continue
				}
			}
		}

		if _, ok := b.graph.fileNodes[fileKey]; !ok {
			if err := b.ensureNodeBudget(1); err != nil {
				return err
			}
			b.graph.fileNodes[fileKey] = struct{}{}
		}

		label := edgeLabel(b.settings.EdgeNotes, rf.Note)
		edgeKey := docPathAbs + "\x00" + fileKey + "\x00" + label
		if _, ok := b.edgeSet[edgeKey]; ok {
			continue
		}
		if err := b.ensureEdgeBudget(1); err != nil {
			return err
		}
		b.edgeSet[edgeKey] = struct{}{}
		b.graph.edges = append(b.graph.edges, ticketGraphEdge{
			fromDocPath: docPathAbs,
			fromTicket:  doc.Ticket,
			fromTitle:   doc.Title,
			toFileKey:   fileKey,
			label:       label,
		})
	}

	return nil
}

func (b *ticketGraphBuilder) docResolver(docPath string) *paths.Resolver {
	docPathAbs := filepath.ToSlash(filepath.Clean(docPath))
	return paths.NewResolver(paths.ResolverOptions{
		DocsRoot:  b.ws.Context().Root,
		DocPath:   filepath.FromSlash(docPathAbs),
		ConfigDir: b.ws.Context().ConfigDir,
		RepoRoot:  b.ws.Context().RepoRoot,
	})
}

func (b *ticketGraphBuilder) ensureNodeBudget(delta int) error {
	if delta <= 0 {
		return nil
	}
	current := len(b.graph.docNodes) + len(b.graph.fileNodes)
	next := current + delta
	if next > b.settings.MaxNodes {
		return fmt.Errorf("graph would exceed --max-nodes=%d (current=%d, next=%d); increase --max-nodes or reduce --depth/--scope/--expand-files", b.settings.MaxNodes, current, next)
	}
	return nil
}

func (b *ticketGraphBuilder) ensureEdgeBudget(delta int) error {
	if delta <= 0 {
		return nil
	}
	current := len(b.graph.edges)
	next := current + delta
	if next > b.settings.MaxEdges {
		return fmt.Errorf("graph would exceed --max-edges=%d (current=%d, next=%d); increase --max-edges or reduce --depth/--scope/--expand-files", b.settings.MaxEdges, current, next)
	}
	return nil
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
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return fmt.Sprintf("%016x", h.Sum64())[:10]
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
