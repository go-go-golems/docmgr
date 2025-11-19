package commands

import (
	"context"
	"fmt"
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

type ChangelogUpdateCommand struct{ *cmds.CommandDescription }

type ChangelogUpdateSettings struct {
	Ticket        string `glazed.parameter:"ticket"`
	Root          string `glazed.parameter:"root"`
	ChangelogFile string `glazed.parameter:"changelog-file"`
	Title         string `glazed.parameter:"title"`
	Entry         string `glazed.parameter:"entry"`
	// Deprecated: kept only to emit a friendly migration error if provided
	Files            []string `glazed.parameter:"files"`
	FileNotes        []string `glazed.parameter:"file-note"`
	Suggest          bool     `glazed.parameter:"suggest"`
	ApplySuggestions bool     `glazed.parameter:"apply-suggestions"`
	Query            string   `glazed.parameter:"query"`
	Topics           []string `glazed.parameter:"topics"`
}

func NewChangelogUpdateCommand() (*ChangelogUpdateCommand, error) {
	cmd := cmds.NewCommandDescription(
		"update",
		cmds.WithShort("Append an entry to changelog.md for a ticket"),
		cmds.WithLong(`Append a dated changelog entry with optional title, message, and related files.

Examples:
  # Append an entry with a message
  docmgr changelog update --ticket MEN-4242 --entry "Normalized chat API paths"

  # Include related files with notes
  docmgr changelog update --ticket MEN-4242 \
    --file-note "backend/chat/api/register.go:Source of path normalization" \
    --file-note "web/src/store/api/chatApi.ts=Frontend integration"

  # Use suggestions (git + ripgrep + existing docs); print suggestions only
  docmgr changelog update --ticket MEN-4242 --suggest --query WebSocket

  # Apply suggestions and write them into the entry
  docmgr changelog update --ticket MEN-4242 --suggest --apply-suggestions --query WebSocket
`),
		cmds.WithFlags(
			parameters.NewParameterDefinition("ticket", parameters.ParameterTypeString, parameters.WithHelp("Ticket identifier for the target workspace"), parameters.WithDefault("")),
			parameters.NewParameterDefinition("root", parameters.ParameterTypeString, parameters.WithHelp("Root directory for docs"), parameters.WithDefault("ttmp")),
			parameters.NewParameterDefinition("changelog-file", parameters.ParameterTypeString, parameters.WithHelp("Path to changelog.md (overrides --ticket)"), parameters.WithDefault("")),
			parameters.NewParameterDefinition("title", parameters.ParameterTypeString, parameters.WithHelp("Optional entry title"), parameters.WithDefault("")),
			parameters.NewParameterDefinition("entry", parameters.ParameterTypeString, parameters.WithHelp("Entry text to append"), parameters.WithDefault("")),
			parameters.NewParameterDefinition("files", parameters.ParameterTypeStringList, parameters.WithHelp("DEPRECATED (removed) — use repeated --file-note 'path:note'"), parameters.WithDefault([]string{})),
			parameters.NewParameterDefinition("file-note", parameters.ParameterTypeStringList, parameters.WithHelp("Repeatable path-to-note mapping (path:note or path=note)"), parameters.WithDefault([]string{})),
			parameters.NewParameterDefinition("suggest", parameters.ParameterTypeBool, parameters.WithHelp("Suggest related files using heuristics (git + ripgrep + existing docs)"), parameters.WithDefault(false)),
			parameters.NewParameterDefinition("apply-suggestions", parameters.ParameterTypeBool, parameters.WithHelp("Apply suggested files to this changelog entry"), parameters.WithDefault(false)),
			parameters.NewParameterDefinition("query", parameters.ParameterTypeString, parameters.WithHelp("Optional query to seed suggestions"), parameters.WithDefault("")),
			parameters.NewParameterDefinition("topics", parameters.ParameterTypeStringList, parameters.WithHelp("Topics to seed suggestions (comma-separated)"), parameters.WithDefault([]string{})),
		),
	)
	return &ChangelogUpdateCommand{CommandDescription: cmd}, nil
}

func (c *ChangelogUpdateCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	pl *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &ChangelogUpdateSettings{}
	if err := pl.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Enforce deprecation: --files is no longer supported for changelog updates.
	if len(s.Files) > 0 {
		return fmt.Errorf("--files has been removed from 'docmgr changelog update'. Use repeated --file-note 'path:note' instead. Example: docmgr changelog update --file-note 'a/b.go:reason' --file-note 'c/d.ts:reason'")
	}

	// Resolve changelog path
	var changelogPath string
	if s.ChangelogFile != "" {
		changelogPath = s.ChangelogFile
	} else {
		s.Root = workspace.ResolveRoot(s.Root)
		if s.Ticket == "" {
			return fmt.Errorf("must specify --ticket or --changelog-file")
		}
		td, err := findTicketDirectory(s.Root, s.Ticket)
		if err != nil {
			return fmt.Errorf("failed to find ticket directory: %w", err)
		}
		changelogPath = filepath.Join(td, "changelog.md")
	}

	// Collect suggestions if requested
	type reasonSet map[string]bool
	suggestions := map[string]reasonSet{}
	existingNotes := map[string]map[string]bool{}

	if s.Suggest {
		// Determine search root
		searchRoot := s.Root
		if s.ChangelogFile == "" && s.Ticket != "" {
			if td, err := findTicketDirectory(s.Root, s.Ticket); err == nil && td != "" {
				searchRoot = td
			}
		}

		// From existing docs' RelatedFiles within the search root
		existing := map[string]bool{}
		_ = filepath.Walk(searchRoot, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
				return nil
			}
			doc, err := readDocumentFrontmatter(path)
			if err != nil {
				return nil
			}
			// Filter by topics if provided
			if len(s.Topics) > 0 {
				match := false
				for _, ft := range s.Topics {
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

		// Git history
		terms := []string{}
		if s.Query != "" {
			terms = append(terms, s.Query)
		}
		terms = append(terms, s.Topics...)
		if files, err := suggestFilesFromGit(searchRoot, terms); err == nil {
			for _, f := range files {
				if _, ok := suggestions[f]; !ok {
					suggestions[f] = reasonSet{}
				}
				suggestions[f]["recent commit activity"] = true
			}
		}

		// ripgrep / grep
		if files, err := suggestFilesFromRipgrep(searchRoot, terms); err == nil {
			for _, f := range files {
				if _, ok := suggestions[f]; !ok {
					suggestions[f] = reasonSet{}
				}
				suggestions[f][fmt.Sprintf("content match: %s", firstTerm(terms))] = true
			}
		}

		// git status
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

		// If not applying and no explicit files provided, print suggestions and exit
		if !s.ApplySuggestions && len(s.Files) == 0 && len(s.FileNotes) == 0 {
			var keys []string
			for f := range suggestions {
				if f != "" {
					keys = append(keys, f)
				}
			}
			sort.Strings(keys)
			for _, f := range keys {
				// merge reasons and attach known notes
				reasons := make([]string, 0, len(suggestions[f]))
				for r := range suggestions[f] {
					reasons = append(reasons, r)
				}
				sort.Strings(reasons)
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
				row := types.NewRow(types.MRP("file", f), types.MRP("source", "suggested"), types.MRP("reason", strings.Join(reasons, "; ")))
				if err := gp.AddRow(ctx, row); err != nil {
					return err
				}
			}
			return nil
		}
	}

	// Build file->note map from provided file-notes
	noteMap := map[string]string{}
	for _, m := range s.FileNotes {
		str := strings.TrimSpace(m)
		if str == "" {
			continue
		}
		var key, val string
		if i := strings.IndexAny(str, ":="); i >= 0 {
			key = strings.TrimSpace(str[:i])
			val = strings.TrimSpace(str[i+1:])
		} else {
			continue
		}
		if key != "" {
			noteMap[key] = val
		}
	}

	// Validate that provided file-note mappings contain non-empty notes
	for p, n := range noteMap {
		if strings.TrimSpace(n) == "" {
			return fmt.Errorf("--file-note requires a non-empty note for %s (use 'path:reason')", p)
		}
	}

	// Build final list of related files to include in the entry from noteMap only
	final := map[string]string{} // path -> note
	for p, n := range noteMap {
		pp := strings.TrimSpace(p)
		if pp != "" {
			final[pp] = n
		}
	}
	if s.Suggest && s.ApplySuggestions {
		for f, rs := range suggestions {
			// Build note from reasons unless overridden
			var reasons []string
			for r := range rs {
				reasons = append(reasons, r)
			}
			sort.Strings(reasons)
			note := noteMap[f]
			if note == "" {
				note = strings.Join(reasons, "; ")
			}
			final[f] = note
		}
	}

	// Ensure changelog file exists; create if missing with header
	if _, err := os.Stat(changelogPath); os.IsNotExist(err) {
		_ = os.MkdirAll(filepath.Dir(changelogPath), 0755)
		// minimal header
		_ = os.WriteFile(changelogPath, []byte("# Changelog\n\n"), 0644)
	}

	// Compose entry
	today := time.Now().Format("2006-01-02")
	heading := "## " + today
	if strings.TrimSpace(s.Title) != "" {
		heading += " - " + s.Title
	}

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(heading)
	sb.WriteString("\n\n")
	if strings.TrimSpace(s.Entry) != "" {
		sb.WriteString(s.Entry)
		sb.WriteString("\n\n")
	}
	if len(final) > 0 {
		sb.WriteString("### Related Files\n\n")
		// stable order
		var files []string
		for f := range final {
			files = append(files, f)
		}
		sort.Strings(files)
		for _, f := range files {
			note := strings.TrimSpace(final[f])
			if note != "" {
				sb.WriteString("- " + f + " — " + note + "\n")
			} else {
				sb.WriteString("- " + f + "\n")
			}
		}
		sb.WriteString("\n")
	}

	// Append to changelog
	fp, err := os.OpenFile(changelogPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open changelog.md: %w", err)
	}
	defer func() { _ = fp.Close() }()
	if _, err := fp.WriteString(sb.String()); err != nil {
		return fmt.Errorf("failed to write changelog entry: %w", err)
	}

	row := types.NewRow(
		types.MRP("file", changelogPath),
		types.MRP("status", "updated"),
		types.MRP("date", today),
		types.MRP("files_count", len(final)),
	)
	return gp.AddRow(ctx, row)
}

var _ cmds.GlazeCommand = &ChangelogUpdateCommand{}

// Implement BareCommand for human-friendly output with reminders
func (c *ChangelogUpdateCommand) Run(
	ctx context.Context,
	pl *layers.ParsedLayers,
) error {
	s := &ChangelogUpdateSettings{}
	if err := pl.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Enforce deprecation: --files is no longer supported for changelog updates.
	if len(s.Files) > 0 {
		return fmt.Errorf("--files has been removed from 'docmgr changelog update'. Use repeated --file-note 'path:note' instead. Example: docmgr changelog update --file-note 'a/b.go:reason' --file-note 'c/d.ts:reason'")
	}

	// Resolve changelog path
	var changelogPath string
	if s.ChangelogFile != "" {
		changelogPath = s.ChangelogFile
	} else {
		s.Root = workspace.ResolveRoot(s.Root)
		if s.Ticket == "" {
			return fmt.Errorf("must specify --ticket or --changelog-file")
		}
		td, err := findTicketDirectory(s.Root, s.Ticket)
		if err != nil {
			return fmt.Errorf("failed to find ticket directory: %w", err)
		}
		changelogPath = filepath.Join(td, "changelog.md")
	}

	// Suggestions: if requested without apply/files, print suggestions and exit
	type reasonSet map[string]bool
	suggestions := map[string]reasonSet{}
	if s.Suggest && !s.ApplySuggestions && len(s.Files) == 0 && len(s.FileNotes) == 0 {
		searchRoot := s.Root
		if s.ChangelogFile == "" && s.Ticket != "" {
			if td, err := findTicketDirectory(s.Root, s.Ticket); err == nil && td != "" {
				searchRoot = td
			}
		}
		// Minimal suggestion pass: from git status only to keep output concise
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
		var files []string
		for f := range suggestions {
			files = append(files, f)
		}
		sort.Strings(files)
		for _, f := range files {
			var reasons []string
			for r := range suggestions[f] {
				reasons = append(reasons, r)
			}
			sort.Strings(reasons)
			fmt.Printf("suggested: %s (%s)\n", f, strings.Join(reasons, "; "))
		}
		return nil
	}

	// Build file-note map
	noteMap := map[string]string{}
	for _, m := range s.FileNotes {
		str := strings.TrimSpace(m)
		if str == "" {
			continue
		}
		var key, val string
		if i := strings.IndexAny(str, ":="); i >= 0 {
			key = strings.TrimSpace(str[:i])
			val = strings.TrimSpace(str[i+1:])
		}
		if key != "" {
			noteMap[key] = val
		}
	}

	// Validate notes are present for provided mappings
	for p, n := range noteMap {
		if strings.TrimSpace(n) == "" {
			return fmt.Errorf("--file-note requires a non-empty note for %s (use 'path:reason')", p)
		}
	}

	// Final list (from noteMap only)
	final := map[string]string{}
	for p, n := range noteMap {
		pp := strings.TrimSpace(p)
		if pp != "" {
			final[pp] = n
		}
	}

	// Ensure file exists
	if _, err := os.Stat(changelogPath); os.IsNotExist(err) {
		_ = os.MkdirAll(filepath.Dir(changelogPath), 0755)
		_ = os.WriteFile(changelogPath, []byte("# Changelog\n\n"), 0644)
	}

	today := time.Now().Format("2006-01-02")
	heading := "## " + today
	if strings.TrimSpace(s.Title) != "" {
		heading += " - " + s.Title
	}

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(heading)
	sb.WriteString("\n\n")
	if strings.TrimSpace(s.Entry) != "" {
		sb.WriteString(s.Entry)
		sb.WriteString("\n\n")
	}
	if len(final) > 0 {
		sb.WriteString("### Related Files\n\n")
		var files []string
		for f := range final {
			files = append(files, f)
		}
		sort.Strings(files)
		for _, f := range files {
			note := strings.TrimSpace(final[f])
			if note != "" {
				sb.WriteString("- " + f + " — " + note + "\n")
			} else {
				sb.WriteString("- " + f + "\n")
			}
		}
		sb.WriteString("\n")
	}

	fp, err := os.OpenFile(changelogPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open changelog.md: %w", err)
	}
	defer func() { _ = fp.Close() }()
	if _, err := fp.WriteString(sb.String()); err != nil {
		return fmt.Errorf("failed to write changelog entry: %w", err)
	}

	fmt.Printf("Changelog updated: %s\n", changelogPath)
	fmt.Println("Reminder: update the ticket index (docmgr relate/meta) and refresh file relationships in any impacted docs if needed.")
	return nil
}

var _ cmds.BareCommand = &ChangelogUpdateCommand{}
