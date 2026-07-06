package commands

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-go-golems/docmgr/internal/documents"
	"github.com/go-go-golems/docmgr/internal/paths"
	"github.com/go-go-golems/docmgr/internal/searchsvc"
	"github.com/go-go-golems/docmgr/internal/tickets"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/docmgr/pkg/models"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

// RelateCommand updates RelatedFiles metadata and can suggest files to relate
type RelateCommand struct {
	*cmds.CommandDescription
}

type RelateSettings struct {
	Ticket           string   `glazed:"ticket"`
	Doc              string   `glazed:"doc"`
	RemoveFiles      []string `glazed:"remove-files"`
	FileNotes        []string `glazed:"file-note"`
	Suggest          bool     `glazed:"suggest"`
	ApplySuggestions bool     `glazed:"apply-suggestions"`
	FromGit          bool     `glazed:"from-git"`
	Query            string   `glazed:"query"`
	Topics           []string `glazed:"topics"`
	Root             string   `glazed:"root"`
}

type RelateSuggestion struct {
	File    string
	Reasons []string
}

type RelateUpdateSummary struct {
	DocPath string
	Added   int
	Updated int
	Removed int
	Total   int
}

type RelateResult struct {
	Suggestions []RelateSuggestion
	Update      *RelateUpdateSummary
}

type reasonSet map[string]bool

func NewRelateCommand() (*RelateCommand, error) {
	return &RelateCommand{
		CommandDescription: cmds.NewCommandDescription(
			"relate",
			cmds.WithShort("Relate code files to a document or ticket"),
			cmds.WithLong(`Update RelatedFiles in a document's frontmatter or the ticket index.

Examples:
  # Relate multiple files to the ticket index (notes required; repeat --file-note)
  docmgr doc relate --ticket MEN-4242 \
    --file-note "backend/chat/api/register.go:Registers API routes" \
    --file-note "backend/chat/ws/manager.go:WebSocket lifecycle management" \
    --file-note "web/src/store/api/chatApi.ts:Frontend integration"

  # Relate multiple files to a specific document (notes required)
  docmgr doc relate --doc ttmp/YYYY/MM/DD/MEN-4242--.../design/path-normalization-strategy.md \
    --file-note "backend/chat/ws/manager.go:WebSocket lifecycle management" \
    --file-note "backend/chat/ws/heartbeat.go:Ping/pong behavior and timeouts"

  # Remove multiple related files (comma-separated)
  docmgr doc relate --ticket MEN-4242 --remove-files "backend/chat/ws/heartbeat.go,web/src/store/api/chatApi.ts"
`),
			cmds.WithFlags(
				fields.New(
					"ticket",
					fields.TypeString,
					fields.WithHelp("Ticket identifier (updates ticket index when --doc not provided)"),
					fields.WithDefault(""),
				),
				fields.New(
					"doc",
					fields.TypeString,
					fields.WithHelp("Path to a specific document to update"),
					fields.WithDefault(""),
				),
				fields.New(
					"remove-files",
					fields.TypeStringList,
					fields.WithHelp("Comma-separated list of files to remove from RelatedFiles"),
					fields.WithDefault([]string{}),
				),
				fields.New(
					"file-note",
					fields.TypeStringList,
					fields.WithHelp("Repeatable path-to-note mapping (format: path:note or path=note)"),
					fields.WithDefault([]string{}),
				),
				fields.New(
					"suggest",
					fields.TypeBool,
					fields.WithHelp("Suggest related files using heuristics (git + ripgrep + existing docs)"),
					fields.WithDefault(false),
				),
				fields.New(
					"apply-suggestions",
					fields.TypeBool,
					fields.WithHelp("Apply suggested files to the target document (requires --suggest)"),
					fields.WithDefault(false),
				),
				fields.New(
					"from-git",
					fields.TypeBool,
					fields.WithHelp("Limit suggestions to changed files from git status (modified, staged, untracked)"),
					fields.WithDefault(false),
				),
				fields.New(
					"query",
					fields.TypeString,
					fields.WithHelp("Optional query to seed suggestions (e.g., a keyword)"),
					fields.WithDefault(""),
				),
				fields.New(
					"topics",
					fields.TypeStringList,
					fields.WithHelp("Topics to seed suggestions (comma-separated)"),
					fields.WithDefault([]string{}),
				),
				fields.New(
					"root",
					fields.TypeString,
					fields.WithHelp("Root directory for docs"),
					fields.WithDefault("ttmp"),
				),
			),
		),
	}, nil
}

