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
	"github.com/go-go-golems/docmgr/internal/paths"
	"github.com/go-go-golems/docmgr/internal/templates"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/core"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/docmgr"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/docmgrctx"
	"github.com/go-go-golems/docmgr/pkg/models"
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
	Root            string   `glazed.parameter:"root"`
	Ticket          string   `glazed.parameter:"ticket"`
	Doc             string   `glazed.parameter:"doc"`
	All             bool     `glazed.parameter:"all"`
	IgnoreDirs      []string `glazed.parameter:"ignore-dir"`
	IgnoreGlobs     []string `glazed.parameter:"ignore-glob"`
	StaleAfterDays  int      `glazed.parameter:"stale-after"`
	FailOn          string   `glazed.parameter:"fail-on"`
	DiagnosticsJSON string   `glazed.parameter:"diagnostics-json"`
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

Common findings (doctor message ⇒ likely cause ⇒ how to fix):
  • invalid_frontmatter — YAML block can’t be parsed. Ensure the file starts with '---',
    quote strings containing ':' or '#', and use 'docmgr meta update' to rewrite fields safely.
  • missing_required_fields — Title/Summary/DocType/etc. are missing. Run
    'docmgr meta update --ticket T --field FieldName --value ...' (or edit frontmatter) to add them.
  • missing_index — Ticket directories without index.md. Re-run 'docmgr ticket create-ticket'
    or copy the template back into place.
  • unknown_topics / unknown_status — Value not present in vocabulary.yaml. Either add it via
    'docmgr vocab add --category topics --slug your-topic' (or status/doc-type) or update the doc’s fields.
  • stale — LastUpdated is older than '--stale-after' days (default 30). Review the doc, make an update,
    or pass '--stale-after N' if the cadence should be different for this run.

Tips:
  • Use '--fail-on warning' (or 'error') to make CI fail when issues are detected.
  • '--diagnostics-json path' captures rule results as JSON (use '-' for stdout) for CI/automation.
  • '--ignore-glob' is handy for suppressing known noisy paths; the command also reads patterns from
    both repository and docs-root .docmgrignore files.

