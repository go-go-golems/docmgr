package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/go-go-golems/docmgr/internal/templates"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	glazedMiddlewares "github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/mattn/go-isatty"
)

// DoctorCommand validates document workspaces
type DoctorCommand struct {
	*cmds.CommandDescription
}

// DoctorSettings holds the parameters for the doctor command
type DoctorSettings struct {
	Root           string   `glazed.parameter:"root"`
	Ticket         string   `glazed.parameter:"ticket"`
	All            bool     `glazed.parameter:"all"`
	IgnoreDirs     []string `glazed.parameter:"ignore-dir"`
	IgnoreGlobs    []string `glazed.parameter:"ignore-glob"`
	StaleAfterDays int      `glazed.parameter:"stale-after"`
	FailOn         string   `glazed.parameter:"fail-on"`
	// Schema printing flags (human mode only)
	PrintTemplateSchema bool   `glazed.parameter:"print-template-schema"`
	SchemaFormat        string `glazed.parameter:"schema-format"`
}

func NewDoctorCommand() (*DoctorCommand, error) {
	return &DoctorCommand{
		CommandDescription: cmds.NewCommandDescription(
			"doctor",
			cmds.WithShort("Validate document workspaces"),
			cmds.WithLong(`Checks document workspaces for issues like missing frontmatter,
invalid metadata, or broken structure. Respects a repository-level .docmgrignore file
for path exclusions (similar to .gitignore). Each non-empty line is a glob or name to
ignore; lines starting with # are treated as comments.

Example:
  docmgr doctor --ticket MEN-3475
  docmgr doctor --all
`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"root",
					parameters.ParameterTypeString,
					parameters.WithHelp("Root directory for docs"),
					parameters.WithDefault("ttmp"),
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
				parameters.NewParameterDefinition(
					"ticket",
					parameters.ParameterTypeString,
					parameters.WithHelp("Check specific ticket"),
					parameters.WithDefault(""),
				),
				parameters.NewParameterDefinition(
					"all",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Check all tickets"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"ignore-dir",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Directory names at root or within tickets to ignore (can be repeated)"),
					parameters.WithDefault([]string{}),
				),
				parameters.NewParameterDefinition(
					"ignore-glob",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Glob patterns (applied to path or basename) to ignore during scanning"),
					parameters.WithDefault([]string{}),
				),
				parameters.NewParameterDefinition(
					"stale-after",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Days after which a document is considered stale (default 30)"),
					parameters.WithDefault(30),
				),
				parameters.NewParameterDefinition(
					"fail-on",
					parameters.ParameterTypeString,
					parameters.WithHelp("Fail with non-zero exit on severity: none|warning|error (default none)"),
					parameters.WithDefault("none"),
				),
			),
		),
	}, nil
}