func (c *RelateCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedValues *values.Values,
	gp middlewares.Processor,
) error {
	settings := &RelateSettings{}
	if err := parsedValues.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Apply config root if present
	settings.Root = workspace.ResolveRoot(settings.Root)

	// Discover workspace + build ephemeral index so we can resolve target docs via QueryDocs
	// (Spec §11.2.4 / §5.1).
	ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{RootOverride: settings.Root})
	if err != nil {
		return fmt.Errorf("failed to discover workspace: %w", err)
	}
	if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{IncludeBody: false}); err != nil {
		return fmt.Errorf("failed to initialize workspace index: %w", err)
	}

	// Resolve target document path
	var targetDocPath string
	var ticketDir string

	if settings.Doc != "" {
		// Resolve the --doc reference forgivingly (absolute / repo-relative /
		// docs-root-relative / unique suffix), then verify it via the index.
		resolvedDoc, err := resolveDocRef(ctx, ws, settings.Root, settings.Doc)
		if err != nil {
			return err
		}
		res, err := ws.QueryDocs(ctx, workspace.DocQuery{
			Scope:   workspace.Scope{Kind: workspace.ScopeDoc, DocPath: resolvedDoc},
			Options: workspace.DocQueryOptions{IncludeErrors: true},
		})
		if err != nil {
			return fmt.Errorf("failed to resolve --doc via workspace index: %w", err)
		}
		if len(res.Docs) != 1 {
			return fmt.Errorf("--doc %q resolved to %s, but it is not in the docs index (is it under the docs root?)", settings.Doc, resolvedDoc)
		}
		if res.Docs[0].ReadErr != nil {
			return fmt.Errorf("document has invalid frontmatter (fix before relating files): %s: %v", res.Docs[0].Path, res.Docs[0].ReadErr)
		}
		targetDocPath = res.Docs[0].Path
	} else {
		if settings.Ticket == "" {
			return fmt.Errorf("must specify either --doc or --ticket")
		}
		// Forgiving ticket resolution (exact ID, unique prefix, directory slug, ...).
		ticketRes, err := tickets.Resolve(ctx, ws, settings.Ticket)
		if err != nil {
			return err
		}
		settings.Ticket = ticketRes.TicketID
		targetDocPath = ticketRes.IndexPathAbs
	}

	targetDocPath = filepath.Clean(strings.TrimSpace(targetDocPath))
	if targetDocPath == "" {
		return fmt.Errorf("failed to resolve target document path")
	}

	resolver := paths.NewResolver(paths.ResolverOptions{
		DocsRoot:      ws.Context().Root,
		DocPath:       targetDocPath,
		ConfigDir:     ws.Context().ConfigDir,
		RepoRoot:      ws.Context().RepoRoot,
		WorkspaceRoot: ws.Context().WorkspaceRoot,
	})

	// Optional: collect suggestions (keyed by resolved absolute path; the
	// anchored write/display form lives in suggestionPaths).
	suggestions := map[string]reasonSet{}
	suggestionPaths := map[string]string{}
	// Optional notes from existing documents for the same file
	existingNotes := map[string]map[string]bool{}
	if settings.Suggest {
		// Determine search root: ticket dir if inferable else repo root/docs root.
		// This is used for git/ripgrep heuristics (not for doc scanning).
		searchRoot := ws.Context().RepoRoot
		if searchRoot == "" {
			searchRoot = ws.Context().Root
		}
		if ticketDir == "" && settings.Ticket != "" {
			if inferred := inferTicketDirFromDocPath(ws.Context().Root, targetDocPath); inferred != "" {
				ticketDir = inferred
			}
		}
		if ticketDir != "" {
			searchRoot = ticketDir
		}

		if settings.FromGit {
			// Only from git status (changed files)
			if modified, staged, untracked, err := searchsvc.SuggestFilesFromGitStatus(searchRoot); err == nil {
				for _, f := range modified {
					addSuggestion(suggestions, suggestionPaths, resolver, f, "working tree modified")
				}
				for _, f := range staged {
					addSuggestion(suggestions, suggestionPaths, resolver, f, "staged for commit")
				}
				for _, f := range untracked {
					addSuggestion(suggestions, suggestionPaths, resolver, f, "untracked new file")
				}
			}
		} else {
			// Default heuristic blend: existing docs, git history, ripgrep, git status
			// Use QueryDocs instead of ad-hoc filesystem walking so we share the same skip rules
			// and parsing behavior as the rest of the tool.
			docScope := workspace.Scope{Kind: workspace.ScopeRepo}
			if strings.TrimSpace(settings.Ticket) != "" {
				docScope = workspace.Scope{Kind: workspace.ScopeTicket, TicketID: strings.TrimSpace(settings.Ticket)}
			}
			docRes, err := ws.QueryDocs(ctx, workspace.DocQuery{
				Scope: docScope,
				Options: workspace.DocQueryOptions{
					IncludeErrors:       false,
					IncludeArchivedPath: true,
					IncludeScriptsPath:  true,
					IncludeControlDocs:  true,
					IncludeDiagnostics:  false,
					OrderBy:             workspace.OrderByPath,
				},
			})
			if err == nil {
				for _, h := range docRes.Docs {
					if h.Doc == nil || h.ReadErr != nil {
						continue
					}
					// Topic filter (same behavior as before).
					if len(settings.Topics) > 0 {
						match := false
						for _, ft := range settings.Topics {
							for _, dt := range h.Doc.Topics {
								if strings.EqualFold(strings.TrimSpace(ft), strings.TrimSpace(dt)) {
									match = true
									break
								}
							}
							if match {
								break
							}
						}
						if !match {
							continue
						}
					}
					docResolver := paths.NewResolver(paths.ResolverOptions{
						DocsRoot:      ws.Context().Root,
						DocPath:       h.Path,
						ConfigDir:     ws.Context().ConfigDir,
						RepoRoot:      ws.Context().RepoRoot,
						WorkspaceRoot: ws.Context().WorkspaceRoot,
					})
					for _, rf := range h.Doc.RelatedFiles {
						if strings.TrimSpace(rf.Path) == "" {
							continue
						}
						// Resolve with the source doc's resolver (doc-relative legacy
						// entries), then key/anchor with the target resolver.
						sourceKey := resolveKey(docResolver, rf.Path)
						if sourceKey == "" {
							continue
						}
						key := addSuggestion(suggestions, suggestionPaths, resolver, sourceKey, "referenced by documents")
						if key != "" && strings.TrimSpace(rf.Note) != "" {
							if _, ok := existingNotes[key]; !ok {
								existingNotes[key] = map[string]bool{}
							}
							existingNotes[key][rf.Note] = true
						}
					}
				}
			}

			terms := []string{}
			if settings.Query != "" {
				terms = append(terms, settings.Query)
			}
			terms = append(terms, settings.Topics...)
			if files, err := searchsvc.SuggestFilesFromGit(searchRoot); err == nil {
				for _, f := range files {
					addSuggestion(suggestions, suggestionPaths, resolver, f, "recent commit activity")
				}
			}
			if files, err := searchsvc.SuggestFilesFromRipgrep(searchRoot, terms); err == nil {
				label := fmt.Sprintf("content match: %s", searchsvc.FirstTerm(terms))
				for _, f := range files {
					addSuggestion(suggestions, suggestionPaths, resolver, f, label)
				}
			}
			if modified, staged, untracked, err := searchsvc.SuggestFilesFromGitStatus(searchRoot); err == nil {
				for _, f := range modified {
					addSuggestion(suggestions, suggestionPaths, resolver, f, "working tree modified")
				}
				for _, f := range staged {
					addSuggestion(suggestions, suggestionPaths, resolver, f, "staged for commit")
				}
				for _, f := range untracked {
					addSuggestion(suggestions, suggestionPaths, resolver, f, "untracked new file")
				}
			}
		}

		// Deduplicate suggestions
		var dedup []string
		for f := range suggestions {
			if f != "" {
				dedup = append(dedup, f)
			}
		}
		sort.Slice(dedup, func(i, j int) bool {
			return suggestionDisplay(suggestionPaths, dedup[i]) < suggestionDisplay(suggestionPaths, dedup[j])
		})

		// If we are not applying suggestions, just output them
		if !settings.ApplySuggestions {
			for _, f := range dedup {
				// join reasons
				reasons := make([]string, 0, len(suggestions[f]))
				for r := range suggestions[f] {
					reasons = append(reasons, r)
				}
				sort.Strings(reasons)
				// append notes if known from existing docs
				if notesSet, ok := existingNotes[f]; ok {
					var notes []string
					for n := range notesSet {
						notes = append(notes, n)
					}
					sort.Strings(notes)
					if len(notes) > 0 {
						reasons = append(reasons, "note: "+strings.Join(notes, "; "))
					}
				}
				row := types.NewRow(
					types.MRP("file", suggestionDisplay(suggestionPaths, f)),
					types.MRP("source", "suggested"),
					types.MRP("reason", strings.Join(reasons, "; ")),
				)
				if err := gp.AddRow(ctx, row); err != nil {
					return err
				}
			}
			return nil
		}
	}

	// Read the target document
	doc, content, err := documents.ReadDocumentWithFrontmatter(targetDocPath)
	if err != nil {
		return err
	}

	// Build maps for add/remove with notes retained. Entries are keyed by
	// their resolved absolute path so anchored, legacy and raw forms of the
	// same file collapse into one entry. Existing entries keep their stored
	// path form verbatim (legacy strings are preserved as-is; migration is a
	// separate, explicit step: 'docmgr doctor --fix-anchors').
	current := map[string]models.RelatedFile{}
	for _, rf := range doc.RelatedFiles {
		trimmedPath := strings.TrimSpace(rf.Path)
		if trimmedPath == "" {
			continue
		}
		key := resolveKey(resolver, trimmedPath)
		if key == "" {
			continue
		}
		rf.Path = trimmedPath
		if existing, ok := current[key]; ok {
			if strings.TrimSpace(rf.Note) != "" {
				if merged, changed := appendNote(existing.Note, rf.Note); changed {
					existing.Note = merged
					current[key] = existing
				}
			}
			continue
		}
		current[key] = rf
	}

	// Parse provided file-note mappings
	rawNotes, err := parseFileNotes(settings.FileNotes)
	if err != nil {
		return err
	}

	// Resolve provided paths to identity keys; new entries are written with an
	// explicit anchor (tightest containing anchor rule).
	noteMap := map[string]string{}
	writePaths := map[string]string{}
	for rawPath, note := range rawNotes {
		key := resolveKey(resolver, rawPath)
		if key == "" {
			continue
		}
		if _, ok := writePaths[key]; !ok {
			writePaths[key] = anchoredForWrite(resolver, rawPath)
		}
		if existing, ok := noteMap[key]; ok {
			if merged, changed := appendNote(existing, note); changed {
				noteMap[key] = merged
			}
			continue
		}
		noteMap[key] = note
	}

	// Apply removals
	removedCount := 0
	skippedRemovals := 0
	unchangedNotes := []string{}
	seenUnchanged := map[string]struct{}{}
	for _, rf := range settings.RemoveFiles {
		key := resolveKey(resolver, rf)
		if key == "" {
			key = filepath.ToSlash(strings.TrimSpace(rf))
		}
		if key == "" {
			continue
		}
		if _, ok := current[key]; ok {
			delete(current, key)
			removedCount++
		} else {
			skippedRemovals++
		}
	}

	// Apply additions / updates from file-note mappings only
	addedCount := 0
	updatedCount := 0
	for key, note := range noteMap {
		if rf, ok := current[key]; ok {
			if strings.TrimSpace(note) != "" {
				merged, changed := appendNote(rf.Note, note)
				if changed {
					rf.Note = merged
					current[key] = rf
					updatedCount++
				} else {
					if _, seen := seenUnchanged[key]; !seen {
						unchangedNotes = append(unchangedNotes, rf.Path)
						seenUnchanged[key] = struct{}{}
					}
				}
			}
		} else {
			current[key] = models.RelatedFile{Path: writePaths[key], Note: note}
			addedCount++
		}
	}

	// Apply suggestions if requested
	if settings.Suggest && settings.ApplySuggestions {
		for f, rs := range suggestions {
			if _, ok := current[f]; !ok {
				// Build note from reasons unless an explicit note was provided
				reasonList := make([]string, 0, len(rs))
				for r := range rs {
					reasonList = append(reasonList, r)
				}
				sort.Strings(reasonList)
				note := noteMap[f]
				if note == "" {
					note = strings.Join(reasonList, "; ")
				}
				current[f] = models.RelatedFile{Path: suggestionDisplay(suggestionPaths, f), Note: note}
				addedCount++
			} else if note := noteMap[f]; note != "" {
				rf := current[f]
				if merged, changed := appendNote(rf.Note, note); changed {
					rf.Note = merged
					current[f] = rf
					updatedCount++
				} else {
					if _, seen := seenUnchanged[f]; !seen {
						unchangedNotes = append(unchangedNotes, rf.Path)
						seenUnchanged[f] = struct{}{}
					}
				}
			}
		}
	}

	// When not in suggestion-listing mode, ensure at least one change was requested
	if !settings.Suggest && addedCount == 0 && removedCount == 0 && updatedCount == 0 {
		reasons := []string{}
		if len(noteMap) > 0 {
			if len(unchangedNotes) > 0 {
				reasons = append(reasons, fmt.Sprintf("file-note entries already present for: %s", strings.Join(unchangedNotes, ", ")))
			} else {
				reasons = append(reasons, "file-note entries matched existing notes")
			}
		}
		if len(settings.RemoveFiles) > 0 && skippedRemovals > 0 {
			reasons = append(reasons, "remove targets were not present")
		}
		if len(reasons) == 0 {
			reasons = append(reasons, "no changes requested")
		}
		row := types.NewRow(
			types.MRP("doc", targetDocPath),
			types.MRP("added", addedCount),
			types.MRP("updated", updatedCount),
			types.MRP("removed", removedCount),
			types.MRP("total", len(doc.RelatedFiles)),
			types.MRP("status", "noop"),
			types.MRP("reason", strings.Join(reasons, "; ")),
			types.MRP("unchanged", strings.Join(unchangedNotes, ", ")),
		)
		return gp.AddRow(ctx, row)
	}

	// Serialize back to a slice sorted by the stored path form
	out := make(models.RelatedFiles, 0, len(current))
	for _, rf := range current {
		out = append(out, rf)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	doc.RelatedFiles = out

	// Preserve existing content: rewrite file with updated frontmatter
	if err := documents.WriteDocumentWithFrontmatter(targetDocPath, doc, content, true); err != nil {
		return fmt.Errorf("failed to write document: %w", err)
	}

	row := types.NewRow(
		types.MRP("doc", targetDocPath),
		types.MRP("added", addedCount),
		types.MRP("updated", updatedCount),
		types.MRP("removed", removedCount),
		types.MRP("total", len(doc.RelatedFiles)),
		types.MRP("status", "updated"),
		types.MRP("unchanged", strings.Join(unchangedNotes, ", ")),
	)
	return gp.AddRow(ctx, row)
}