Examples:
  # Validate a specific ticket workspace
  docmgr doctor --ticket MEN-3475

  # Validate all tickets
  docmgr doctor --all

  # Tighten staleness and fail CI on warnings
  docmgr doctor --all --stale-after 14 --fail-on warning

  # Emit diagnostics JSON for scripts/CI
  docmgr doctor --ticket MEN-3475 --diagnostics-json - --with-glaze-output --output json
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
					"doc",
					parameters.ParameterTypeString,
					parameters.WithHelp("Validate a single markdown file (overrides --ticket/--all)"),
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
				parameters.NewParameterDefinition(
					"diagnostics-json",
					parameters.ParameterTypeString,
					parameters.WithHelp("Write diagnostics rule output to JSON (file path or '-' for stdout)"),
					parameters.WithDefault(""),
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

	// Optional diagnostics collector for JSON output
	var diagRenderer *docmgr.Renderer
	if settings.DiagnosticsJSON != "" {
		diagRenderer = docmgr.NewRenderer(docmgr.WithCollector())
		ctx = docmgr.ContextWithRenderer(ctx, diagRenderer)
	}

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

	// Load .docmgrignore patterns and merge with provided ignore-globs
	// NOTE: doctor historically supports ad-hoc ignore-globs on top of the canonical ingest skip policy.
	// The workspace index is built using the canonical policy; we apply ignore-globs as a post-filter
	// over QueryDocs results so doctor output matches prior behavior.
	// 1) Try repository root
	repoRoot, _ := workspace.FindRepositoryRoot()
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
	topicList := make([]string, 0, len(vocab.Topics))
	for _, it := range vocab.Topics {
		topicSet[it.Slug] = struct{}{}
		topicList = append(topicList, it.Slug)
	}
	docTypeSet := map[string]struct{}{}
	docTypeList := make([]string, 0, len(vocab.DocTypes))
	for _, it := range vocab.DocTypes {
		docTypeSet[it.Slug] = struct{}{}
		docTypeList = append(docTypeList, it.Slug)
	}
	intentSet := map[string]struct{}{}
	intentList := make([]string, 0, len(vocab.Intent))
	for _, it := range vocab.Intent {
		intentSet[it.Slug] = struct{}{}
		intentList = append(intentList, it.Slug)
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

	// Single-file mode: validate one doc and return.
	if settings.Doc != "" {
		docPath := settings.Doc
		if !filepath.IsAbs(docPath) {
			docPath = filepath.Join(settings.Root, docPath)
		}
		sev, err := validateSingleDoc(ctx, docPath, topicSet, topicList, docTypeSet, docTypeList, intentSet, intentList, statusSet, statusList, statusValidText, gp)
		highestSeverity = maxInt(highestSeverity, sev)
		if diagRenderer != nil && settings.DiagnosticsJSON != "" {
			if err := writeDiagnosticsJSON(diagRenderer, settings.DiagnosticsJSON); err != nil {
				return err
			}
		}
		threshold := severityThreshold(settings.FailOn)
		if threshold >= 0 && highestSeverity >= threshold && threshold > 0 {
			return fmt.Errorf("doctor failed: severity >= %s", settings.FailOn)
		}
		return err
	}

	// Build workspace + index once; doctor is an index-backed scan (Spec §11.2.3).
	ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{RootOverride: settings.Root})
	if err != nil {
		return fmt.Errorf("failed to discover workspace: %w", err)
	}
	if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{IncludeBody: false}); err != nil {
		return fmt.Errorf("failed to initialize workspace index: %w", err)
	}

	missingIndexDirs, err := workspace.FindTicketScaffoldsMissingIndex(ctx, settings.Root, skipFn)
	if err != nil {
		return fmt.Errorf("failed to detect missing index.md files: %w", err)
	}
	// If the user asked for a specific ticket, don't report missing-index scaffolds outside that ticket.
	if strings.TrimSpace(settings.Ticket) != "" && !settings.All {
		filteredMissing := make([]string, 0, len(missingIndexDirs))
		for _, p := range missingIndexDirs {
			base := filepath.Base(p)
			if matchesTicketDir(strings.TrimSpace(settings.Ticket), base) {
				filteredMissing = append(filteredMissing, p)
			}
		}
		missingIndexDirs = filteredMissing
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
		docmgr.RenderTaxonomy(ctx, docmgrctx.NewWorkspaceMissingIndex(missing))
	}

	// Determine scope: default is repo-wide unless --ticket is provided.
	scope := workspace.Scope{Kind: workspace.ScopeRepo}
	if strings.TrimSpace(settings.Ticket) != "" {
		scope = workspace.Scope{Kind: workspace.ScopeTicket, TicketID: strings.TrimSpace(settings.Ticket)}
	}
	if settings.All {
		scope = workspace.Scope{Kind: workspace.ScopeRepo}
	}

	query := workspace.DocQuery{
		Scope: scope,
		Options: workspace.DocQueryOptions{
			IncludeErrors:      true,
			IncludeDiagnostics: true,
			// Doctor historically scans "everything"; opt into visibility for the special path tags.
			IncludeArchivedPath: true,
			IncludeScriptsPath:  true,
			IncludeControlDocs:  true,
			OrderBy:             workspace.OrderByPath,
		},
	}

	qr, err := ws.QueryDocs(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query docs: %w", err)
	}

	// Post-filter based on ignore globs/dirs.
	filtered := make([]workspace.DocHandle, 0, len(qr.Docs))
	for _, h := range qr.Docs {
		if shouldSkipDoctorDoc(settings.Root, h.Path, settings.IgnoreDirs, settings.IgnoreGlobs) {
			continue
		}
		filtered = append(filtered, h)
	}

	// Group by ticket directory inferred from ttmp layout.
	tickets := groupDoctorDocsByTicket(settings.Root, filtered)

	// Per-ticket validations.
	prefixRe := regexp.MustCompile(`^(\d{2,3})-`)
	for _, bucket := range tickets {
		ticketPath := bucket.TicketDir
		indexPath := filepath.Join(ticketPath, "index.md")

		hasIssues := false

		// Check for unique index.md (should only be one per ticket root)
		indexFiles := findIndexFiles(ticketPath, settings.IgnoreDirs, settings.IgnoreGlobs)
		if len(indexFiles) > 1 {
			hasIssues = true
			row := types.NewRow(
				types.MRP("ticket", bucket.TicketID),
				types.MRP("issue", "multiple_index"),
				types.MRP("severity", "warning"),
				types.MRP("message", fmt.Sprintf("Multiple index.md files found (%d), should be only one", len(indexFiles))),
				types.MRP("path", ticketPath),
				types.MRP("index_files", fmt.Sprintf("%v", indexFiles)),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return fmt.Errorf("failed to emit doctor row (multiple_index) for %s: %w", bucket.TicketID, err)
			}
			highestSeverity = maxInt(highestSeverity, 1)
		}

		// Validate index doc (if present in the index).
		if idx := bucket.IndexByPath[indexPath]; idx != nil {
			if idx.ReadErr != nil {
				hasIssues = true
				row := types.NewRow(
					types.MRP("ticket", bucket.TicketID),
					types.MRP("issue", "invalid_frontmatter"),
					types.MRP("severity", "error"),
					types.MRP("message", fmt.Sprintf("Failed to parse frontmatter: %v", idx.ReadErr)),
					types.MRP("path", indexPath),
				)
				if err := gp.AddRow(ctx, row); err != nil {
					return fmt.Errorf("failed to emit doctor row (invalid_frontmatter) for %s: %w", bucket.TicketID, err)
				}
				highestSeverity = maxInt(highestSeverity, 2)
				// Re-parse to preserve taxonomy details when possible.
				_, parseErr := readDocumentFrontmatter(indexPath)
				if parseErr != nil {
					renderFrontmatterParseTaxonomy(ctx, parseErr, indexPath)
				} else {
					renderFrontmatterParseTaxonomy(ctx, idx.ReadErr, indexPath)
				}
			} else if idx.Doc != nil {
				doc := idx.Doc

				// Staleness.
				if !doc.LastUpdated.IsZero() {
					daysSinceUpdate := time.Since(doc.LastUpdated).Hours() / 24
					if daysSinceUpdate > float64(settings.StaleAfterDays) {
						hasIssues = true
						row := types.NewRow(
							types.MRP("ticket", bucket.TicketID),
							types.MRP("issue", "stale"),
							types.MRP("severity", "warning"),
							types.MRP("message", fmt.Sprintf("LastUpdated is %.0f days old (threshold: %d days)", daysSinceUpdate, settings.StaleAfterDays)),
							types.MRP("path", indexPath),
							types.MRP("last_updated", doc.LastUpdated.Format("2006-01-02")),
						)
						if err := gp.AddRow(ctx, row); err != nil {
							return fmt.Errorf("failed to emit doctor row (stale) for %s: %w", bucket.TicketID, err)
						}
						highestSeverity = maxInt(highestSeverity, 1)
						docmgr.RenderTaxonomy(ctx, docmgrctx.NewWorkspaceStale(indexPath, doc.LastUpdated, settings.StaleAfterDays))
					}
				}

				// Required fields.
				if err := doc.Validate(); err != nil {
					hasIssues = true
					row := types.NewRow(
						types.MRP("ticket", bucket.TicketID),
						types.MRP("issue", "missing_required_fields"),
						types.MRP("severity", "error"),
						types.MRP("message", err.Error()),
						types.MRP("path", indexPath),
					)
					if err := gp.AddRow(ctx, row); err != nil {
						return fmt.Errorf("failed to emit doctor row (missing_required_fields) for %s: %w", bucket.TicketID, err)
					}
					highestSeverity = maxInt(highestSeverity, 2)
					for _, field := range missingRequiredFields(doc) {
						detail := fmt.Sprintf("%s is required", field)
						docmgr.RenderTaxonomy(ctx, docmgrctx.NewFrontmatterSchema(indexPath, field, detail, core.SeverityError))
					}
				}

				// Optional fields.
				optionalIssues := []struct {
					field  string
					detail string
				}{}
				if doc.Status == "" {
					optionalIssues = append(optionalIssues, struct {
						field  string
						detail string
					}{field: "Status", detail: "missing Status"})
				}
				if len(doc.Topics) == 0 {
					optionalIssues = append(optionalIssues, struct {
						field  string
						detail string
					}{field: "Topics", detail: "missing Topics"})
				}

				// Vocab checks.
				var unknownTopics []string
				for _, t := range doc.Topics {
					if _, ok := topicSet[t]; !ok && t != "" {
						unknownTopics = append(unknownTopics, t)
					}
				}
				if len(unknownTopics) > 0 {
					hasIssues = true
					row := types.NewRow(
						types.MRP("ticket", bucket.TicketID),
						types.MRP("issue", "unknown_topics"),
						types.MRP("severity", "warning"),
						types.MRP("message", fmt.Sprintf("unknown topics: %v", unknownTopics)),
						types.MRP("path", indexPath),
					)
					if err := gp.AddRow(ctx, row); err != nil {
						return fmt.Errorf("failed to emit doctor row (unknown_topics) for %s: %w", bucket.TicketID, err)
					}
					highestSeverity = maxInt(highestSeverity, 1)
					docmgr.RenderTaxonomy(ctx, docmgrctx.NewVocabularyUnknown(indexPath, "Topics", strings.Join(unknownTopics, ","), topicList))
				}
				if doc.DocType != "" {
					if _, ok := docTypeSet[doc.DocType]; !ok {
						hasIssues = true
						row := types.NewRow(
							types.MRP("ticket", bucket.TicketID),
							types.MRP("issue", "unknown_doc_type"),
							types.MRP("severity", "warning"),
							types.MRP("message", fmt.Sprintf("unknown docType: %s", doc.DocType)),
							types.MRP("path", indexPath),
						)
						if err := gp.AddRow(ctx, row); err != nil {
							return fmt.Errorf("failed to emit doctor row (unknown_doc_type) for %s: %w", bucket.TicketID, err)
						}
						highestSeverity = maxInt(highestSeverity, 1)
						docmgr.RenderTaxonomy(ctx, docmgrctx.NewVocabularyUnknown(indexPath, "DocType", doc.DocType, docTypeList))
					}
				}
				if doc.Intent != "" {
					if _, ok := intentSet[doc.Intent]; !ok {
						hasIssues = true
						row := types.NewRow(
							types.MRP("ticket", bucket.TicketID),
							types.MRP("issue", "unknown_intent"),
							types.MRP("severity", "warning"),
							types.MRP("message", fmt.Sprintf("unknown intent: %s", doc.Intent)),
							types.MRP("path", indexPath),
						)
						if err := gp.AddRow(ctx, row); err != nil {
							return fmt.Errorf("failed to emit doctor row (unknown_intent) for %s: %w", bucket.TicketID, err)
						}
						highestSeverity = maxInt(highestSeverity, 1)
						docmgr.RenderTaxonomy(ctx, docmgrctx.NewVocabularyUnknown(indexPath, "Intent", doc.Intent, intentList))
					}
				}
				if doc.Status != "" {
					if _, ok := statusSet[doc.Status]; !ok {
						hasIssues = true
						row := types.NewRow(
							types.MRP("ticket", bucket.TicketID),
							types.MRP("issue", "unknown_status"),
							types.MRP("severity", "warning"),
							types.MRP("message", fmt.Sprintf("unknown status: %s (valid values: %s; list via 'docmgr vocab list --category status')", doc.Status, statusValidText)),
							types.MRP("path", indexPath),
						)
						if err := gp.AddRow(ctx, row); err != nil {
							return fmt.Errorf("failed to emit doctor row (unknown_status) for %s: %w", bucket.TicketID, err)
						}
						highestSeverity = maxInt(highestSeverity, 1)
						docmgr.RenderTaxonomy(ctx, docmgrctx.NewVocabularyUnknown(indexPath, "Status", doc.Status, statusList))
					}
				}

				// RelatedFiles existence checks using a doc-anchored resolver (Spec §7.3).
				for _, rf := range doc.RelatedFiles {
					if strings.TrimSpace(rf.Path) == "" {
						continue
					}
					if strings.TrimSpace(rf.Note) == "" {
						hasIssues = true
						row := types.NewRow(
							types.MRP("ticket", bucket.TicketID),
							types.MRP("issue", "missing_related_file_note"),
							types.MRP("severity", "warning"),
							types.MRP("message", fmt.Sprintf("related file has no Note: %s", rf.Path)),
							types.MRP("path", indexPath),
						)
						if err := gp.AddRow(ctx, row); err != nil {
							return fmt.Errorf("failed to emit doctor row (missing_related_file_note) for %s: %w", bucket.TicketID, err)
						}
						highestSeverity = maxInt(highestSeverity, 1)
					}

					resolver := paths.NewResolver(paths.ResolverOptions{
						DocsRoot:  ws.Context().Root,
						DocPath:   indexPath,
						ConfigDir: ws.Context().ConfigDir,
						RepoRoot:  ws.Context().RepoRoot,
					})
					n := resolver.Normalize(rf.Path)
					if !n.Exists {
						hasIssues = true
						row := types.NewRow(
							types.MRP("ticket", bucket.TicketID),
							types.MRP("issue", "missing_related_file"),
							types.MRP("severity", "warning"),
							types.MRP("message", fmt.Sprintf("related file not found: %s", rf.Path)),
							types.MRP("path", indexPath),
						)
						if err := gp.AddRow(ctx, row); err != nil {
							return fmt.Errorf("failed to emit doctor row (missing_related_file) for %s: %w", bucket.TicketID, err)
						}
						highestSeverity = maxInt(highestSeverity, 1)
						docmgr.RenderTaxonomy(ctx, docmgrctx.NewRelatedFileMissing(indexPath, rf.Path, rf.Note))
					}
				}

				if len(optionalIssues) > 0 {
					hasIssues = true
					for _, issue := range optionalIssues {
						row := types.NewRow(
							types.MRP("ticket", bucket.TicketID),
							types.MRP("issue", "missing_field"),
							types.MRP("severity", "warning"),
							types.MRP("message", issue.detail),
							types.MRP("path", indexPath),
						)
						if err := gp.AddRow(ctx, row); err != nil {
							return fmt.Errorf("failed to emit doctor row (missing_field) for %s: %w", bucket.TicketID, err)
						}
						highestSeverity = maxInt(highestSeverity, 1)
						docmgr.RenderTaxonomy(ctx, docmgrctx.NewFrontmatterSchema(indexPath, issue.field, issue.detail, core.SeverityWarning))
					}
				}
			}
		}

		// Validate all markdown docs in the ticket bucket (invalid frontmatter + numeric prefix policy).
		for _, h := range bucket.Docs {
			// Only consider markdown files (index builder is md-only, but keep the guard).
			if !strings.HasSuffix(strings.ToLower(filepath.Base(h.Path)), ".md") {
				continue
			}
			// Skip root-level control files.
			dir := filepath.Dir(h.Path)
			isRootLevel := filepath.Clean(dir) == filepath.Clean(ticketPath)
			if isRootLevel {
				bn := filepath.Base(h.Path)
				if bn == "index.md" || bn == "README.md" || bn == "tasks.md" || bn == "changelog.md" {
					continue
				}
			}

			// invalid frontmatter
			if h.ReadErr != nil && filepath.Base(h.Path) != "index.md" {
				hasIssues = true
				// Re-parse to preserve taxonomy details when possible.
				_, parseErr := readDocumentFrontmatter(h.Path)
				msgErr := h.ReadErr
				if parseErr != nil {
					msgErr = parseErr
				}
				row := types.NewRow(
					types.MRP("ticket", bucket.TicketID),
					types.MRP("issue", "invalid_frontmatter"),
					types.MRP("severity", "error"),
					types.MRP("message", fmt.Sprintf("Failed to parse frontmatter: %v", msgErr)),
					types.MRP("path", h.Path),
				)
				if err := gp.AddRow(ctx, row); err != nil {
					return fmt.Errorf("failed to emit doctor row (invalid_frontmatter) for %s: %w", h.Path, err)
				}
				highestSeverity = maxInt(highestSeverity, 2)
				renderFrontmatterParseTaxonomy(ctx, msgErr, h.Path)
			}

			// numeric prefix policy (subdirectory files only).
			if !isRootLevel {
				bn := filepath.Base(h.Path)
				if !prefixRe.MatchString(bn) {
					hasIssues = true
					row := types.NewRow(
						types.MRP("ticket", bucket.TicketID),
						types.MRP("issue", "missing_numeric_prefix"),
						types.MRP("severity", "warning"),
						types.MRP("message", "file without numeric prefix"),
						types.MRP("path", h.Path),
					)
					if err := gp.AddRow(ctx, row); err != nil {
						return fmt.Errorf("failed to emit doctor row (missing_numeric_prefix) for %s: %w", h.Path, err)
					}
					highestSeverity = maxInt(highestSeverity, 1)
				}
			}
		}

		if !hasIssues {
			row := types.NewRow(
				types.MRP("ticket", bucket.TicketID),
				types.MRP("issue", "none"),
				types.MRP("severity", "ok"),
				types.MRP("message", "All checks passed"),
				types.MRP("path", ticketPath),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return fmt.Errorf("failed to emit doctor summary row for %s: %w", bucket.TicketID, err)
			}
		}
	}

	// Enforce fail-on behavior
	if diagRenderer != nil && settings.DiagnosticsJSON != "" {
		if err := writeDiagnosticsJSON(diagRenderer, settings.DiagnosticsJSON); err != nil {
			return err
		}
	}
	threshold := severityThreshold(settings.FailOn)
	if threshold >= 0 && highestSeverity >= threshold && threshold > 0 {
		return fmt.Errorf("doctor failed: severity >= %s", settings.FailOn)
	}

	return nil
}