func (c *DoctorCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp glazedMiddlewares.Processor,
) error {
	settings := &DoctorSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Apply config root if present
	settings.Root = workspace.ResolveRoot(settings.Root)

	// If only printing template schema, skip all other processing and output
	if settings.PrintTemplateSchema {
		type Finding struct {
			Issue    string
			Severity string
			Message  string
			Path     string
		}
		type TicketFindings struct {
			Ticket   string
			Findings []Finding
		}
		templateData := map[string]interface{}{
			"TotalFindings": 0,
			"Tickets": []TicketFindings{
				{
					Ticket:   "",
					Findings: []Finding{{}},
				},
			},
		}
		_ = templates.PrintSchema(os.Stdout, templateData, settings.SchemaFormat)
		return nil
	}

	if _, err := os.Stat(settings.Root); os.IsNotExist(err) {
		return fmt.Errorf("root directory does not exist: %s", settings.Root)
	}

	// Track highest severity encountered to support --fail-on
	highestSeverity := 0 // 0=ok,1=warning,2=error

	// Determine repository root
	repoRoot, _ := workspace.FindRepositoryRoot()

	// Load .docmgrignore patterns and merge with provided ignore-globs
	// 1) Try repository root
	if repoRoot != "" {
		if patterns, err := loadDocmgrIgnore(repoRoot); err == nil {
			settings.IgnoreGlobs = append(settings.IgnoreGlobs, patterns...)
		}
	}
	// 2) Also try docs root (settings.Root), to support non-git environments
	// Avoid double-loading if paths are identical
	if settings.Root != "" && filepath.Clean(settings.Root) != filepath.Clean(repoRoot) {
		if patterns, err := loadDocmgrIgnore(settings.Root); err == nil {
			settings.IgnoreGlobs = append(settings.IgnoreGlobs, patterns...)
		}
	}

	// Load vocabulary for validation (best-effort)
	vocab, _ := LoadVocabulary()
	topicSet := map[string]struct{}{}
	for _, it := range vocab.Topics {
		topicSet[it.Slug] = struct{}{}
	}
	docTypeSet := map[string]struct{}{}
	for _, it := range vocab.DocTypes {
		docTypeSet[it.Slug] = struct{}{}
	}
	intentSet := map[string]struct{}{}
	for _, it := range vocab.Intent {
		intentSet[it.Slug] = struct{}{}
	}
	statusSet := map[string]struct{}{}
	statusList := make([]string, 0, len(vocab.Status))
	for _, it := range vocab.Status {
		statusSet[it.Slug] = struct{}{}
		statusList = append(statusList, it.Slug)
	}
	statusValidText := "none defined (add via 'docmgr vocab add --category status --slug <slug>')"
	if len(statusList) > 0 {
		statusValidText = strings.Join(statusList, ", ")
	}

	skipFn := func(relPath, base string) bool {
		if containsString(settings.IgnoreDirs, base) {
			return true
		}
		full := filepath.Join(settings.Root, relPath)
		if relPath == "." {
			full = settings.Root
		}
		if matchesAnyGlob(settings.IgnoreGlobs, base) || matchesAnyGlob(settings.IgnoreGlobs, full) {
			return true
		}
		return false
	}

	workspaces, err := workspace.CollectTicketWorkspaces(settings.Root, skipFn)
	if err != nil {
		return fmt.Errorf("failed to discover ticket workspaces: %w", err)
	}

	missingIndexDirs, err := workspace.CollectTicketScaffoldsWithoutIndex(settings.Root, skipFn)
	if err != nil {
		return fmt.Errorf("failed to detect missing index.md files: %w", err)
	}
	for _, missing := range missingIndexDirs {
		row := types.NewRow(
			types.MRP("ticket", filepath.Base(missing)),
			types.MRP("issue", "missing_index"),
			types.MRP("severity", "error"),
			types.MRP("message", "index.md not found"),
			types.MRP("path", missing),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return fmt.Errorf("failed to emit doctor row (missing_index) for %s: %w", missing, err)
		}
		highestSeverity = maxInt(highestSeverity, 2)
	}

	for _, ws := range workspaces {
		ticketPath := ws.Path
		indexPath := filepath.Join(ticketPath, "index.md")
		if ws.Doc == nil {
			row := types.NewRow(
				types.MRP("ticket", filepath.Base(ticketPath)),
				types.MRP("issue", "invalid_frontmatter"),
				types.MRP("severity", "error"),
				types.MRP("message", fmt.Sprintf("Failed to parse frontmatter: %v", ws.FrontmatterErr)),
				types.MRP("path", indexPath),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return fmt.Errorf("failed to emit doctor row (invalid_frontmatter) for %s: %w", filepath.Base(ticketPath), err)
			}
			highestSeverity = maxInt(highestSeverity, 2)
			continue
		}

		doc := ws.Doc
		if settings.Ticket != "" && doc.Ticket != settings.Ticket {
			continue
		}

		// Track all issues found
		hasIssues := false

		// Check for unique index.md (should only be one per workspace)
		indexFiles := findIndexFiles(ticketPath, settings.IgnoreDirs, settings.IgnoreGlobs)
		if len(indexFiles) > 1 {
			hasIssues = true
			row := types.NewRow(
				types.MRP("ticket", doc.Ticket),
				types.MRP("issue", "multiple_index"),
				types.MRP("severity", "warning"),
				types.MRP("message", fmt.Sprintf("Multiple index.md files found (%d), should be only one", len(indexFiles))),
				types.MRP("path", ticketPath),
				types.MRP("index_files", fmt.Sprintf("%v", indexFiles)),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return fmt.Errorf("failed to emit doctor row (multiple_index) for %s: %w", doc.Ticket, err)
			}
			highestSeverity = maxInt(highestSeverity, 1)
		}

		// Check for staleness (LastUpdated > stale-after days)
		if !doc.LastUpdated.IsZero() {
			daysSinceUpdate := time.Since(doc.LastUpdated).Hours() / 24
			if daysSinceUpdate > float64(settings.StaleAfterDays) {
				hasIssues = true
				row := types.NewRow(
					types.MRP("ticket", doc.Ticket),
					types.MRP("issue", "stale"),
					types.MRP("severity", "warning"),
					types.MRP("message", fmt.Sprintf("LastUpdated is %.0f days old (threshold: 14 days)", daysSinceUpdate)),
					types.MRP("path", indexPath),
					types.MRP("last_updated", doc.LastUpdated.Format("2006-01-02")),
				)
				if err := gp.AddRow(ctx, row); err != nil {
					return fmt.Errorf("failed to emit doctor row (stale) for %s: %w", doc.Ticket, err)
				}
				highestSeverity = maxInt(highestSeverity, 1)
			}
		}

		// Validate required fields using Document.Validate()
		if err := doc.Validate(); err != nil {
			hasIssues = true
			row := types.NewRow(
				types.MRP("ticket", doc.Ticket),
				types.MRP("issue", "missing_required_fields"),
				types.MRP("severity", "error"),
				types.MRP("message", err.Error()),
				types.MRP("path", indexPath),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return fmt.Errorf("failed to emit doctor row (missing_required_fields) for %s: %w", doc.Ticket, err)
			}
			highestSeverity = maxInt(highestSeverity, 2)
		}

		// Additional validation: Status and Topics (not in Validate() but checked by doctor)
		issues := []string{}
		if doc.Status == "" {
			issues = append(issues, "missing Status")
		}
		if len(doc.Topics) == 0 {
			issues = append(issues, "missing Topics")
		}

		// Validate vocabulary: Topics, DocType, Intent
		// Unknown topics
		var unknownTopics []string
		for _, t := range doc.Topics {
			if _, ok := topicSet[t]; !ok && t != "" {
				unknownTopics = append(unknownTopics, t)
			}
		}
		if len(unknownTopics) > 0 {
			hasIssues = true
			row := types.NewRow(
				types.MRP("ticket", doc.Ticket),
				types.MRP("issue", "unknown_topics"),
				types.MRP("severity", "warning"),
				types.MRP("message", fmt.Sprintf("unknown topics: %v", unknownTopics)),
				types.MRP("path", indexPath),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return fmt.Errorf("failed to emit doctor row (unknown_topics) for %s: %w", doc.Ticket, err)
			}
			highestSeverity = maxInt(highestSeverity, 1)
		}

		// Unknown docType
		if doc.DocType != "" {
			if _, ok := docTypeSet[doc.DocType]; !ok {
				hasIssues = true
				row := types.NewRow(
					types.MRP("ticket", doc.Ticket),
					types.MRP("issue", "unknown_doc_type"),
					types.MRP("severity", "warning"),
					types.MRP("message", fmt.Sprintf("unknown docType: %s", doc.DocType)),
					types.MRP("path", indexPath),
				)
				if err := gp.AddRow(ctx, row); err != nil {
					return fmt.Errorf("failed to emit doctor row (unknown_doc_type) for %s: %w", doc.Ticket, err)
				}
				highestSeverity = maxInt(highestSeverity, 1)
			}
		}

		// Unknown intent
		if doc.Intent != "" {
			if _, ok := intentSet[doc.Intent]; !ok {
				hasIssues = true
				row := types.NewRow(
					types.MRP("ticket", doc.Ticket),
					types.MRP("issue", "unknown_intent"),
					types.MRP("severity", "warning"),
					types.MRP("message", fmt.Sprintf("unknown intent: %s", doc.Intent)),
					types.MRP("path", indexPath),
				)
				if err := gp.AddRow(ctx, row); err != nil {
					return fmt.Errorf("failed to emit doctor row (unknown_intent) for %s: %w", doc.Ticket, err)
				}
				highestSeverity = maxInt(highestSeverity, 1)
			}
		}

		// Unknown status (vocabulary-guided, warn only)
		if doc.Status != "" {
			if _, ok := statusSet[doc.Status]; !ok {
				hasIssues = true
				row := types.NewRow(
					types.MRP("ticket", doc.Ticket),
					types.MRP("issue", "unknown_status"),
					types.MRP("severity", "warning"),
					types.MRP("message", fmt.Sprintf("unknown status: %s (valid values: %s; list via 'docmgr vocab list --category status')", doc.Status, statusValidText)),
					types.MRP("path", indexPath),
				)
				if err := gp.AddRow(ctx, row); err != nil {
					return fmt.Errorf("failed to emit doctor row (unknown_status) for %s: %w", doc.Ticket, err)
				}
				highestSeverity = maxInt(highestSeverity, 1)
			}
		}

		// Validate RelatedFiles existence with robust resolution
		for _, rf := range doc.RelatedFiles {
			if rf.Path == "" {
				continue
			}
			// Warn when a related file is missing an explanatory Note
			if strings.TrimSpace(rf.Note) == "" {
				hasIssues = true
				row := types.NewRow(
					types.MRP("ticket", doc.Ticket),
					types.MRP("issue", "missing_related_file_note"),
					types.MRP("severity", "warning"),
					types.MRP("message", fmt.Sprintf("related file has no Note: %s", rf.Path)),
					types.MRP("path", indexPath),
				)
				if err := gp.AddRow(ctx, row); err != nil {
					return fmt.Errorf("failed to emit doctor row (missing_related_file_note) for %s: %w", doc.Ticket, err)
				}
				highestSeverity = maxInt(highestSeverity, 1)
			}
			// Build resolution candidates
			candidates := []string{}
			if filepath.IsAbs(rf.Path) {
				candidates = append(candidates, rf.Path)
			} else {
				// 1) Repo root (if any)
				if repoRoot != "" {
					candidates = append(candidates, filepath.Join(repoRoot, rf.Path))
				}
				// 2) .ttmp.yaml directory (config base)
				if cfgPath, errCfg := workspace.FindTTMPConfigPath(); errCfg == nil {
					cfgBase := filepath.Dir(cfgPath)
					candidates = append(candidates, filepath.Join(cfgBase, rf.Path))
					// 3) Parent of config base (supports multi-repo workspace siblings)
					parentBase := filepath.Dir(cfgBase)
					if parentBase != cfgBase { // guard root
						candidates = append(candidates, filepath.Join(parentBase, rf.Path))
					}
				}
				// 4) Current working directory as last resort
				if cwd, errCwd := os.Getwd(); errCwd == nil {
					candidates = append(candidates, filepath.Join(cwd, rf.Path))
				}
			}

			found := false
			for _, p := range candidates {
				if p == "" {
					continue
				}
				if _, err := os.Stat(p); err == nil {
					found = true
					break
				}
			}

			if !found {
				hasIssues = true
				row := types.NewRow(
					types.MRP("ticket", doc.Ticket),
					types.MRP("issue", "missing_related_file"),
					types.MRP("severity", "warning"),
					types.MRP("message", fmt.Sprintf("related file not found: %s", rf.Path)),
					types.MRP("path", indexPath),
				)
				if err := gp.AddRow(ctx, row); err != nil {
					return fmt.Errorf("failed to emit doctor row (missing_related_file) for %s: %w", doc.Ticket, err)
				}
				highestSeverity = maxInt(highestSeverity, 1)
			}
		}

		if len(issues) > 0 {
			hasIssues = true
			for _, issue := range issues {
				row := types.NewRow(
					types.MRP("ticket", doc.Ticket),
					types.MRP("issue", "missing_field"),
					types.MRP("severity", "warning"),
					types.MRP("message", issue),
					types.MRP("path", indexPath),
				)
				if err := gp.AddRow(ctx, row); err != nil {
					return fmt.Errorf("failed to emit doctor row (missing_field) for %s: %w", doc.Ticket, err)
				}
				highestSeverity = maxInt(highestSeverity, 1)
			}
		}

		// Check all markdown files for invalid frontmatter and enforce numeric prefix policy
		{
			prefixRe := regexp.MustCompile(`^(\d{2,3})-`)
			_ = filepath.Walk(ticketPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if info.IsDir() {
					return nil
				}
				// Only consider markdown files
				if !strings.HasSuffix(strings.ToLower(info.Name()), ".md") {
					return nil
				}
				// Skip root-level control files
				dir := filepath.Dir(path)
				isRootLevel := filepath.Clean(dir) == filepath.Clean(ticketPath)
				if isRootLevel {
					bn := info.Name()
					if bn == "index.md" || bn == "README.md" || bn == "tasks.md" || bn == "changelog.md" {
						return nil
					}
				}

				// Check for invalid frontmatter (try to parse it)
				// Skip index.md as it's already checked above
				if info.Name() != "index.md" {
					_, err := readDocumentFrontmatter(path)
					if err != nil {
						hasIssues = true
						row := types.NewRow(
							types.MRP("ticket", doc.Ticket),
							types.MRP("issue", "invalid_frontmatter"),
							types.MRP("severity", "error"),
							types.MRP("message", fmt.Sprintf("Failed to parse frontmatter: %v", err)),
							types.MRP("path", path),
						)
						if err := gp.AddRow(ctx, row); err != nil {
							return fmt.Errorf("failed to emit doctor row (invalid_frontmatter) for %s: %w", path, err)
						}
						highestSeverity = maxInt(highestSeverity, 2)
					}
				}

				// Enforce prefix on subdirectory files
				bn := info.Name()
				if !isRootLevel && !prefixRe.MatchString(bn) {
					hasIssues = true
					row := types.NewRow(
						types.MRP("ticket", doc.Ticket),
						types.MRP("issue", "missing_numeric_prefix"),
						types.MRP("severity", "warning"),
						types.MRP("message", "file without numeric prefix"),
						types.MRP("path", path),
					)
					if err := gp.AddRow(ctx, row); err != nil {
						return fmt.Errorf("failed to emit doctor row (missing_numeric_prefix) for %s: %w", path, err)
					}
					highestSeverity = maxInt(highestSeverity, 1)
				}
				return nil
			})
		}

		// Only report "All checks passed" if there are truly no issues
		if !hasIssues {
			row := types.NewRow(
				types.MRP("ticket", doc.Ticket),
				types.MRP("issue", "none"),
				types.MRP("severity", "ok"),
				types.MRP("message", "All checks passed"),
				types.MRP("path", ticketPath),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return fmt.Errorf("failed to emit doctor summary row for %s: %w", doc.Ticket, err)
			}
		}
	}

	// Enforce fail-on behavior
	threshold := severityThreshold(settings.FailOn)
	if threshold >= 0 && highestSeverity >= threshold && threshold > 0 {
		return fmt.Errorf("doctor failed: severity >= %s", settings.FailOn)
	}

	return nil
}