// parseFileNotes parses repeated --file-note values into a path -> note map.
// Each value must use the 'path:note' (or 'path=note') format with a non-empty
// path and note. Anchored paths (repo://..., ws://..., abs:///..., ...) are
// supported: the delimiter search starts after the '<scheme>://' marker so the
// scheme's colon is not mistaken for the path/note separator.
func parseFileNotes(fileNotes []string) (map[string]string, error) {
	notes := map[string]string{}
	for _, m := range fileNotes {
		s := strings.TrimSpace(m)
		if s == "" {
			continue
		}
		searchFrom := 0
		if idx := anchoredSchemePrefixLen(s); idx > 0 {
			searchFrom = idx
		}
		i := strings.IndexAny(s[searchFrom:], ":=")
		if i < 0 {
			return nil, fmt.Errorf("malformed --file-note value %q: expected 'path:note' (or 'path=note')", m)
		}
		i += searchFrom
		key := strings.TrimSpace(s[:i])
		val := strings.TrimSpace(s[i+1:])
		if key == "" {
			return nil, fmt.Errorf("malformed --file-note value %q: empty path (expected 'path:note')", m)
		}
		if val == "" {
			return nil, fmt.Errorf("--file-note requires a non-empty note for %s (use 'path:reason')", key)
		}
		notes[key] = val
	}
	return notes, nil
}