type doctorTicketBucket struct {
	TicketID    string
	TicketDir   string
	Docs        []workspace.DocHandle
	IndexByPath map[string]*workspace.DocHandle
}

func groupDoctorDocsByTicket(docsRoot string, docs []workspace.DocHandle) []doctorTicketBucket {
	docsRoot = filepath.Clean(strings.TrimSpace(docsRoot))
	type key struct {
		dir string
	}
	m := map[key]*doctorTicketBucket{}
	order := []key{}

	for _, h := range docs {
		abs := filepath.Clean(strings.TrimSpace(h.Path))
		if abs == "" {
			continue
		}
		ticketDir, ticketID := inferTicketDirAndID(docsRoot, abs, h)
		if ticketDir == "" {
			continue
		}
		k := key{dir: ticketDir}
		b, ok := m[k]
		if !ok {
			b = &doctorTicketBucket{
				TicketID:    ticketID,
				TicketDir:   ticketDir,
				Docs:        []workspace.DocHandle{},
				IndexByPath: map[string]*workspace.DocHandle{},
			}
			m[k] = b
			order = append(order, k)
		}
		// Prefer non-empty ticket IDs as we see more docs.
		if strings.TrimSpace(b.TicketID) == "" && strings.TrimSpace(ticketID) != "" {
			b.TicketID = ticketID
		}
		b.Docs = append(b.Docs, h)
		b.IndexByPath[abs] = &b.Docs[len(b.Docs)-1]
	}

	out := make([]doctorTicketBucket, 0, len(order))
	for _, k := range order {
		if b := m[k]; b != nil {
			out = append(out, *b)
		}
	}
	return out
}