// findIndexFiles recursively searches for all index.md files in a directory tree
func findIndexFiles(rootPath string, ignoreDirNames []string, ignoreGlobs []string) []string {
	var indexFiles []string

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors, continue walking
		}
		// Skip ignored directories
		if info.IsDir() {
			base := filepath.Base(path)
			if containsString(ignoreDirNames, base) || matchesAnyGlob(ignoreGlobs, base) || matchesAnyGlob(ignoreGlobs, path) {
				return filepath.SkipDir
			}
			return nil
		}
		// Skip ignored files
		if matchesAnyGlob(ignoreGlobs, info.Name()) || matchesAnyGlob(ignoreGlobs, path) {
			return nil
		}
		if !info.IsDir() && info.Name() == "index.md" {
			indexFiles = append(indexFiles, path)
		}
		return nil
	})

	if err != nil {
		// Return what we found even if there was an error
		return indexFiles
	}

	return indexFiles
}

// containsString returns true if s is in list
func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// matchesAnyGlob checks if path matches any of the provided glob patterns
func matchesAnyGlob(patterns []string, path string) bool {
	for _, p := range patterns {
		p = normalizeIgnorePattern(p)
		if ok, _ := filepath.Match(p, path); ok {
			return true
		}
	}
	return false
}

// normalizeIgnorePattern trims whitespace and trailing separators to make simple
// directory entries like ".git/" match both names and paths.
func normalizeIgnorePattern(p string) string {
	p = strings.TrimSpace(p)
	for len(p) > 0 && (p[len(p)-1] == '/' || p[len(p)-1] == os.PathSeparator) {
		p = p[:len(p)-1]
	}
	return p
}