// anchoredSchemePrefixLen returns the length of a known anchor-scheme prefix
// ("repo://", "ws://", "docs://", "doc://", "abs://") at the start of s, or 0
// when s does not start with one.
func anchoredSchemePrefixLen(s string) int {
	i := strings.Index(s, "://")
	if i <= 0 {
		return 0
	}
	switch paths.Scheme(strings.ToLower(s[:i])) {
	case paths.SchemeRepo, paths.SchemeWs, paths.SchemeDocs, paths.SchemeDoc, paths.SchemeAbs:
		return i + len("://")
	case paths.SchemeLegacy:
		return 0
	default:
		return 0
	}
}

func appendNote(existing, addition string) (string, bool) {
	addition = strings.TrimSpace(addition)
	if addition == "" {
		return existing, false
	}
	if existing == "" {
		return addition, true
	}
	for _, line := range strings.Split(existing, "\n") {
		if strings.TrimSpace(line) == addition {
			return existing, false
		}
	}
	if strings.HasSuffix(existing, "\n") {
		return existing + addition, true
	}
	return existing + "\n" + addition, true
}

// resolveKey returns the identity key used to deduplicate and remove
// RelatedFiles entries: the resolved absolute path (one resolver for anchored
// and legacy forms alike), falling back to the cleaned raw string when the
// path cannot be resolved against any anchor.
func resolveKey(resolver *paths.Resolver, raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if resolver == nil {
		return filepath.ToSlash(raw)
	}
	n := resolver.Resolve(raw)
	if strings.TrimSpace(n.Abs) != "" {
		return filepath.ToSlash(strings.TrimSpace(n.Abs))
	}
	if strings.TrimSpace(n.Canonical) != "" {
		return strings.TrimSpace(n.Canonical)
	}
	return filepath.ToSlash(raw)
}