func inferTicketDirAndID(docsRoot string, absDocPath string, h workspace.DocHandle) (string, string) {
	docsRoot = filepath.Clean(strings.TrimSpace(docsRoot))
	absDocPath = filepath.Clean(strings.TrimSpace(absDocPath))
	if docsRoot == "" || absDocPath == "" {
		return "", ""
	}
	rel, err := filepath.Rel(docsRoot, absDocPath)
	if err != nil {
		return "", ""
	}
	rel = filepath.Clean(rel)
	if rel == "." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." {
		return "", ""
	}
	parts := strings.Split(rel, string(filepath.Separator))
	// Need at least YYYY/MM/DD/<TICKET--slug>/...
	if len(parts) < 4 {
		return "", ""
	}
	ticketDirName := strings.TrimSpace(parts[3])
	if ticketDirName == "" {
		return "", ""
	}
	ticketDir := filepath.Join(docsRoot, parts[0], parts[1], parts[2], ticketDirName)

	// Best-effort ticket ID:
	// 1) from parsed doc (preferred)
	ticketID := ""
	if h.Doc != nil && strings.TrimSpace(h.Doc.Ticket) != "" {
		ticketID = strings.TrimSpace(h.Doc.Ticket)
	}
	// 2) from directory "<TICKET>--<slug>"
	if ticketID == "" {
		if i := strings.Index(ticketDirName, "--"); i > 0 {
			ticketID = strings.TrimSpace(ticketDirName[:i])
		}
	}
	// 3) fallback to directory name
	if ticketID == "" {
		ticketID = ticketDirName
	}

	return ticketDir, ticketID
}

