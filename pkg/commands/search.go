package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

// SearchCommand searches documents by content and metadata
type SearchCommand struct {
	*cmds.CommandDescription
}

// SearchSettings holds the parameters for the search command
type SearchSettings struct {
	Query          string   `glazed.parameter:"query"`
	Ticket         string   `glazed.parameter:"ticket"`
	Topics         []string `glazed.parameter:"topics"`
	DocType        string   `glazed.parameter:"doc-type"`
	Status         string   `glazed.parameter:"status"`
	Files          bool     `glazed.parameter:"files"`
	File           string   `glazed.parameter:"file"`
	Dir            string   `glazed.parameter:"dir"`
	ExternalSource string   `glazed.parameter:"external-source"`
	Since          string   `glazed.parameter:"since"`
	Until          string   `glazed.parameter:"until"`
	CreatedSince   string   `glazed.parameter:"created-since"`
	UpdatedSince   string   `glazed.parameter:"updated-since"`
	Root           string   `glazed.parameter:"root"`
}

func NewSearchCommand() (*SearchCommand, error) {
	return &SearchCommand{
		CommandDescription: cmds.NewCommandDescription(
			"search",
			cmds.WithShort("Search documents by content and metadata"),
			cmds.WithLong(`Search documents by full-text content and metadata filters.

The search command supports:
- Full-text search across document content
- Metadata filtering (ticket, topics, doc-type, status)
- File suggestions using heuristics (--files flag)
- Reverse lookup: find docs for a file/directory (--file, --dir)
- External source search (--external-source)
- Date range filtering (--since, --until, --created-since, --updated-since)

Example:
  docmgr search "authentication"
  docmgr search "API" --ticket MEN-3475
  docmgr search "database" --topics backend --doc-type design-doc
  docmgr search --ticket MEN-3475 --topics chat --files
  docmgr search --file pkg/commands/add.go
  docmgr search --dir pkg/commands/
  docmgr search --external-source "https://github.com/..."
  docmgr search --updated-since "2 weeks ago"
  docmgr search --created-since "2025-01-01" --until "2025-01-31"
`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"query",
					parameters.ParameterTypeString,
					parameters.WithHelp("Search query text (searches document content)"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"ticket",
					parameters.ParameterTypeString,
					parameters.WithHelp("Filter by ticket identifier"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"topics",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Filter by topics (comma-separated, matches any)"),
					parameters.WithDefault([]string{}),
				),
				parameters.NewParameterDefinition(
					"doc-type",
					parameters.ParameterTypeString,
					parameters.WithHelp("Filter by document type"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"status",
					parameters.ParameterTypeString,
					parameters.WithHelp("Filter by status"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"files",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Suggest related files using heuristics (git + ripgrep)"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"file",
					parameters.ParameterTypeString,
					parameters.WithHelp("Find documents that reference this file path"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"dir",
					parameters.ParameterTypeString,
					parameters.WithHelp("Find documents in this directory or referencing files in it"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"external-source",
					parameters.ParameterTypeString,
					parameters.WithHelp("Find documents that reference this external source URL"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"since",
					parameters.ParameterTypeString,
					parameters.WithHelp("Find documents updated since this date (relative: '2 weeks ago', 'last month', or absolute: '2025-01-01')"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"until",
					parameters.ParameterTypeString,
					parameters.WithHelp("Find documents updated until this date (relative: '2 weeks ago', 'last month', or absolute: '2025-01-01')"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"created-since",
					parameters.ParameterTypeString,
					parameters.WithHelp("Find documents created since this date (relative: '2 weeks ago', 'last month', or absolute: '2025-01-01')"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"updated-since",
					parameters.ParameterTypeString,
					parameters.WithHelp("Find documents updated since this date (relative: '2 weeks ago', 'last month', or absolute: '2025-01-01')"),
					parameters.WithDefault(""),
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

func (c *SearchCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &SearchSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Apply config root if present
	settings.Root = workspace.ResolveRoot(settings.Root)

	// If --files flag is set, suggest files instead of searching documents
	if settings.Files {
		return c.suggestFiles(ctx, settings, gp)
	}

	// Validate that we have at least a query or some filters
	if settings.Query == "" && settings.Ticket == "" && len(settings.Topics) == 0 && settings.DocType == "" && settings.Status == "" &&
		settings.File == "" && settings.Dir == "" && settings.ExternalSource == "" &&
		settings.Since == "" && settings.Until == "" && settings.CreatedSince == "" && settings.UpdatedSince == "" {
		return fmt.Errorf("must provide at least a query or filter")
	}

	// Parse date filters
	sinceTime, err := parseDate(settings.Since)
	if err != nil {
		return fmt.Errorf("invalid --since date: %w", err)
	}
	untilTime, err2 := parseDate(settings.Until)
	if err2 != nil {
		return fmt.Errorf("invalid --until date: %w", err2)
	}
	createdSinceTime, err3 := parseDate(settings.CreatedSince)
	if err3 != nil {
		return fmt.Errorf("invalid --created-since date: %w", err3)
	}
	updatedSinceTime, err4 := parseDate(settings.UpdatedSince)
	if err4 != nil {
		return fmt.Errorf("invalid --updated-since date: %w", err4)
	}

	if _, err := os.Stat(settings.Root); os.IsNotExist(err) {
		return fmt.Errorf("root directory does not exist: %s", settings.Root)
	}

	queryLower := strings.ToLower(settings.Query)

	// Search all markdown files
	err = filepath.Walk(settings.Root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Skip templates and guidelines directories
		if strings.Contains(path, "/_templates/") || strings.Contains(path, "/_guidelines/") {
			return nil
		}

		// Read document
		doc, content, err := readDocumentWithContent(path)
		if err != nil {
			return nil // Skip files with invalid frontmatter
		}

		// Apply metadata filters
		if settings.Ticket != "" && doc.Ticket != settings.Ticket {
			return nil
		}
		if settings.Status != "" && doc.Status != settings.Status {
			return nil
		}
		if settings.DocType != "" && doc.DocType != settings.DocType {
			return nil
		}
		if len(settings.Topics) > 0 {
			topicMatch := false
			for _, filterTopic := range settings.Topics {
				for _, docTopic := range doc.Topics {
					if strings.EqualFold(strings.TrimSpace(filterTopic), strings.TrimSpace(docTopic)) {
						topicMatch = true
						break
					}
				}
				if topicMatch {
					break
				}
			}
			if !topicMatch {
				return nil
			}
		}

		// Apply file filter (reverse lookup)
		var matchedFiles []string
		var matchedNotes []string
		if settings.File != "" {
			fileMatch := false
			for _, rf := range doc.RelatedFiles {
				relatedFile := rf.Path
				if relatedFile != "" && settings.File != "" && (strings.Contains(relatedFile, settings.File) || strings.Contains(settings.File, relatedFile)) {
					fileMatch = true
					matchedFiles = append(matchedFiles, relatedFile)
					if strings.TrimSpace(rf.Note) != "" {
						matchedNotes = append(matchedNotes, rf.Note)
					}
				}
			}
			if !fileMatch {
				return nil
			}
		}

		// Apply directory filter (reverse lookup)
		if settings.Dir != "" {
			dirMatch := false
			// Check if document is in the directory
			relPath, _ := filepath.Rel(settings.Root, path)
			if strings.HasPrefix(relPath, settings.Dir) {
				dirMatch = true
			}
			// Check if any RelatedFiles are in the directory
			if !dirMatch {
				for _, rf := range doc.RelatedFiles {
					if strings.HasPrefix(rf.Path, settings.Dir) {
						dirMatch = true
						break
					}
				}
			}
			if !dirMatch {
				return nil
			}
		}

		// Apply external source filter
		if settings.ExternalSource != "" {
			sourceMatch := false
			for _, externalSource := range doc.ExternalSources {
				if strings.Contains(externalSource, settings.ExternalSource) || strings.Contains(settings.ExternalSource, externalSource) {
					sourceMatch = true
					break
				}
			}
			if !sourceMatch {
				return nil
			}
		}

		// Apply date filters
		// Get file modification time for Created checks
		fileInfo, err := os.Stat(path)
		if err == nil {
			createdTime := fileInfo.ModTime()
			if !createdSinceTime.IsZero() && createdTime.Before(createdSinceTime) {
				return nil
			}
		}

		// Check LastUpdated field
		if !doc.LastUpdated.IsZero() {
			if !sinceTime.IsZero() && doc.LastUpdated.Before(sinceTime) {
				return nil
			}
			if !untilTime.IsZero() && doc.LastUpdated.After(untilTime) {
				return nil
			}
			if !updatedSinceTime.IsZero() && doc.LastUpdated.Before(updatedSinceTime) {
				return nil
			}
		}

		// Apply content search
		if settings.Query != "" {
			contentLower := strings.ToLower(content)
			if !strings.Contains(contentLower, queryLower) {
				return nil
			}
		}

		// Get relative path from root
		relPath, err := filepath.Rel(settings.Root, path)
		if err != nil {
			relPath = path
		}

		// Extract snippet around query match
		snippet := extractSnippet(content, settings.Query, 100)

		row := types.NewRow(
			types.MRP("ticket", doc.Ticket),
			types.MRP("title", doc.Title),
			types.MRP("doc_type", doc.DocType),
			types.MRP("status", doc.Status),
			types.MRP("topics", strings.Join(doc.Topics, ", ")),
			types.MRP("path", relPath),
			types.MRP("snippet", snippet),
		)

		// When filtering by --file, include matched file and note columns for context
		if settings.File != "" {
			if len(matchedFiles) > 0 {
				row.Set("file", strings.Join(matchedFiles, ", "))
			}
			if len(matchedNotes) > 0 {
				row.Set("file_note", strings.Join(matchedNotes, " | "))
			}
		}

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}

		return nil
	})

	return err
}

// suggestFiles suggests related files using heuristics (git + ripgrep)
func (c *SearchCommand) suggestFiles(
	ctx context.Context,
	settings *SearchSettings,
	gp middlewares.Processor,
) error {
	// Find ticket directory if specified
	var ticketDir string
	var err error
	if settings.Ticket != "" {
		ticketDir, err = findTicketDirectory(settings.Root, settings.Ticket)
		if err != nil {
			return fmt.Errorf("failed to find ticket directory: %w", err)
		}
	} else {
		ticketDir = settings.Root
	}

	// Collect search terms from query and topics
	searchTerms := []string{}
	if settings.Query != "" {
		searchTerms = append(searchTerms, settings.Query)
	}
	searchTerms = append(searchTerms, settings.Topics...)

	// Use heuristics to suggest files
	// For now, we'll use a simple approach:
	// 1. If query is provided, search for it in code files
	// 2. If topics are provided, search for topic-related terms
	// 3. Look at RelatedFiles in documents for hints

	suggestedFiles := make(map[string]bool)

	// Search documents for RelatedFiles
	err = filepath.Walk(ticketDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Apply topic filter if specified
		if len(settings.Topics) > 0 {
			doc, err := readDocumentFrontmatter(path)
			if err != nil {
				return nil
			}
			topicMatch := false
			for _, filterTopic := range settings.Topics {
				for _, docTopic := range doc.Topics {
					if strings.EqualFold(strings.TrimSpace(filterTopic), strings.TrimSpace(docTopic)) {
						topicMatch = true
						break
					}
				}
				if topicMatch {
					break
				}
			}
			if !topicMatch {
				return nil
			}
		}

		// Collect RelatedFiles from documents
		doc, err := readDocumentFrontmatter(path)
		if err != nil {
			return nil
		}
		for _, rf := range doc.RelatedFiles {
			if rf.Path != "" {
				suggestedFiles[rf.Path] = true
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking directory: %w", err)
	}

	// Output suggested files from RelatedFiles
	for file := range suggestedFiles {
		row := types.NewRow(
			types.MRP("file", file),
			types.MRP("source", "related_files"),
			types.MRP("reason", "referenced by documents"),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	// Add git-based heuristics (recent commits, changed files)
	gitFiles, err := suggestFilesFromGit(ticketDir, searchTerms)
	if err == nil {
		for _, file := range gitFiles {
			if !suggestedFiles[file] {
				row := types.NewRow(
					types.MRP("file", file),
					types.MRP("source", "git_history"),
					types.MRP("reason", "recent commit activity"),
				)
				if err := gp.AddRow(ctx, row); err != nil {
					return err
				}
			}
		}
	}

	// Add git status heuristics (modified, staged, untracked)
	if modified, staged, untracked, err := suggestFilesFromGitStatus(ticketDir); err == nil {
		for _, file := range modified {
			if !suggestedFiles[file] {
				row := types.NewRow(
					types.MRP("file", file),
					types.MRP("source", "git_modified"),
					types.MRP("reason", "working tree modified"),
				)
				if err := gp.AddRow(ctx, row); err != nil {
					return err
				}
			}
		}
		for _, file := range staged {
			if !suggestedFiles[file] {
				row := types.NewRow(
					types.MRP("file", file),
					types.MRP("source", "git_staged"),
					types.MRP("reason", "staged for commit"),
				)
				if err := gp.AddRow(ctx, row); err != nil {
					return err
				}
			}
		}
		for _, file := range untracked {
			if !suggestedFiles[file] {
				row := types.NewRow(
					types.MRP("file", file),
					types.MRP("source", "git_untracked"),
					types.MRP("reason", "untracked new file"),
				)
				if err := gp.AddRow(ctx, row); err != nil {
					return err
				}
			}
		}
	}

	// Add ripgrep-based heuristics (search for query/topics in code)
	if settings.Query != "" || len(settings.Topics) > 0 {
		ripgrepFiles, err := suggestFilesFromRipgrep(ticketDir, searchTerms)
		if err == nil {
			for _, file := range ripgrepFiles {
				if !suggestedFiles[file] {
					row := types.NewRow(
						types.MRP("file", file),
						types.MRP("source", "ripgrep"),
						types.MRP("reason", fmt.Sprintf("content match: %s", firstTerm(searchTerms))),
					)
					if err := gp.AddRow(ctx, row); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

// suggestFilesFromGit suggests files based on git history
func suggestFilesFromGit(repoPath string, searchTerms []string) ([]string, error) {
	// Check if we're in a git repository
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("not a git repository")
	}

	// Get recently modified files (last 30 commits)
	cmd = exec.Command("git", "log", "--name-only", "--pretty=format:", "-30")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git log: %w", err)
	}

	files := make(map[string]bool)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Skip non-file paths
		if strings.Contains(line, " ") {
			continue
		}
		// Only include common code file extensions
		if isCodeFile(line) {
			files[line] = true
		}
	}

	// Convert to slice
	result := make([]string, 0, len(files))
	for file := range files {
		result = append(result, file)
	}

	return result, nil
}

// suggestFilesFromGitStatus returns modified, staged, and untracked files
func suggestFilesFromGitStatus(repoPath string) ([]string, []string, []string, error) {
	// Check if in a git repo
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return nil, nil, nil, fmt.Errorf("not a git repository")
	}

	// Unstaged modified
	cmd = exec.Command("git", "diff", "--name-only")
	cmd.Dir = repoPath
	outMod, _ := cmd.Output()
	modified := nonEmptyLines(string(outMod))

	// Staged changes
	cmd = exec.Command("git", "diff", "--name-only", "--cached")
	cmd.Dir = repoPath
	outStaged, _ := cmd.Output()
	staged := nonEmptyLines(string(outStaged))

	// Untracked files
	cmd = exec.Command("git", "ls-files", "--others", "--exclude-standard")
	cmd.Dir = repoPath
	outUntracked, _ := cmd.Output()
	untracked := nonEmptyLines(string(outUntracked))

	// Filter to code-like files
	modified = filterCodeFiles(modified)
	staged = filterCodeFiles(staged)
	untracked = filterCodeFiles(untracked)

	return modified, staged, untracked, nil
}

// suggestFilesFromRipgrep suggests files using ripgrep
func suggestFilesFromRipgrep(searchPath string, searchTerms []string) ([]string, error) {
	if len(searchTerms) == 0 {
		return nil, nil
	}

	// Try to find ripgrep
	rgPath, err := exec.LookPath("rg")
	if err != nil {
		// Fallback to grep if ripgrep not available
		return suggestFilesFromGrep(searchPath, searchTerms)
	}

	// Build ripgrep command
	args := []string{
		"--files-with-matches",
		"--type", "go",
		"--type", "typescript",
		"--type", "javascript",
		"--type", "python",
		"--type", "rust",
		"--type", "java",
		"--type", "kotlin",
		"--type", "scala",
	}

	// Use first search term as query
	if len(searchTerms) > 0 {
		args = append(args, searchTerms[0])
	}

	cmd := exec.Command(rgPath, args...)
	cmd.Dir = searchPath
	output, err := cmd.Output()
	if err != nil {
		// ripgrep returns exit code 1 when no matches found, which is OK
		if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() == 1 {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to run ripgrep: %w", err)
	}

	files := []string{}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}

	return files, nil
}

// suggestFilesFromGrep suggests files using grep (fallback when ripgrep not available)
func suggestFilesFromGrep(searchPath string, searchTerms []string) ([]string, error) {
	if len(searchTerms) == 0 {
		return nil, nil
	}

	// Use grep to find files
	args := []string{
		"-r", "-l",
		"--include=*.go",
		"--include=*.ts",
		"--include=*.js",
		"--include=*.py",
		"--include=*.rs",
		"--include=*.java",
		searchTerms[0],
		searchPath,
	}

	cmd := exec.Command("grep", args...)
	output, err := cmd.Output()
	if err != nil {
		// grep returns exit code 1 when no matches found, which is OK
		if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() == 1 {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to run grep: %w", err)
	}

	files := []string{}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}

	return files, nil
}

// isCodeFile checks if a file path looks like a code file
func isCodeFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	codeExts := map[string]bool{
		".go": true, ".ts": true, ".js": true, ".py": true,
		".rs": true, ".java": true, ".kt": true, ".scala": true,
		".cpp": true, ".c": true, ".h": true, ".hpp": true,
		".rb": true, ".php": true, ".swift": true,
	}
	return codeExts[ext]
}

func nonEmptyLines(s string) []string {
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

func filterCodeFiles(files []string) []string {
	out := make([]string, 0, len(files))
	for _, f := range files {
		if isCodeFile(f) {
			out = append(out, f)
		}
	}
	return out
}

func firstTerm(terms []string) string {
	if len(terms) == 0 {
		return ""
	}
	return terms[0]
}

// extractSnippet extracts a snippet of text around a query match
func extractSnippet(content, query string, contextLen int) string {
	if query == "" {
		// Return first contextLen characters
		if len(content) <= contextLen {
			return content
		}
		return content[:contextLen] + "..."
	}

	queryLower := strings.ToLower(query)
	contentLower := strings.ToLower(content)

	idx := strings.Index(contentLower, queryLower)
	if idx == -1 {
		// Query not found, return beginning
		if len(content) <= contextLen {
			return content
		}
		return content[:contextLen] + "..."
	}

	start := idx - contextLen
	if start < 0 {
		start = 0
	}
	end := idx + len(query) + contextLen
	if end > len(content) {
		end = len(content)
	}

	snippet := content[start:end]
	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(content) {
		snippet = snippet + "..."
	}

	return snippet
}

// parseDate parses relative and absolute date strings
// Supports formats like:
// - Relative: "2 weeks ago", "last month", "1 day ago", "3 months ago"
// - Absolute: "2025-01-01", "2025-01-01 15:04:05"
// - Predefined: "today", "yesterday", "last week", "this month"
func parseDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, nil
	}

	dateStr = strings.TrimSpace(dateStr)
	dateStrLower := strings.ToLower(dateStr)

	now := time.Now()

	// Handle predefined ranges
	switch dateStrLower {
	case "today":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()), nil
	case "yesterday":
		yesterday := now.AddDate(0, 0, -1)
		return time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location()), nil
	case "last week", "lastweek":
		weekday := int(now.Weekday())
		if weekday == 0 { // Sunday
			weekday = 7
		}
		thisWeekStart := now.AddDate(0, 0, -(weekday - 1))
		start := thisWeekStart.AddDate(0, 0, -7)
		return time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location()), nil
	case "this week", "thisweek":
		weekday := int(now.Weekday())
		if weekday == 0 { // Sunday
			weekday = 7
		}
		start := now.AddDate(0, 0, -(weekday - 1))
		return time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location()), nil
	case "last month", "lastmonth":
		thisMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		start := thisMonthStart.AddDate(0, -1, 0)
		return start, nil
	case "this month", "thismonth":
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()), nil
	case "last year", "lastyear":
		return time.Date(now.Year()-1, 1, 1, 0, 0, 0, 0, now.Location()), nil
	case "this year", "thisyear":
		return time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location()), nil
	}

	// Handle relative dates like "2 weeks ago", "1 day ago", "3 months ago"
	relativePattern := regexp.MustCompile(`^(\d+)\s+(day|week|month|year)(s?)\s+ago$`)
	matches := relativePattern.FindStringSubmatch(dateStrLower)
	if len(matches) == 4 {
		num, err := strconv.Atoi(matches[1])
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid number in relative date: %s", dateStr)
		}
		unit := matches[2]
		var result time.Time
		switch unit {
		case "day":
			result = now.AddDate(0, 0, -num)
		case "week":
			result = now.AddDate(0, 0, -num*7)
		case "month":
			result = now.AddDate(0, -num, 0)
		case "year":
			result = now.AddDate(-num, 0, 0)
		default:
			return time.Time{}, fmt.Errorf("unknown time unit: %s", unit)
		}
		return time.Date(result.Year(), result.Month(), result.Day(), 0, 0, 0, 0, result.Location()), nil
	}

	// Handle "last week", "last month" without numbers
	if strings.HasPrefix(dateStrLower, "last ") {
		rest := strings.TrimPrefix(dateStrLower, "last ")
		switch rest {
		case "week":
			weekday := int(now.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			thisWeekStart := now.AddDate(0, 0, -(weekday - 1))
			start := thisWeekStart.AddDate(0, 0, -7)
			return time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location()), nil
		case "month":
			thisMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			start := thisMonthStart.AddDate(0, -1, 0)
			return start, nil
		case "year":
			return time.Date(now.Year()-1, 1, 1, 0, 0, 0, 0, now.Location()), nil
		}
	}

	// Try absolute date formats
	formats := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		time.RFC3339,
		time.RFC3339Nano,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

var _ cmds.GlazeCommand = &SearchCommand{}

// Implement BareCommand for human-friendly output
func (c *SearchCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
) error {
	settings := &SearchSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}
	settings.Root = workspace.ResolveRoot(settings.Root)

	// Suggest files mode
	if settings.Files {
		// derive ticketDir
		ticketDir := settings.Root
		if settings.Ticket != "" {
			if td, err := findTicketDirectory(settings.Root, settings.Ticket); err == nil {
				ticketDir = td
			}
		}
		// collect search terms
		terms := []string{}
		if settings.Query != "" {
			terms = append(terms, settings.Query)
		}
		terms = append(terms, settings.Topics...)

		// existing docs' RelatedFiles
		existing := map[string]bool{}
		existingNotes := map[string]map[string]bool{}
		_ = filepath.Walk(ticketDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
				return nil
			}
			doc, err := readDocumentFrontmatter(path)
			if err != nil {
				return nil
			}
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
			if f == "" {
				continue
			}
			note := "referenced by documents"
			if notes, ok := existingNotes[f]; ok {
				var ns []string
				for n := range notes {
					ns = append(ns, n)
				}
				sort.Strings(ns)
				if len(ns) > 0 {
					note += "; note: " + strings.Join(ns, "; ")
				}
			}
			fmt.Printf("%s — %s (source=related_files)\n", f, note)
		}
		if files, err := suggestFilesFromGit(ticketDir, terms); err == nil {
			for _, f := range files {
				fmt.Printf("%s — recent commit activity (source=git_history)\n", f)
			}
		}
		if files, err := suggestFilesFromRipgrep(ticketDir, terms); err == nil {
			reason := fmt.Sprintf("content match: %s", firstTerm(terms))
			for _, f := range files {
				fmt.Printf("%s — %s (source=ripgrep)\n", f, reason)
			}
		}
		if modified, staged, untracked, err := suggestFilesFromGitStatus(ticketDir); err == nil {
			for _, f := range modified {
				fmt.Printf("%s — working tree modified (source=git_modified)\n", f)
			}
			for _, f := range staged {
				fmt.Printf("%s — staged for commit (source=git_staged)\n", f)
			}
			for _, f := range untracked {
				fmt.Printf("%s — untracked new file (source=git_untracked)\n", f)
			}
		}
		return nil
	}

	// Validate query/filters presence
	if settings.Query == "" && settings.Ticket == "" && len(settings.Topics) == 0 && settings.DocType == "" && settings.Status == "" &&
		settings.File == "" && settings.Dir == "" && settings.ExternalSource == "" &&
		settings.Since == "" && settings.Until == "" && settings.CreatedSince == "" && settings.UpdatedSince == "" {
		return fmt.Errorf("must provide at least a query or filter")
	}

	sinceTime, err := parseDate(settings.Since)
	if err != nil {
		return fmt.Errorf("invalid --since date: %w", err)
	}
	untilTime, err := parseDate(settings.Until)
	if err != nil {
		return fmt.Errorf("invalid --until date: %w", err)
	}
	createdSinceTime, err := parseDate(settings.CreatedSince)
	if err != nil {
		return fmt.Errorf("invalid --created-since date: %w", err)
	}
	updatedSinceTime, err := parseDate(settings.UpdatedSince)
	if err != nil {
		return fmt.Errorf("invalid --updated-since date: %w", err)
	}
	if _, err := os.Stat(settings.Root); os.IsNotExist(err) {
		return fmt.Errorf("root directory does not exist: %s", settings.Root)
	}

	queryLower := strings.ToLower(settings.Query)

	_ = filepath.Walk(settings.Root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}
		if strings.Contains(path, "/_templates/") || strings.Contains(path, "/_guidelines/") {
			return nil
		}

		doc, content, err := readDocumentWithContent(path)
		if err != nil {
			return nil
		}

		if settings.Ticket != "" && doc.Ticket != settings.Ticket {
			return nil
		}
		if settings.Status != "" && doc.Status != settings.Status {
			return nil
		}
		if settings.DocType != "" && doc.DocType != settings.DocType {
			return nil
		}
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

		var matchedFiles []string
		var matchedNotes []string
		if settings.File != "" {
			fileMatch := false
			for _, rf := range doc.RelatedFiles {
				relatedFile := rf.Path
				if relatedFile != "" && (strings.Contains(relatedFile, settings.File) || strings.Contains(settings.File, relatedFile)) {
					fileMatch = true
					matchedFiles = append(matchedFiles, relatedFile)
					if strings.TrimSpace(rf.Note) != "" {
						matchedNotes = append(matchedNotes, rf.Note)
					}
				}
			}
			if !fileMatch {
				return nil
			}
		}

		if settings.Dir != "" {
			dirMatch := false
			relPath, _ := filepath.Rel(settings.Root, path)
			if strings.HasPrefix(relPath, settings.Dir) {
				dirMatch = true
			}
			if !dirMatch {
				for _, rf := range doc.RelatedFiles {
					if strings.HasPrefix(rf.Path, settings.Dir) {
						dirMatch = true
						break
					}
				}
			}
			if !dirMatch {
				return nil
			}
		}

		if settings.ExternalSource != "" {
			sourceMatch := false
			for _, es := range doc.ExternalSources {
				if strings.Contains(es, settings.ExternalSource) || strings.Contains(settings.ExternalSource, es) {
					sourceMatch = true
					break
				}
			}
			if !sourceMatch {
				return nil
			}
		}

		if fi, err := os.Stat(path); err == nil {
			createdTime := fi.ModTime()
			if !createdSinceTime.IsZero() && createdTime.Before(createdSinceTime) {
				return nil
			}
		}
		if !doc.LastUpdated.IsZero() {
			if !sinceTime.IsZero() && doc.LastUpdated.Before(sinceTime) {
				return nil
			}
			if !untilTime.IsZero() && doc.LastUpdated.After(untilTime) {
				return nil
			}
			if !updatedSinceTime.IsZero() && doc.LastUpdated.Before(updatedSinceTime) {
				return nil
			}
		}

		if settings.Query != "" {
			if !strings.Contains(strings.ToLower(content), queryLower) {
				return nil
			}
		}

		relPath, err := filepath.Rel(settings.Root, path)
		if err != nil {
			relPath = path
		}
		snippet := extractSnippet(content, settings.Query, 100)

		if settings.File != "" {
			extra := ""
			if len(matchedFiles) > 0 {
				extra += " file=" + strings.Join(matchedFiles, ", ")
			}
			if len(matchedNotes) > 0 {
				extra += " note=" + strings.Join(matchedNotes, " | ")
			}
			fmt.Printf("%s — %s [%s] :: %s%s\n", relPath, doc.Title, doc.Ticket, snippet, extra)
		} else {
			fmt.Printf("%s — %s [%s] :: %s\n", relPath, doc.Title, doc.Ticket, snippet)
		}
		return nil
	})

	return nil
}

var _ cmds.BareCommand = &SearchCommand{}