// anchoredForWrite returns the anchored path string persisted for a newly
// related file (design doc DOCMGR-200 §8.1): resolve the input to an absolute
// path, then stamp the tightest containing anchor (repo:// > ws://<member> >
// docs:// > abs://). Repo-escaping ../ chains are never emitted. Unresolvable
// inputs fall back to the cleaned raw string (legacy form).
func anchoredForWrite(resolver *paths.Resolver, raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if resolver == nil {
		return filepath.ToSlash(raw)
	}
	n := resolver.Resolve(raw)
	abs := strings.TrimSpace(n.Abs)
	if abs == "" {
		if strings.TrimSpace(n.Canonical) != "" {
			return strings.TrimSpace(n.Canonical)
		}
		return filepath.ToSlash(raw)
	}
	return resolver.AnchoredFor(abs).String()
}

// suggestionDisplay returns the anchored display/write form of a suggestion key.
func suggestionDisplay(pathsByKey map[string]string, key string) string {
	if p := strings.TrimSpace(pathsByKey[key]); p != "" {
		return p
	}
	return key
}

func addSuggestion(out map[string]reasonSet, pathsByKey map[string]string, resolver *paths.Resolver, rawPath, reason string) string {
	key := resolveKey(resolver, rawPath)
	if key == "" {
		return ""
	}
	if _, ok := out[key]; !ok {
		out[key] = reasonSet{}
	}
	if pathsByKey != nil {
		if _, ok := pathsByKey[key]; !ok {
			pathsByKey[key] = anchoredForWrite(resolver, rawPath)
		}
	}
	if reason != "" {
		out[key][reason] = true
	}
	return key
}

