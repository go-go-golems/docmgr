package ticketgraph

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
)

type Stats struct {
	Nodes int `json:"nodes"`
	Edges int `json:"edges"`
}

type graph struct {
	direction string
	docNodes  map[string]docNode  // key: abs doc path (slash-cleaned)
	fileNodes map[string]struct{} // key: canonical file key
	edges     []edge
}

type docNode struct {
	pathRel string
	ticket  string
	title   string
	docType string
}

type edge struct {
	fromDocPath string
	fromTicket  string
	fromTitle   string
	toFileKey   string
	label       string
}

func BuildMermaid(ctx context.Context, ws *workspace.Workspace, ticketID string, direction string, includeArchived bool, includeScripts bool, includeControlDocs bool) (string, Stats, error) {
	if ws == nil {
		return "", Stats{}, fmt.Errorf("nil workspace")
	}
	ticketID = strings.TrimSpace(ticketID)
	if ticketID == "" {
		return "", Stats{}, fmt.Errorf("missing ticket")
	}
	direction = strings.TrimSpace(direction)
	if direction == "" {
		direction = "TD"
	}
	if direction != "TD" && direction != "LR" {
		return "", Stats{}, fmt.Errorf("invalid direction %q", direction)
	}

	g := &graph{
		direction: direction,
		docNodes:  map[string]docNode{},
		fileNodes: map[string]struct{}{},
	}
	edgeSet := map[string]struct{}{}

	res, err := ws.QueryDocs(ctx, workspace.DocQuery{
		Scope: workspace.Scope{Kind: workspace.ScopeTicket, TicketID: ticketID},
		Options: workspace.DocQueryOptions{
			IncludeErrors:       false,
			IncludeDiagnostics:  false,
			IncludeBody:         false,
			IncludeControlDocs:  includeControlDocs,
			IncludeArchivedPath: includeArchived,
			IncludeScriptsPath:  includeScripts,
			OrderBy:             workspace.OrderByPath,
		},
	})
	if err != nil {
		return "", Stats{}, err
	}

	for _, h := range res.Docs {
		if h.Doc == nil {
			continue
		}
		addDocAndEdges(ws, g, edgeSet, h.Path, h.Doc)
	}

	sort.Slice(g.edges, func(i, j int) bool {
		if g.edges[i].fromDocPath != g.edges[j].fromDocPath {
			return g.edges[i].fromDocPath < g.edges[j].fromDocPath
		}
		if g.edges[i].toFileKey != g.edges[j].toFileKey {
			return g.edges[i].toFileKey < g.edges[j].toFileKey
		}
		return g.edges[i].label < g.edges[j].label
	})

	out := renderMermaid(ws, g)
	stats := Stats{
		Nodes: len(g.docNodes) + len(g.fileNodes),
		Edges: len(g.edges),
	}
	return out, stats, nil
}

func addDocAndEdges(ws *workspace.Workspace, g *graph, edgeSet map[string]struct{}, docPathAbs string, doc *models.Document) {
	if ws == nil || g == nil || doc == nil {
		return
	}

	docPathAbs = filepath.ToSlash(filepath.Clean(docPathAbs))
	if _, ok := g.docNodes[docPathAbs]; !ok {
		docPathRel := docPathAbs
		if rel, err := filepath.Rel(ws.Context().Root, filepath.FromSlash(docPathAbs)); err == nil {
			docPathRel = filepath.ToSlash(rel)
		}
		g.docNodes[docPathAbs] = docNode{
			pathRel: docPathRel,
			ticket:  doc.Ticket,
			title:   doc.Title,
			docType: doc.DocType,
		}
	}

	docResolver := paths.NewResolver(paths.ResolverOptions{
		DocsRoot:  ws.Context().Root,
		ConfigDir: ws.Context().ConfigDir,
		RepoRoot:  ws.Context().RepoRoot,
		DocPath:   filepath.FromSlash(docPathAbs),
	})

	for _, rf := range doc.RelatedFiles {
		fileKey := canonicalizeForGraph(docResolver, rf.Path)
		if strings.TrimSpace(fileKey) == "" {
			continue
		}
		g.fileNodes[fileKey] = struct{}{}
		label := edgeLabel(rf.Note)
		k := docPathAbs + "\x00" + fileKey + "\x00" + label
		if _, ok := edgeSet[k]; ok {
			continue
		}
		edgeSet[k] = struct{}{}
		g.edges = append(g.edges, edge{
			fromDocPath: docPathAbs,
			fromTicket:  doc.Ticket,
			fromTitle:   doc.Title,
			toFileKey:   fileKey,
			label:       label,
		})
	}
}

func edgeLabel(note string) string {
	note = strings.TrimSpace(note)
	if note == "" {
		return ""
	}
	if len(note) > 40 {
		note = strings.TrimSpace(note[:40]) + "…"
	}
	return note
}

func canonicalizeForGraph(resolver *paths.Resolver, raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || resolver == nil {
		return ""
	}
	n := resolver.Normalize(raw)
	switch {
	case strings.TrimSpace(n.Canonical) != "":
		return filepath.ToSlash(strings.TrimSpace(n.Canonical))
	case strings.TrimSpace(n.RepoRelative) != "":
		return filepath.ToSlash(strings.TrimSpace(n.RepoRelative))
	case strings.TrimSpace(n.OriginalClean) != "":
		return filepath.ToSlash(strings.TrimSpace(n.OriginalClean))
	default:
		return ""
	}
}

type node struct {
	id    string
	label string
	class string
}

func renderMermaid(ws *workspace.Workspace, g *graph) string {
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

	docIDByPath := map[string]string{}
	fileIDByKey := map[string]string{}

	nodes := make([]node, 0, len(docKeys)+len(fileKeys))
	for _, k := range docKeys {
		n := g.docNodes[k]
		id := mermaidID("doc", k)
		docIDByPath[k] = id
		label := n.title
		if strings.TrimSpace(label) == "" {
			label = filepath.Base(n.pathRel)
		}
		nodes = append(nodes, node{id: id, label: label, class: "doc"})
	}
	for _, k := range fileKeys {
		id := mermaidID("file", k)
		fileIDByKey[k] = id
		label := k
		nodes = append(nodes, node{id: id, label: label, class: "file"})
	}

	var b strings.Builder
	b.WriteString("graph ")
	b.WriteString(g.direction)
	b.WriteString("\n")

	for _, n := range nodes {
		b.WriteString("  ")
		b.WriteString(n.id)
		b.WriteString("[\"")
		b.WriteString(escapeMermaidLabel(n.label))
		b.WriteString("\"]:::")
		b.WriteString(n.class)
		b.WriteString("\n")
	}

	for _, e := range g.edges {
		from := docIDByPath[e.fromDocPath]
		to := fileIDByKey[e.toFileKey]
		if from == "" || to == "" {
			continue
		}
		b.WriteString("  ")
		b.WriteString(from)
		if strings.TrimSpace(e.label) != "" {
			b.WriteString(" -- \"")
			b.WriteString(escapeMermaidLabel(e.label))
			b.WriteString("\" --> ")
			b.WriteString(to)
		} else {
			b.WriteString(" --> ")
			b.WriteString(to)
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString("classDef doc fill:#e8f2ff,stroke:#2b6cb0,stroke-width:1px;\n")
	b.WriteString("classDef file fill:#f6f6f6,stroke:#555,stroke-width:1px;\n")

	_ = ws
	return b.String()
}

func mermaidID(prefix string, key string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return fmt.Sprintf("%s_%08x", prefix, h.Sum32())
}

func escapeMermaidLabel(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}