func shouldSkipDoctorDoc(docsRoot string, absPath string, ignoreDirNames []string, ignoreGlobs []string) bool {
	absPath = filepath.Clean(strings.TrimSpace(absPath))
	if absPath == "" {
		return true
	}
	base := filepath.Base(absPath)
	if containsString(ignoreDirNames, base) {
		return true
	}
	// Apply globs to both the absolute path and the docs-root-relative path (if possible),
	// plus the basename for convenience.
	if matchesAnyGlob(ignoreGlobs, base) || matchesAnyGlob(ignoreGlobs, absPath) {
		return true
	}
	docsRoot = filepath.Clean(strings.TrimSpace(docsRoot))
	if docsRoot != "" {
		if rel, err := filepath.Rel(docsRoot, absPath); err == nil {
			rel = filepath.Clean(rel)
			if rel != "." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != ".." {
				if matchesAnyGlob(ignoreGlobs, rel) {
					return true
				}
			}
		}
	}
	return false
}

func matchesTicketDir(ticketID string, base string) bool {
	ticketID = strings.TrimSpace(ticketID)
	base = strings.TrimSpace(base)
	if ticketID == "" || base == "" {
		return false
	}
	return base == ticketID ||
		strings.HasPrefix(base, ticketID+"--") ||
		strings.HasPrefix(base, ticketID+"-")
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

func missingRequiredFields(doc *models.Document) []string {
	fields := []string{}
	if strings.TrimSpace(doc.Title) == "" {
		fields = append(fields, "Title")
	}
	if strings.TrimSpace(doc.Ticket) == "" {
		fields = append(fields, "Ticket")
	}
	if strings.TrimSpace(doc.DocType) == "" {
		fields = append(fields, "DocType")
	}
	return fields
}

func renderFrontmatterParseTaxonomy(ctx context.Context, err error, path string) {
	if err == nil {
		return
	}
	if tax, ok := core.AsTaxonomy(err); ok && tax != nil {
		docmgr.RenderTaxonomy(ctx, tax)
		return
	}
	docmgr.RenderTaxonomy(ctx, docmgrctx.NewFrontmatterParse(path, 0, 0, "", err.Error(), err))
}

func writeDiagnosticsJSON(renderer *docmgr.Renderer, destination string) error {
	if renderer == nil || destination == "" {
		return nil
	}
	data, err := renderer.JSON()
	if err != nil {
		return fmt.Errorf("failed to marshal diagnostics JSON: %w", err)
	}
	if len(data) == 0 {
		data = []byte("[]")
	}
	if data[len(data)-1] != '\n' {
		data = append(data, '\n')
	}
	if destination == "-" {
		if _, err := os.Stdout.Write(data); err != nil {
			return fmt.Errorf("failed to write diagnostics JSON to stdout: %w", err)
		}
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return fmt.Errorf("failed to create diagnostics JSON directory: %w", err)
	}
	if err := os.WriteFile(destination, data, 0o644); err != nil {
		return fmt.Errorf("failed to write diagnostics JSON: %w", err)
	}
	return nil
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

// validateSingleDoc validates one markdown file (frontmatter parse + required fields + vocab warnings).
func validateSingleDoc(
	ctx context.Context,
	docPath string,
	topicSet map[string]struct{},
	topicList []string,
	docTypeSet map[string]struct{},
	docTypeList []string,
	intentSet map[string]struct{},
	intentList []string,
	statusSet map[string]struct{},
	statusList []string,
	statusValidText string,
	gp glazedMiddlewares.Processor,
) (int, error) {
	highestSeverity := 0

	doc, err := readDocumentFrontmatter(docPath)
	if err != nil {
		row := types.NewRow(
			types.MRP("ticket", ""),
			types.MRP("issue", "invalid_frontmatter"),
			types.MRP("severity", "error"),
			types.MRP("message", fmt.Sprintf("Failed to parse frontmatter: %v", err)),
			types.MRP("path", docPath),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return highestSeverity, fmt.Errorf("failed to emit doctor row (invalid_frontmatter) for %s: %w", docPath, err)
		}
		highestSeverity = maxInt(highestSeverity, 2)
		renderFrontmatterParseTaxonomy(ctx, err, docPath)
		return highestSeverity, nil
	}

	// Required fields
	if err := doc.Validate(); err != nil {
		row := types.NewRow(
			types.MRP("ticket", doc.Ticket),
			types.MRP("issue", "missing_required_fields"),
			types.MRP("severity", "error"),
			types.MRP("message", err.Error()),
			types.MRP("path", docPath),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return highestSeverity, fmt.Errorf("failed to emit doctor row (missing_required_fields) for %s: %w", docPath, err)
		}
		highestSeverity = maxInt(highestSeverity, 2)
		for _, field := range missingRequiredFields(doc) {
			detail := fmt.Sprintf("%s is required", field)
			docmgr.RenderTaxonomy(ctx, docmgrctx.NewFrontmatterSchema(docPath, field, detail, core.SeverityError))
		}
	}

	// Optional fields: Status, Topics
	if doc.Status == "" {
		row := types.NewRow(
			types.MRP("ticket", doc.Ticket),
			types.MRP("issue", "missing_field"),
			types.MRP("severity", "warning"),
			types.MRP("message", "missing Status"),
			types.MRP("path", docPath),
		)
		_ = gp.AddRow(ctx, row)
		highestSeverity = maxInt(highestSeverity, 1)
		docmgr.RenderTaxonomy(ctx, docmgrctx.NewFrontmatterSchema(docPath, "Status", "missing Status", core.SeverityWarning))
	}
	if len(doc.Topics) == 0 {
		row := types.NewRow(
			types.MRP("ticket", doc.Ticket),
			types.MRP("issue", "missing_field"),
			types.MRP("severity", "warning"),
			types.MRP("message", "missing Topics"),
			types.MRP("path", docPath),
		)
		_ = gp.AddRow(ctx, row)
		highestSeverity = maxInt(highestSeverity, 1)
		docmgr.RenderTaxonomy(ctx, docmgrctx.NewFrontmatterSchema(docPath, "Topics", "missing Topics", core.SeverityWarning))
	}

	// Vocabulary checks
	if len(doc.Topics) > 0 {
		var unknownTopics []string
		for _, t := range doc.Topics {
			if _, ok := topicSet[t]; !ok && t != "" {
				unknownTopics = append(unknownTopics, t)
			}
		}
		if len(unknownTopics) > 0 {
			row := types.NewRow(
				types.MRP("ticket", doc.Ticket),
				types.MRP("issue", "unknown_topics"),
				types.MRP("severity", "warning"),
				types.MRP("message", fmt.Sprintf("unknown topics: %v", unknownTopics)),
				types.MRP("path", docPath),
			)
			_ = gp.AddRow(ctx, row)
			highestSeverity = maxInt(highestSeverity, 1)
			docmgr.RenderTaxonomy(ctx, docmgrctx.NewVocabularyUnknown(docPath, "Topics", strings.Join(unknownTopics, ","), topicList))
		}
	}
	if doc.DocType != "" {
		if _, ok := docTypeSet[doc.DocType]; !ok {
			row := types.NewRow(
				types.MRP("ticket", doc.Ticket),
				types.MRP("issue", "unknown_doc_type"),
				types.MRP("severity", "warning"),
				types.MRP("message", fmt.Sprintf("unknown docType: %s", doc.DocType)),
				types.MRP("path", docPath),
			)
			_ = gp.AddRow(ctx, row)
			highestSeverity = maxInt(highestSeverity, 1)
			docmgr.RenderTaxonomy(ctx, docmgrctx.NewVocabularyUnknown(docPath, "DocType", doc.DocType, docTypeList))
		}
	}
	if doc.Intent != "" {
		if _, ok := intentSet[doc.Intent]; !ok {
			row := types.NewRow(
				types.MRP("ticket", doc.Ticket),
				types.MRP("issue", "unknown_intent"),
				types.MRP("severity", "warning"),
				types.MRP("message", fmt.Sprintf("unknown intent: %s", doc.Intent)),
				types.MRP("path", docPath),
			)
			_ = gp.AddRow(ctx, row)
			highestSeverity = maxInt(highestSeverity, 1)
			docmgr.RenderTaxonomy(ctx, docmgrctx.NewVocabularyUnknown(docPath, "Intent", doc.Intent, intentList))
		}
	}
	if doc.Status != "" {
		if _, ok := statusSet[doc.Status]; !ok {
			row := types.NewRow(
				types.MRP("ticket", doc.Ticket),
				types.MRP("issue", "unknown_status"),
				types.MRP("severity", "warning"),
				types.MRP("message", fmt.Sprintf("unknown status: %s (valid values: %s; list via 'docmgr vocab list --category status')", doc.Status, statusValidText)),
				types.MRP("path", docPath),
			)
			_ = gp.AddRow(ctx, row)
			highestSeverity = maxInt(highestSeverity, 1)
			docmgr.RenderTaxonomy(ctx, docmgrctx.NewVocabularyUnknown(docPath, "Status", doc.Status, statusList))
		}
	}

	// Success row
	if highestSeverity == 0 {
		row := types.NewRow(
			types.MRP("ticket", doc.Ticket),
			types.MRP("issue", "none"),
			types.MRP("severity", "ok"),
			types.MRP("message", "All checks passed"),
			types.MRP("path", docPath),
		)
		_ = gp.AddRow(ctx, row)
	}

	return highestSeverity, nil
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