// inferTicketDirFromDocPath best-effort returns the ticket directory for a doc under docsRoot.
//
// Expected docs layout:
//
//	<docsRoot>/<YYYY>/<MM>/<DD>/<TICKET--slug>/...
func inferTicketDirFromDocPath(docsRoot string, absDocPath string) string {
	docsRoot = filepath.Clean(strings.TrimSpace(docsRoot))
	absDocPath = filepath.Clean(strings.TrimSpace(absDocPath))
	if docsRoot == "" || absDocPath == "" {
		return ""
	}
	rel, err := filepath.Rel(docsRoot, absDocPath)
	if err != nil {
		return ""
	}
	rel = filepath.Clean(rel)
	if rel == "." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." {
		return ""
	}
	parts := strings.Split(rel, string(filepath.Separator))
	if len(parts) < 4 {
		return ""
	}
	ticketDir := filepath.Join(docsRoot, parts[0], parts[1], parts[2], parts[3])
	return ticketDir
}

var _ cmds.GlazeCommand = &RelateCommand{}

type relateRowCollector struct {
	rows []types.Row
}

func (c *relateRowCollector) AddRow(ctx context.Context, row types.Row) error {
	c.rows = append(c.rows, row)
	return nil
}

func (c *relateRowCollector) Close(ctx context.Context) error {
	return nil
}

