package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-go-golems/docmgr/pkg/models"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

// RelateCommand updates RelatedFiles metadata and can suggest files to relate
type RelateCommand struct {
	*cmds.CommandDescription
}

type RelateSettings struct {
	Ticket           string   `glazed.parameter:"ticket"`
	Doc              string   `glazed.parameter:"doc"`
	Files            []string `glazed.parameter:"files"`
	RemoveFiles      []string `glazed.parameter:"remove-files"`
	FileNotes        []string `glazed.parameter:"file-note"`
	Suggest          bool     `glazed.parameter:"suggest"`
	ApplySuggestions bool     `glazed.parameter:"apply-suggestions"`
	FromGit          bool     `glazed.parameter:"from-git"`
	Query            string   `glazed.parameter:"query"`
	Topics           []string `glazed.parameter:"topics"`
	Root             string   `glazed.parameter:"root"`
}

func NewRelateCommand() (*RelateCommand, error) {
	return &RelateCommand{
		CommandDescription: cmds.NewCommandDescription(
			"relate",
			cmds.WithShort("Relate code files to a document or ticket"),
			cmds.WithLong(`Update RelatedFiles in a document's frontmatter or the ticket index.

Examples:
  # Relate files to the ticket index
  docmgr relate --ticket MEN-4242 --files backend/chat/api/register.go,web/src/store/api/chatApi.ts

  # Relate files to a specific document
  docmgr relate --doc ttmp/YYYY/MM/DD/MEN-4242-.../design/path-normalization-strategy.md --files backend/chat/ws/manager.go

  # Suggest files using heuristics (git + ripgrep + existing RelatedFiles)
  docmgr relate --ticket MEN-4242 --suggest --query WebSocket --topics chat,backend

  # Apply suggestions automatically to the ticket index
  docmgr relate --ticket MEN-4242 --suggest --apply-suggestions --query WebSocket
`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"ticket",
					parameters.ParameterTypeString,
					parameters.WithHelp("Ticket identifier (updates ticket index when --doc not provided)"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"doc",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path to a specific document to update"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"files",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Comma-separated list of files to add to RelatedFiles"),
					parameters.WithDefault([]string{}),
				),
				parameters.NewParameterDefinition(
					"remove-files",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Comma-separated list of files to remove from RelatedFiles"),
					parameters.WithDefault([]string{}),
				),
				parameters.NewParameterDefinition(
					"file-note",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Repeatable path-to-note mapping (format: path:note or path=note)"),
					parameters.WithDefault([]string{}),
				),
				parameters.NewParameterDefinition(
					"suggest",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Suggest related files using heuristics (git + ripgrep + existing docs)"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"apply-suggestions",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Apply suggested files to the target document (requires --suggest)"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"from-git",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Limit suggestions to changed files from git status (modified, staged, untracked)"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"query",
					parameters.ParameterTypeString,
					parameters.WithHelp("Optional query to seed suggestions (e.g., a keyword)"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"topics",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Topics to seed suggestions (comma-separated)"),
					parameters.WithDefault([]string{}),
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

func (c *RelateCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &RelateSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Apply config root if present
	settings.Root = ResolveRoot(settings.Root)

	// Resolve target document path
	var targetDocPath string
	var ticketDir string
	var err error

	if settings.Doc != "" {
		targetDocPath = settings.Doc
	} else {
		if settings.Ticket == "" {
			return fmt.Errorf("must specify either --doc or --ticket")
		}
		ticketDir, err = findTicketDirectory(settings.Root, settings.Ticket)
		if err != nil {
			return fmt.Errorf("failed to find ticket directory: %w", err)
		}
		targetDocPath = filepath.Join(ticketDir, "index.md")
	}

	// Optional: collect suggestions
	type reasonSet map[string]bool
	suggestions := map[string]reasonSet{}
	// Optional notes from existing documents for the same file
	existingNotes := map[string]map[string]bool{}
	if settings.Suggest {
		// Determine search root: ticket dir if available else repo root
		searchRoot := settings.Root
		if ticketDir == "" && settings.Ticket != "" {
			ticketDir, _ = findTicketDirectory(settings.Root, settings.Ticket)
		}
		if ticketDir != "" {
			searchRoot = ticketDir
		}

		if settings.FromGit {
			// Only from git status (changed files)
			if modified, staged, untracked, err := suggestFilesFromGitStatus(searchRoot); err == nil {
				for _, f := range modified {
					if _, ok := suggestions[f]; !ok {
						suggestions[f] = reasonSet{}
					}
					suggestions[f]["working tree modified"] = true
				}
				for _, f := range staged {
					if _, ok := suggestions[f]; !ok {
						suggestions[f] = reasonSet{}
					}
					suggestions[f]["staged for commit"] = true
				}
				for _, f := range untracked {
					if _, ok := suggestions[f]; !ok {
						suggestions[f] = reasonSet{}
					}
					suggestions[f]["untracked new file"] = true
				}
			}
		} else {
			// Default heuristic blend: existing docs, git history, ripgrep, git status
			existing := map[string]bool{}
			_ = filepath.Walk(searchRoot, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
					return nil
				}
				doc, err := readDocumentFrontmatter(path)
				if err != nil {
					return nil
				}
				// If topics provided, filter
				if len(settings.Topics) > 0 {
					match := false
					for _, ft := range settings.Topics {
						for _, dt := range doc.Topics {
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
						return nil
					}
				}
				for _, rf := range doc.RelatedFiles {
					if rf.Path != "" {
						existing[rf.Path] = true
						if rf.Note != "" {
							if _, ok := existingNotes[rf.Path]; !ok {
								existingNotes[rf.Path] = map[string]bool{}
							}
							existingNotes[rf.Path][rf.Note] = true
						}
					}
				}
				return nil
			})

			for f := range existing {
				if _, ok := suggestions[f]; !ok {
					suggestions[f] = reasonSet{}
				}
				suggestions[f]["referenced by documents"] = true
			}

			terms := []string{}
			if settings.Query != "" {
				terms = append(terms, settings.Query)
			}
			terms = append(terms, settings.Topics...)
			if files, err := suggestFilesFromGit(searchRoot, terms); err == nil {
				for _, f := range files {
					if _, ok := suggestions[f]; !ok {
						suggestions[f] = reasonSet{}
					}
					suggestions[f]["recent commit activity"] = true
				}
			}
			if files, err := suggestFilesFromRipgrep(searchRoot, terms); err == nil {
				for _, f := range files {
					if _, ok := suggestions[f]; !ok {
						suggestions[f] = reasonSet{}
					}
					suggestions[f][fmt.Sprintf("content match: %s", firstTerm(terms))] = true
				}
			}
			if modified, staged, untracked, err := suggestFilesFromGitStatus(searchRoot); err == nil {
				for _, f := range modified {
					if _, ok := suggestions[f]; !ok {
						suggestions[f] = reasonSet{}
					}
					suggestions[f]["working tree modified"] = true
				}
				for _, f := range staged {
					if _, ok := suggestions[f]; !ok {
						suggestions[f] = reasonSet{}
					}
					suggestions[f]["staged for commit"] = true
				}
				for _, f := range untracked {
					if _, ok := suggestions[f]; !ok {
						suggestions[f] = reasonSet{}
					}
					suggestions[f]["untracked new file"] = true
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
		sort.Strings(dedup)

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
					types.MRP("file", f),
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
	doc, err := readDocumentFrontmatter(targetDocPath)
	if err != nil {
		return fmt.Errorf("failed to read document frontmatter: %w", err)
	}

	// Build maps for add/remove with notes retained
	current := map[string]models.RelatedFile{}
	for _, rf := range doc.RelatedFiles {
		if rf.Path != "" {
			current[rf.Path] = rf
		}
	}

	// Parse provided file-note mappings
	noteMap := map[string]string{}
	for _, m := range settings.FileNotes {
		s := strings.TrimSpace(m)
		if s == "" {
			continue
		}
		var key, val string
		if i := strings.IndexAny(s, ":="); i >= 0 {
			key = strings.TrimSpace(s[:i])
			val = strings.TrimSpace(s[i+1:])
		} else {
			// No delimiter, skip
			continue
		}
		if key != "" {
			noteMap[key] = val
		}
	}

	// Apply removals
	removedCount := 0
	for _, rf := range settings.RemoveFiles {
		rf = strings.TrimSpace(rf)
		if rf == "" {
			continue
		}
		if _, ok := current[rf]; ok {
			delete(current, rf)
			removedCount++
		}
	}

	// Apply additions
	addedCount := 0
	for _, af := range settings.Files {
		af = strings.TrimSpace(af)
		if af == "" {
			continue
		}
		if _, ok := current[af]; !ok {
			current[af] = models.RelatedFile{Path: af, Note: noteMap[af]}
			addedCount++
		} else if note, ok := noteMap[af]; ok && note != "" {
			rf := current[af]
			rf.Note = note
			current[af] = rf
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
				current[f] = models.RelatedFile{Path: f, Note: note}
				addedCount++
			} else if note := noteMap[f]; note != "" {
				rf := current[f]
				rf.Note = note
				current[f] = rf
			}
		}
	}

	// Serialize back to sorted slice
	keys := make([]string, 0, len(current))
	for f := range current {
		keys = append(keys, f)
	}
	sort.Strings(keys)
	out := make(models.RelatedFiles, 0, len(keys))
	for _, k := range keys {
		out = append(out, current[k])
	}
	doc.RelatedFiles = out

	// Preserve existing content: rewrite file with updated frontmatter
	// We need the content; use readDocumentWithContent
	_, content, err := readDocumentWithContent(targetDocPath)
	if err != nil {
		return fmt.Errorf("failed to read document content: %w", err)
	}
	if err := writeDocumentWithFrontmatter(targetDocPath, doc, content, true); err != nil {
		return fmt.Errorf("failed to write document: %w", err)
	}

	row := types.NewRow(
		types.MRP("doc", targetDocPath),
		types.MRP("added", addedCount),
		types.MRP("removed", removedCount),
		types.MRP("total", len(doc.RelatedFiles)),
		types.MRP("status", "updated"),
	)
	return gp.AddRow(ctx, row)
}

var _ cmds.GlazeCommand = &RelateCommand{}