// loadDocmgrIgnore reads ignore patterns from <repoRoot>/.docmgrignore.
// Lines starting with '#' are comments; empty lines are skipped.
func loadDocmgrIgnore(repoRoot string) ([]string, error) {
	path := filepath.Join(repoRoot, ".docmgrignore")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(b), "\n")
	var patterns []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		patterns = append(patterns, l)
	}
	return patterns, nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// severityThreshold maps fail-on string to numeric threshold
// none=0 (disabled), warning=1, error=2
func severityThreshold(s string) int {
	switch strings.ToLower(s) {
	case "none":
		return 0
	case "warning":
		return 1
	case "error":
		return 2
	default:
		return 0
	}
}

var _ cmds.GlazeCommand = &DoctorCommand{}

type doctorRowCollector struct {
	rows []types.Row
}

func (c *doctorRowCollector) AddRow(ctx context.Context, row types.Row) error {
	c.rows = append(c.rows, row)
	return nil
}

func (c *doctorRowCollector) Close(ctx context.Context) error {
	return nil
}

func (c *DoctorCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
) error {
	settings := &DoctorSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Apply config root if present
	settings.Root = workspace.ResolveRoot(settings.Root)

	// If only printing template schema, skip all other processing and output
	if settings.PrintTemplateSchema {
		type Finding struct {
			Issue    string
			Severity string
			Message  string
			Path     string
		}
		type TicketFindings struct {
			Ticket   string
			Findings []Finding
		}
		templateData := map[string]interface{}{
			"TotalFindings": 0,
			"Tickets": []TicketFindings{
				{
					Ticket:   "",
					Findings: []Finding{{}},
				},
			},
		}
		_ = templates.PrintSchema(os.Stdout, templateData, settings.SchemaFormat)
		return nil
	}

	collector := &doctorRowCollector{}
	if err := c.RunIntoGlazeProcessor(ctx, parsedLayers, collector); err != nil {
		return err
	}

	rows := collector.rows
	if len(rows) == 0 {
		fmt.Println("No tickets checked.")
		return nil
	}

	grouped := map[string][]types.Row{}
	order := []string{}
	for _, row := range rows {
		ticket := getRowString(row, ColTicket)
		if ticket == "" {
			ticket = "(unknown)"
		}
		if _, ok := grouped[ticket]; !ok {
			grouped[ticket] = []types.Row{}
			order = append(order, ticket)
		}
		grouped[ticket] = append(grouped[ticket], row)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "## Doctor Report (%d findings)\n\n", len(rows))
	for _, ticket := range order {
		fmt.Fprintf(&b, "### %s\n\n", ticket)
		entries := grouped[ticket]
		for _, row := range entries {
			issue := getRowString(row, "issue")
			severity := strings.ToUpper(getRowString(row, "severity"))
			message := getRowString(row, "message")
			path := getRowString(row, "path")

			if issue == "none" && severity == "OK" {
				fmt.Fprintf(&b, "- ✅ %s\n", message)
				continue
			}
			if message == "" {
				message = "(no message)"
			}
			if path != "" {
				fmt.Fprintf(&b, "- [%s] %s — %s (path=%s)\n", severity, issue, message, path)
			} else {
				fmt.Fprintf(&b, "- [%s] %s — %s\n", severity, issue, message)
			}
		}
		fmt.Fprintln(&b)
	}

	content := b.String()
	fd := os.Stdout.Fd()
	if isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd) {
		renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(0),
		)
		if err == nil {
			if rendered, err := renderer.Render(content); err == nil {
				fmt.Print(rendered)
			} else {
				fmt.Print(content)
			}
		} else {
			fmt.Print(content)
		}
	} else {
		fmt.Print(content)
	}

	// Render postfix template if it exists
	// Build template data struct
	type Finding struct {
		Issue    string
		Severity string
		Message  string
		Path     string
	}
	type TicketFindings struct {
		Ticket   string
		Findings []Finding
	}

	ticketFindings := make([]TicketFindings, 0, len(order))
	totalFindings := 0
	for _, ticket := range order {
		entries := grouped[ticket]
		findings := make([]Finding, 0)
		for _, row := range entries {
			issue := getRowString(row, "issue")
			severity := getRowString(row, "severity")
			message := getRowString(row, "message")
			path := getRowString(row, "path")

			// Skip "none" issues (all checks passed)
			if issue == "none" {
				continue
			}

			findings = append(findings, Finding{
				Issue:    issue,
				Severity: strings.ToUpper(severity),
				Message:  message,
				Path:     path,
			})
			totalFindings++
		}
		if len(findings) > 0 {
			ticketFindings = append(ticketFindings, TicketFindings{
				Ticket:   ticket,
				Findings: findings,
			})
		}
	}

	templateData := map[string]interface{}{
		"TotalFindings": totalFindings,
		"Tickets":       ticketFindings,
	}

	// Try verb path: ["doctor"]
	verbCandidates := [][]string{
		{"doctor"},
	}
	settingsMap := map[string]interface{}{
		"root":       settings.Root,
		"ticket":     settings.Ticket,
		"all":        settings.All,
		"staleAfter": settings.StaleAfterDays,
		"failOn":     settings.FailOn,
	}
	_ = templates.RenderVerbTemplate(verbCandidates, settings.Root, settingsMap, templateData)

	return nil
}

func getRowString(row types.Row, field string) string {
	if val, ok := row.Get(field); ok {
		return fmt.Sprint(val)
	}
	return ""
}

var _ cmds.BareCommand = &DoctorCommand{}