func (c *RelateCommand) Run(
	ctx context.Context,
	parsedValues *values.Values,
) error {
	collector := &relateRowCollector{}
	if err := c.RunIntoGlazeProcessor(ctx, parsedValues, collector); err != nil {
		return err
	}

	if len(collector.rows) == 0 {
		fmt.Println("No related files updated or suggested.")
		return nil
	}

	first := collector.rows[0]
	if _, ok := first.Get("doc"); ok {
		docPathVal, _ := first.Get("doc")
		docPath := displayPathForCwd(fmt.Sprint(docPathVal))
		statusVal, _ := first.Get("status")
		status := fmt.Sprint(statusVal)
		unchangedVal, _ := first.Get("unchanged")
		unchanged := strings.TrimSpace(fmt.Sprint(unchangedVal))
		if status == "noop" {
			reasonVal, _ := first.Get("reason")
			reason := strings.TrimSpace(fmt.Sprint(reasonVal))
			if reason == "" || reason == "<nil>" {
				fmt.Printf("no related file changes for %v\n", docPath)
			} else {
				fmt.Printf("no related file changes for %v (%s)\n", docPath, reason)
			}
			return nil
		}
		added, _ := first.Get("added")
		updated, _ := first.Get("updated")
		removed, _ := first.Get("removed")
		touched := 0
		for _, v := range []interface{}{added, updated, removed} {
			if n, ok := v.(int); ok {
				touched += n
			}
		}
		fmt.Printf("related %d file(s) to %v (added %v, updated %v, removed %v)\n", touched, docPath, added, updated, removed)
		if VerboseEnabled() && unchanged != "" && unchanged != "<nil>" {
			fmt.Printf("- Unchanged (already present with same note): %s\n", unchanged)
		}
		return nil
	}

	fmt.Println("Suggested related files:")
	for _, row := range collector.rows {
		file, _ := row.Get("file")
		reason, _ := row.Get("reason")
		fmt.Printf("- %v — %v\n", file, reason)
	}

	return nil
}

var _ cmds.BareCommand = &RelateCommand{}
