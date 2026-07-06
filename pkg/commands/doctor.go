package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/go-go-golems/docmgr/internal/documents"
	"github.com/go-go-golems/docmgr/internal/paths"
	"github.com/go-go-golems/docmgr/internal/templates"
	"github.com/go-go-golems/docmgr/internal/tickets"
	"github.com/go-go-golems/docmgr/internal/workspace"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/core"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/docmgr"
	"github.com/go-go-golems/docmgr/pkg/diagnostics/docmgrctx"
	"github.com/go-go-golems/docmgr/pkg/models"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
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
	Root            string   `glazed:"root"`
	Ticket          string   `glazed:"ticket"`
	Doc             string   `glazed:"doc"`
	All             bool     `glazed:"all"`
	IgnoreDirs      []string `glazed:"ignore-dir"`
	IgnoreGlobs     []string `glazed:"ignore-glob"`
	StaleAfterDays  int      `glazed:"stale-after"`
	FailOn          string   `glazed:"fail-on"`
	DiagnosticsJSON string   `glazed:"diagnostics-json"`
	Fix             bool     `glazed:"fix"`
	FixAnchors      bool     `glazed:"fix-anchors"`
	Details         bool     `glazed:"details"`
	IncludeSources  bool     `glazed:"include-sources"`
	// Schema printing flags (human mode only)
	PrintTemplateSchema bool   `glazed:"print-template-schema"`
	SchemaFormat        string `glazed:"schema-format"`
}

// Built-in vocabulary values doctor always treats as known, so a fresh
// workspace produces zero unknown_* warnings for scaffold defaults. DocTypes
// cover the embedded template/scaffold set (internal/templates/embedded).
var (
	doctorBuiltinDocTypes = []string{
		"index", "design-doc", "reference", "playbook", "analysis", "til",
		"skill", "log", "code-review", "script", "task-list", "tutorial",
		"working-note",
	}
	doctorBuiltinIntents = []string{
		"long-term", "ticket-specific", "only-during-ticket", "throwaway", "short-term",
	}
	doctorBuiltinStatuses = []string{
		"draft", "active", "review", "complete", "archived",
	}
)

// doctorVocab bundles the known vocabulary sets used for unknown_* checks.
type doctorVocab struct {
	topicSet    map[string]struct{}
	topicList   []string
	docTypeSet  map[string]struct{}
	docTypeList []string
	intentSet   map[string]struct{}
	intentList  []string
	statusSet   map[string]struct{}
	statusList  []string
	// configured is true when a vocabulary file exists on disk. When false,
	// doctor emits a single info-level 'no vocabulary configured' finding
	// instead of per-value unknown_* warnings.
	configured bool
}

func newDoctorVocab(vocab *models.Vocabulary) *doctorVocab {
	dv := &doctorVocab{
		topicSet:   map[string]struct{}{},
		docTypeSet: map[string]struct{}{},
		intentSet:  map[string]struct{}{},
		statusSet:  map[string]struct{}{},
	}
	add := func(set map[string]struct{}, list *[]string, slug string) {
		slug = strings.TrimSpace(slug)
		if slug == "" {
			return
		}
		if _, ok := set[slug]; ok {
			return
		}
		set[slug] = struct{}{}
		*list = append(*list, slug)
	}
	for _, it := range vocab.Topics {
		add(dv.topicSet, &dv.topicList, it.Slug)
	}
	for _, it := range vocab.DocTypes {
		add(dv.docTypeSet, &dv.docTypeList, it.Slug)
	}
	for _, it := range vocab.Intent {
		add(dv.intentSet, &dv.intentList, it.Slug)
	}
	for _, it := range vocab.Status {
		add(dv.statusSet, &dv.statusList, it.Slug)
	}
	// Built-ins are always known.
	for _, s := range doctorBuiltinDocTypes {
		add(dv.docTypeSet, &dv.docTypeList, s)
	}
	for _, s := range doctorBuiltinIntents {
		add(dv.intentSet, &dv.intentList, s)
	}
	for _, s := range doctorBuiltinStatuses {
		add(dv.statusSet, &dv.statusList, s)
	}
	if path, err := workspace.ResolveVocabularyPath(); err == nil {
		if _, statErr := os.Stat(path); statErr == nil {
			dv.configured = true
		}
	}
	return dv
}

// vocabRemediation renders the canonical remediation hint for an unknown
// vocabulary value.
func vocabRemediation(category string) string {
	return fmt.Sprintf("add via 'docmgr vocab add --category %s --slug <slug> --description \"TODO\"' (valid categories: topics, docTypes, intent, status)", category)
}

const doctorNoVocabularyMessage = "no vocabulary configured; skipping unknown-value checks (run 'docmgr init' to seed defaults or 'docmgr vocab add --category <topics|docTypes|intent|status> --slug <slug> --description \"TODO\"')"

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
    quote strings containing ':' or '#', and use 'docmgr meta update' to rewrite fields safely
    (or run 'docmgr doctor --fix' to apply safe auto-fixes).
  • missing_required_fields — Title/Summary/DocType/etc. are missing. Run
    'docmgr meta update --ticket T --field FieldName --value ...' (or edit frontmatter) to add them.
  • missing_index — Ticket directories without index.md. Re-run 'docmgr ticket create'
    or copy the template back into place.
  • unknown_topics / unknown_status — Value not present in vocabulary.yaml. Either add it via
    'docmgr vocab add --category topics --slug your-topic --description "TODO"'
    (valid categories: topics, docTypes, intent, status) or update the doc’s fields.
  • stale — No document in the ticket was updated within '--stale-after' days (default 30).
    Review the ticket, make an update, or pass '--stale-after N' for a different cadence.

Scope and output:
  • RelatedFiles, vocabulary, and staleness checks run on every document in a ticket,
    not just index.md. Staleness is per ticket: it fires only when no doc was updated recently.
  • Documents under sources/ are skipped by default (imported material); opt in with '--include-sources'.
  • Multi-ticket runs print a one-line rollup per ticket; pass '--details' (or '--ticket ID')
    for the full per-issue report.

Tips:
  • '--fix' applies safe fixes: frontmatter auto-repair (same fixes as
    'validate frontmatter --auto-fix', with .bak backups) and anchor migration.
  • '--fix-anchors' migrates only legacy RelatedFiles paths to explicit anchors
    (repo://pkg/foo.go, ws://<member>/<rel> for go.work siblings, docs://..., abs:///...).
    Only entries that resolve to an existing file are rewritten; the rest are left
    as legacy with an 'anchor_migration_skipped' warning.
  • Use '--fail-on warning' (or 'error') to make CI fail when issues are detected.
  • '--diagnostics-json path' captures rule results as JSON (use '-' for stdout) for CI/automation.
  • '--ignore-glob' is handy for suppressing known noisy paths; the command also reads patterns from
    both repository and docs-root .docmgrignore files.

Examples:
  # Validate a specific ticket workspace (full details by default)
  docmgr doctor --ticket MEN-3475

  # Validate all tickets (rollup summary; add --details for everything)
  docmgr doctor --all

  # Apply safe fixes (frontmatter auto-repair + anchor migration)
  docmgr doctor --ticket MEN-3475 --fix

  # Tighten staleness and fail CI on warnings
  docmgr doctor --all --stale-after 14 --fail-on warning

  # Ignore multiple dirs/globs (repeat the flags)
  docmgr doctor --all --ignore-dir archive --ignore-glob "*.bak" --ignore-glob "*.tmp"

  # Emit diagnostics JSON for scripts/CI
  docmgr doctor --ticket MEN-3475 --diagnostics-json - --with-glaze-output --output json
`),
			cmds.WithFlags(
				fields.New(
					"root",
					fields.TypeString,
					fields.WithHelp("Root directory for docs"),
					fields.WithDefault("ttmp"),
				),
				fields.New(
					"print-template-schema",
					fields.TypeBool,
					fields.WithHelp("Print template schema after output (human mode only)"),
					fields.WithDefault(false),
				),
				fields.New(
					"schema-format",
					fields.TypeString,
					fields.WithHelp("Template schema output format: json|yaml"),
					fields.WithDefault("json"),
				),
				fields.New(
					"ticket",
					fields.TypeString,
					fields.WithHelp("Check specific ticket"),
					fields.WithDefault(""),
				),
				fields.New(
					"doc",
					fields.TypeString,
					fields.WithHelp("Validate a single markdown file (overrides --ticket/--all)"),
					fields.WithDefault(""),
				),
				fields.New(
					"all",
					fields.TypeBool,
					fields.WithHelp("Check all tickets"),
					fields.WithDefault(false),
				),
				fields.New(
					"ignore-dir",
					fields.TypeStringList,
					fields.WithHelp("Directory names at root or within tickets to ignore (can be repeated)"),
					fields.WithDefault([]string{}),
				),
				fields.New(
					"ignore-glob",
					fields.TypeStringList,
					fields.WithHelp("Glob patterns (applied to path or basename) to ignore during scanning"),
					fields.WithDefault([]string{}),
				),
				fields.New(
					"stale-after",
					fields.TypeInteger,
					fields.WithHelp("Days after which a document is considered stale (default 30)"),
					fields.WithDefault(30),
				),
				fields.New(
					"fail-on",
					fields.TypeString,
					fields.WithHelp("Fail with non-zero exit on severity: none|warning|error (default none)"),
					fields.WithDefault("none"),
				),
				fields.New(
					"diagnostics-json",
					fields.TypeString,
					fields.WithHelp("Write diagnostics rule output to JSON (file path or '-' for stdout)"),
					fields.WithDefault(""),
				),
				fields.New(
					"fix",
					fields.TypeBool,
					fields.WithHelp("Apply safe fixes: frontmatter auto-repair (creates .bak backups) and anchor migration, then re-validate."),
					fields.WithDefault(false),
				),
				fields.New(
					"fix-anchors",
					fields.TypeBool,
					fields.WithHelp("Migrate legacy RelatedFiles paths to explicit anchors (repo://, ws://, docs://, abs://). Subset of --fix. Entries that don't resolve to an existing file are left as-is with a warning."),
					fields.WithDefault(false),
				),
				fields.New(
					"details",
					fields.TypeBool,
					fields.WithHelp("Print the full per-issue report instead of the per-ticket rollup (single-ticket runs always show details)."),
					fields.WithDefault(false),
				),
				fields.New(
					"include-sources",
					fields.TypeBool,
					fields.WithHelp("Also check documents under sources/ (imported material; skipped by default)."),
					fields.WithDefault(false),
				),
			),
		),
	}, nil
}

func (c *DoctorCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedValues *values.Values,
	gp glazedMiddlewares.Processor,
) error {
	settings := &DoctorSettings{}
	if err := parsedValues.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Apply config root if present
	settings.Root = workspace.ResolveRoot(settings.Root)

	// Diagnostics renderer: collects for JSON output when requested, and
	// stays text-silent in rollup mode so multi-ticket runs really are
	// summary-first (the per-finding text otherwise streams to stderr).
	detailMode := settings.Details ||
		strings.TrimSpace(settings.Doc) != "" ||
		(strings.TrimSpace(settings.Ticket) != "" && !settings.All)
	rendererOpts := []docmgr.Option{docmgr.WithTextOutput(detailMode)}
	if settings.DiagnosticsJSON != "" {
		rendererOpts = append(rendererOpts, docmgr.WithCollector())
	}
	diagRenderer := docmgr.NewRenderer(rendererOpts...)
	ctx = docmgr.ContextWithRenderer(ctx, diagRenderer)

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

	// Load vocabulary for validation. Invalid vocabulary is a workspace problem,
	// but it should surface as a normal command error rather than a panic.
	vocab, err := LoadVocabulary()
	if err != nil {
		return fmt.Errorf("failed to load vocabulary: %w", err)
	}
	if vocab == nil {
		return fmt.Errorf("failed to load vocabulary: vocabulary is nil")
	}
	dv := newDoctorVocab(vocab)

	// Single-file mode: validate one explicitly requested doc and return. This
	// intentionally ignores workspace ignore policy: if the user names a file,
	// doctor validates that file directly.
	if settings.Doc != "" {
		docPath, err := resolveDocRef(ctx, nil, settings.Root, settings.Doc)
		if err != nil {
			return err
		}
		sev, err := validateSingleDoc(ctx, docPath, dv, gp)
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
	ignoreMatcher := ws.IgnoreMatcher()
	shouldSkipPath := func(relPath, base string, isDir bool) bool {
		if isDir && containsString(settings.IgnoreDirs, base) {
			return true
		}
		full := relPath
		if !filepath.IsAbs(full) {
			full = filepath.Join(settings.Root, relPath)
			if relPath == "." {
				full = settings.Root
			}
		}
		if matchesDoctorIgnoreGlob(settings.IgnoreGlobs, base) || matchesDoctorIgnoreGlob(settings.IgnoreGlobs, full) {
			return true
		}
		if ignoreMatcher != nil && ignoreMatcher.Ignore(full, isDir) {
			return true
		}
		return false
	}
	skipFn := func(relPath, base string) bool {
		return shouldSkipPath(relPath, base, true)
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

	// Determine scope: default is repo-wide unless --ticket is provided. Ticket
	// refs use the same forgiving resolver as the other user-facing ticket
	// commands before ScopeTicket compiles to an exact ticket_id SQL predicate.
	scope := workspace.Scope{Kind: workspace.ScopeRepo}
	requestedTicket := strings.TrimSpace(settings.Ticket)
	if requestedTicket != "" {
		res, err := tickets.Resolve(ctx, ws, requestedTicket)
		if err != nil {
			return fmt.Errorf("failed to resolve ticket %q: %w", requestedTicket, err)
		}
		scope = workspace.Scope{Kind: workspace.ScopeTicket, TicketID: res.TicketID}
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
			// sources/ docs (imported material) are skipped unless --include-sources.
			IncludeArchivedPath: true,
			IncludeScriptsPath:  true,
			IncludeSourcesPath:  settings.IncludeSources,
			IncludeControlDocs:  true,
			OrderBy:             workspace.OrderByPath,
		},
	}

	qr, err := ws.QueryDocs(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query docs: %w", err)
	}

	// The workspace index already prunes .docmgrignore matches before parsing.
	// Preserve explicit doctor CLI --ignore-dir/--ignore-glob as a command-level
	// filter for callers that still use those flags.
	filtered := qr.Docs
	if len(settings.IgnoreDirs) > 0 || len(settings.IgnoreGlobs) > 0 {
		filtered = make([]workspace.DocHandle, 0, len(qr.Docs))
		for _, h := range qr.Docs {
			if shouldSkipDoctorCLIPath(settings.Root, h.Path, settings.IgnoreDirs, settings.IgnoreGlobs) {
				continue
			}
			filtered = append(filtered, h)
		}
	}

	// Group by ticket directory inferred from ttmp layout.
	tickets := groupDoctorDocsByTicket(settings.Root, filtered)
	if requestedTicket != "" && !settings.All && len(tickets) == 0 {
		return fmt.Errorf("doctor checked zero documents for ticket %q", requestedTicket)
	}

	// Safe fixes (--fix / --fix-anchors): rewrite documents before validation,
	// then rebuild the index so the checks below see the fixed state.
	// --fix = frontmatter auto-repair + anchor migration; --fix-anchors is the
	// anchor-only subset (kept as an alias).
	if settings.Fix || settings.FixAnchors {
		migrated := false

		// Frontmatter auto-repair (--fix only): same safe fixes as
		// 'validate frontmatter --auto-fix', with .bak backups.
		if settings.Fix {
			for _, bucket := range tickets {
				for _, h := range bucket.Docs {
					if h.ReadErr == nil {
						continue
					}
					fixed, fixes, fixErr := autoFixDocFrontmatter(h.Path)
					if !fixed && fixErr == nil {
						continue // no safe fix available; validation will report it
					}
					if fixed {
						// The file changed on disk either way; re-index below.
						migrated = true
					}
					if fixErr != nil {
						row := types.NewRow(
							types.MRP("ticket", bucket.TicketID),
							types.MRP("issue", "frontmatter_fix_failed"),
							types.MRP("severity", "warning"),
							types.MRP("message", fmt.Sprintf("auto-fix did not produce parseable frontmatter: %v", fixErr)),
							types.MRP("path", h.Path),
						)
						if err := gp.AddRow(ctx, row); err != nil {
							return fmt.Errorf("failed to emit doctor row (frontmatter_fix_failed) for %s: %w", h.Path, err)
						}
						highestSeverity = maxInt(highestSeverity, 1)
						continue
					}
					migrated = true
					row := types.NewRow(
						types.MRP("ticket", bucket.TicketID),
						types.MRP("issue", "frontmatter_fixed"),
						types.MRP("severity", "ok"),
						types.MRP("message", fmt.Sprintf("auto-fixed frontmatter (%s); backup written to %s.bak", strings.Join(fixes, "; "), filepath.Base(h.Path))),
						types.MRP("path", h.Path),
					)
					if err := gp.AddRow(ctx, row); err != nil {
						return fmt.Errorf("failed to emit doctor row (frontmatter_fixed) for %s: %w", h.Path, err)
					}
				}
			}
		}

		for _, bucket := range tickets {
			for _, h := range bucket.Docs {
				if h.ReadErr != nil || h.Doc == nil {
					continue
				}
				changed, skipped, err := migrateDocAnchors(ws, h.Path)
				if err != nil {
					return fmt.Errorf("failed to migrate anchors for %s: %w", h.Path, err)
				}
				for _, skip := range skipped {
					row := types.NewRow(
						types.MRP("ticket", bucket.TicketID),
						types.MRP("issue", "anchor_migration_skipped"),
						types.MRP("severity", "warning"),
						types.MRP("message", fmt.Sprintf("legacy related file left as-is (does not resolve to an existing file): %s", skip)),
						types.MRP("path", h.Path),
					)
					if err := gp.AddRow(ctx, row); err != nil {
						return fmt.Errorf("failed to emit doctor row (anchor_migration_skipped) for %s: %w", h.Path, err)
					}
					highestSeverity = maxInt(highestSeverity, 1)
				}
				if changed > 0 {
					migrated = true
					row := types.NewRow(
						types.MRP("ticket", bucket.TicketID),
						types.MRP("issue", "anchors_migrated"),
						types.MRP("severity", "ok"),
						types.MRP("message", fmt.Sprintf("migrated %d related file path(s) to explicit anchors", changed)),
						types.MRP("path", h.Path),
					)
					if err := gp.AddRow(ctx, row); err != nil {
						return fmt.Errorf("failed to emit doctor row (anchors_migrated) for %s: %w", h.Path, err)
					}
				}
			}
		}
		if migrated {
			// Re-index and re-query so validations reflect the rewritten docs.
			if err := ws.InitIndex(ctx, workspace.BuildIndexOptions{IncludeBody: false}); err != nil {
				return fmt.Errorf("failed to rebuild workspace index after anchor migration: %w", err)
			}
			qr, err = ws.QueryDocs(ctx, query)
			if err != nil {
				return fmt.Errorf("failed to query docs after anchor migration: %w", err)
			}
			filtered = qr.Docs
			if len(settings.IgnoreDirs) > 0 || len(settings.IgnoreGlobs) > 0 {
				filtered = make([]workspace.DocHandle, 0, len(qr.Docs))
				for _, h := range qr.Docs {
					if shouldSkipDoctorCLIPath(settings.Root, h.Path, settings.IgnoreDirs, settings.IgnoreGlobs) {
						continue
					}
					filtered = append(filtered, h)
				}
			}
			tickets = groupDoctorDocsByTicket(settings.Root, filtered)
		}
	}

	// Vocabulary bootstrap: when no vocabulary file exists at all, emit ONE
	// info-level finding and skip per-value unknown_* warnings entirely.
	if !dv.configured {
		row := types.NewRow(
			types.MRP("ticket", ""),
			types.MRP("issue", "no_vocabulary"),
			types.MRP("severity", "info"),
			types.MRP("message", doctorNoVocabularyMessage),
			types.MRP("path", ""),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return fmt.Errorf("failed to emit doctor row (no_vocabulary): %w", err)
		}
	}

	// Per-ticket validations. RelatedFiles, vocabulary, and staleness checks
	// run on every parsed document in the ticket; per-doc vocabulary findings
	// are aggregated into one row per (ticket, category) to keep output sane.
	prefixRe := regexp.MustCompile(`^(\d{2,3})-`)
	for _, bucket := range tickets {
		ticketPath := bucket.TicketDir
		indexPath := filepath.Join(ticketPath, "index.md")

		hasIssues := false
		emit := func(issue string, severity string, message string, path string) error {
			row := types.NewRow(
				types.MRP("ticket", bucket.TicketID),
				types.MRP("issue", issue),
				types.MRP("severity", severity),
				types.MRP("message", message),
				types.MRP("path", path),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return fmt.Errorf("failed to emit doctor row (%s) for %s: %w", issue, path, err)
			}
			switch severity {
			case "error":
				highestSeverity = maxInt(highestSeverity, 2)
			case "warning":
				highestSeverity = maxInt(highestSeverity, 1)
			}
			hasIssues = true
			return nil
		}

		// Check for unique index.md (should only be one per ticket root)
		indexFiles := findIndexFiles(ticketPath, shouldSkipPath)
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

		// Aggregates across all docs in the ticket.
		unknownVocab := newDoctorVocabAgg()
		var newestUpdate time.Time
		haveTimestamps := false

		for _, h := range bucket.Docs {
			// Only consider markdown files (index builder is md-only, but keep the guard).
			if !strings.HasSuffix(strings.ToLower(filepath.Base(h.Path)), ".md") {
				continue
			}
			dir := filepath.Dir(h.Path)
			isRootLevel := filepath.Clean(dir) == filepath.Clean(ticketPath)
			isIndex := isRootLevel && filepath.Base(h.Path) == "index.md"
			// Skip root-level control files (no frontmatter by design).
			if isRootLevel && !isIndex {
				bn := filepath.Base(h.Path)
				if bn == "README.md" || bn == "tasks.md" || bn == "changelog.md" {
					continue
				}
			}

			// invalid frontmatter (index and non-index alike).
			if h.ReadErr != nil {
				// Re-parse to preserve taxonomy details when possible.
				_, parseErr := readDocumentFrontmatter(h.Path)
				msgErr := h.ReadErr
				if parseErr != nil {
					msgErr = parseErr
				}
				if err := emit("invalid_frontmatter", "error", fmt.Sprintf("Failed to parse frontmatter: %v", msgErr), h.Path); err != nil {
					return err
				}
				renderFrontmatterParseTaxonomy(ctx, msgErr, h.Path)
				continue
			}
			if h.Doc == nil {
				continue
			}
			doc := h.Doc

			// Staleness input: track the newest update across the ticket.
			if !doc.LastUpdated.IsZero() {
				haveTimestamps = true
				if doc.LastUpdated.After(newestUpdate) {
					newestUpdate = doc.LastUpdated
				}
			}

			// Index-only structural checks (required/optional fields).
			if isIndex {
				if err := doc.Validate(); err != nil {
					if emitErr := emit("missing_required_fields", "error", err.Error(), h.Path); emitErr != nil {
						return emitErr
					}
					for _, field := range missingRequiredFields(doc) {
						detail := fmt.Sprintf("%s is required", field)
						docmgr.RenderTaxonomy(ctx, docmgrctx.NewFrontmatterSchema(h.Path, field, detail, core.SeverityError))
					}
				}
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
				for _, issue := range optionalIssues {
					if err := emit("missing_field", "warning", issue.detail, h.Path); err != nil {
						return err
					}
					docmgr.RenderTaxonomy(ctx, docmgrctx.NewFrontmatterSchema(h.Path, issue.field, issue.detail, core.SeverityWarning))
				}
			}

			// Vocabulary checks (all docs), aggregated per ticket+category.
			if dv.configured {
				for _, t := range doc.Topics {
					if _, ok := dv.topicSet[t]; !ok && t != "" {
						unknownVocab.add("topics", t)
						docmgr.RenderTaxonomy(ctx, docmgrctx.NewVocabularyUnknown(h.Path, "Topics", t, dv.topicList))
					}
				}
				if doc.DocType != "" {
					if _, ok := dv.docTypeSet[doc.DocType]; !ok {
						unknownVocab.add("docTypes", doc.DocType)
						docmgr.RenderTaxonomy(ctx, docmgrctx.NewVocabularyUnknown(h.Path, "DocType", doc.DocType, dv.docTypeList))
					}
				}
				if doc.Intent != "" {
					if _, ok := dv.intentSet[doc.Intent]; !ok {
						unknownVocab.add("intent", doc.Intent)
						docmgr.RenderTaxonomy(ctx, docmgrctx.NewVocabularyUnknown(h.Path, "Intent", doc.Intent, dv.intentList))
					}
				}
				if doc.Status != "" {
					if _, ok := dv.statusSet[doc.Status]; !ok {
						unknownVocab.add("status", doc.Status)
						docmgr.RenderTaxonomy(ctx, docmgrctx.NewVocabularyUnknown(h.Path, "Status", doc.Status, dv.statusList))
					}
				}
			}

			// RelatedFiles checks (all docs) using a doc-anchored resolver (Spec §7.3).
			resolver := paths.NewResolver(paths.ResolverOptions{
				DocsRoot:      ws.Context().Root,
				DocPath:       h.Path,
				ConfigDir:     ws.Context().ConfigDir,
				RepoRoot:      ws.Context().RepoRoot,
				WorkspaceRoot: ws.Context().WorkspaceRoot,
			})
			var missingNotes []string
			for _, rf := range doc.RelatedFiles {
				if strings.TrimSpace(rf.Path) == "" {
					continue
				}
				if strings.TrimSpace(rf.Note) == "" {
					missingNotes = append(missingNotes, rf.Path)
				}
				n := resolver.Resolve(rf.Path)
				if !n.Exists {
					if err := emit("missing_related_file", "warning", fmt.Sprintf("related file not found: %s", rf.Path), h.Path); err != nil {
						return err
					}
					docmgr.RenderTaxonomy(ctx, docmgrctx.NewRelatedFileMissing(h.Path, rf.Path, rf.Note))
				}
			}
			if len(missingNotes) > 0 {
				if err := emit("missing_related_file_note", "warning", fmt.Sprintf("%d related file(s) have no Note: %s", len(missingNotes), summarizeList(missingNotes, 5)), h.Path); err != nil {
					return err
				}
			}

			// Numeric prefix policy (subdirectory files only).
			if !isRootLevel {
				bn := filepath.Base(h.Path)
				if !prefixRe.MatchString(bn) {
					if err := emit("missing_numeric_prefix", "warning", "file without numeric prefix", h.Path); err != nil {
						return err
					}
				}
			}
		}

		// Aggregated unknown-vocabulary findings (one row per category).
		for _, cat := range unknownVocab.categories() {
			issue := map[string]string{
				"topics":   "unknown_topics",
				"docTypes": "unknown_doc_type",
				"intent":   "unknown_intent",
				"status":   "unknown_status",
			}[cat]
			msg := fmt.Sprintf("unknown %s value(s): %s; %s", cat, unknownVocab.describe(cat), vocabRemediation(cat))
			if err := emit(issue, "warning", msg, ticketPath); err != nil {
				return err
			}
		}

		// Staleness is a per-ticket concept: stale only when NO doc in the
		// ticket was updated within the window.
		if haveTimestamps {
			daysSinceUpdate := time.Since(newestUpdate).Hours() / 24
			if daysSinceUpdate > float64(settings.StaleAfterDays) {
				row := types.NewRow(
					types.MRP("ticket", bucket.TicketID),
					types.MRP("issue", "stale"),
					types.MRP("severity", "warning"),
					types.MRP("message", fmt.Sprintf("no document updated in %.0f days (threshold: %d days)", daysSinceUpdate, settings.StaleAfterDays)),
					types.MRP("path", ticketPath),
					types.MRP("last_updated", newestUpdate.Format("2006-01-02")),
				)
				if err := gp.AddRow(ctx, row); err != nil {
					return fmt.Errorf("failed to emit doctor row (stale) for %s: %w", bucket.TicketID, err)
				}
				highestSeverity = maxInt(highestSeverity, 1)
				hasIssues = true
				docmgr.RenderTaxonomy(ctx, docmgrctx.NewWorkspaceStale(indexPath, newestUpdate, settings.StaleAfterDays))
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

// doctorVocabAgg aggregates unknown-vocabulary findings per category so that
// a ticket with many docs sharing the same unknown value produces one row.
type doctorVocabAgg struct {
	order      []string
	counts     map[string]map[string]int
	valueOrder map[string][]string
}

func newDoctorVocabAgg() *doctorVocabAgg {
	return &doctorVocabAgg{
		counts:     map[string]map[string]int{},
		valueOrder: map[string][]string{},
	}
}

func (a *doctorVocabAgg) add(category string, value string) {
	if _, ok := a.counts[category]; !ok {
		a.counts[category] = map[string]int{}
		a.order = append(a.order, category)
	}
	if _, ok := a.counts[category][value]; !ok {
		a.valueOrder[category] = append(a.valueOrder[category], value)
	}
	a.counts[category][value]++
}

func (a *doctorVocabAgg) categories() []string {
	return a.order
}

// describe renders "foo (3 docs), bar (1 doc)" for a category.
func (a *doctorVocabAgg) describe(category string) string {
	parts := make([]string, 0, len(a.valueOrder[category]))
	for _, v := range a.valueOrder[category] {
		n := a.counts[category][v]
		unit := "docs"
		if n == 1 {
			unit = "doc"
		}
		parts = append(parts, fmt.Sprintf("%s (%d %s)", v, n, unit))
	}
	return strings.Join(parts, ", ")
}

// summarizeList joins up to limit items, eliding the rest with a count.
func summarizeList(items []string, limit int) string {
	if len(items) <= limit {
		return strings.Join(items, ", ")
	}
	return fmt.Sprintf("%s, … (%d more)", strings.Join(items[:limit], ", "), len(items)-limit)
}

// autoFixDocFrontmatter applies the same safe frontmatter fixes as
// 'validate frontmatter --auto-fix' (quoting unsafe scalars, normalizing
// delimiters, ...), writing a .bak backup first. It returns whether a fix was
// written, the list of applied fixes, and an error when the fix was applied
// but the document still fails to parse.
func autoFixDocFrontmatter(path string) (bool, []string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return false, nil, err
	}
	fixes, fixedContent, err := generateFixes(raw)
	if err != nil || fixedContent == nil {
		// No safe fix available; leave the finding to validation.
		return false, nil, nil
	}
	if err := applyAutoFix(path, raw, fixedContent); err != nil {
		return false, nil, err
	}
	if _, _, parseErr := documents.ReadDocumentWithFrontmatter(path); parseErr != nil {
		return true, fixes, fmt.Errorf("re-parse after auto-fix failed: %w", parseErr)
	}
	return true, fixes, nil
}

// migrateDocAnchors rewrites legacy (bare-string) RelatedFiles entries of one
// document into explicit anchored form (repo://, ws://, docs://, abs://) using
// the tightest-containing-anchor rule. Entries are only migrated when the
// legacy resolution finds an existing file; ambiguous/missing entries are left
// as legacy and reported in skipped. Already-anchored entries are untouched.
func migrateDocAnchors(ws *workspace.Workspace, docPath string) (int, []string, error) {
	doc, content, err := documents.ReadDocumentWithFrontmatter(docPath)
	if err != nil {
		return 0, nil, err
	}

	resolver := paths.NewResolver(paths.ResolverOptions{
		DocsRoot:      ws.Context().Root,
		DocPath:       docPath,
		ConfigDir:     ws.Context().ConfigDir,
		RepoRoot:      ws.Context().RepoRoot,
		WorkspaceRoot: ws.Context().WorkspaceRoot,
	})

	changed := 0
	var skipped []string
	for i, rf := range doc.RelatedFiles {
		raw := strings.TrimSpace(rf.Path)
		if raw == "" {
			continue
		}
		if paths.IsAnchored(raw) {
			continue
		}
		n := resolver.Resolve(raw)
		abs := ""
		if n.Exists {
			abs = strings.TrimSpace(n.Abs)
		}
		if abs == "" && !filepath.IsAbs(raw) {
			// Rescue pass for historical doc-relative ../ chains that escape
			// the repo: the legacy resolver rejects them (containment guard),
			// but for migration a plain doc-relative join is authoritative.
			joined := filepath.Clean(filepath.Join(filepath.Dir(docPath), filepath.FromSlash(raw)))
			if info, statErr := os.Stat(joined); statErr == nil && !info.IsDir() {
				abs = joined
			}
		}
		if abs == "" {
			skipped = append(skipped, raw)
			continue
		}
		anchored := resolver.AnchoredFor(abs).String()
		if anchored == "" || anchored == raw {
			continue
		}
		doc.RelatedFiles[i].Path = anchored
		changed++
	}

	if changed == 0 {
		return 0, skipped, nil
	}
	if err := documents.WriteDocumentWithFrontmatter(docPath, doc, content, true); err != nil {
		return 0, skipped, err
	}
	return changed, skipped, nil
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

func shouldSkipDoctorCLIPath(docsRoot string, absPath string, ignoreDirNames []string, ignoreGlobs []string) bool {
	absPath = filepath.Clean(strings.TrimSpace(absPath))
	if absPath == "" {
		return true
	}
	base := filepath.Base(absPath)
	if containsString(ignoreDirNames, base) {
		return true
	}
	if matchesDoctorIgnoreGlob(ignoreGlobs, base) || matchesDoctorIgnoreGlob(ignoreGlobs, absPath) {
		return true
	}
	docsRoot = filepath.Clean(strings.TrimSpace(docsRoot))
	if docsRoot != "" {
		if rel, err := filepath.Rel(docsRoot, absPath); err == nil {
			rel = filepath.Clean(rel)
			if rel != "." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != ".." {
				if matchesDoctorIgnoreGlob(ignoreGlobs, rel) {
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

// findIndexFiles recursively searches for all non-ignored index.md files in a directory tree.
func findIndexFiles(rootPath string, shouldSkipPath func(path, baseName string, isDir bool) bool) []string {
	var indexFiles []string

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors, continue walking
		}
		base := filepath.Base(path)
		if info.IsDir() {
			if shouldSkipPath != nil && shouldSkipPath(path, base, true) {
				return filepath.SkipDir
			}
			return nil
		}
		if shouldSkipPath != nil && shouldSkipPath(path, base, false) {
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

func matchesDoctorIgnoreGlob(patterns []string, path string) bool {
	path = filepath.Clean(strings.TrimSpace(path))
	if path == "" || path == "." {
		return false
	}
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if ok, _ := filepath.Match(p, path); ok {
			return true
		}
	}
	return false
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
	dv *doctorVocab,
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

	// Vocabulary checks. When no vocabulary file exists, emit one info-level
	// finding instead of per-value unknown_* warnings.
	if !dv.configured {
		row := types.NewRow(
			types.MRP("ticket", doc.Ticket),
			types.MRP("issue", "no_vocabulary"),
			types.MRP("severity", "info"),
			types.MRP("message", doctorNoVocabularyMessage),
			types.MRP("path", ""),
		)
		_ = gp.AddRow(ctx, row)
	} else {
		if len(doc.Topics) > 0 {
			var unknownTopics []string
			for _, t := range doc.Topics {
				if _, ok := dv.topicSet[t]; !ok && t != "" {
					unknownTopics = append(unknownTopics, t)
				}
			}
			if len(unknownTopics) > 0 {
				row := types.NewRow(
					types.MRP("ticket", doc.Ticket),
					types.MRP("issue", "unknown_topics"),
					types.MRP("severity", "warning"),
					types.MRP("message", fmt.Sprintf("unknown topics: %v; %s", unknownTopics, vocabRemediation("topics"))),
					types.MRP("path", docPath),
				)
				_ = gp.AddRow(ctx, row)
				highestSeverity = maxInt(highestSeverity, 1)
				docmgr.RenderTaxonomy(ctx, docmgrctx.NewVocabularyUnknown(docPath, "Topics", strings.Join(unknownTopics, ","), dv.topicList))
			}
		}
		if doc.DocType != "" {
			if _, ok := dv.docTypeSet[doc.DocType]; !ok {
				row := types.NewRow(
					types.MRP("ticket", doc.Ticket),
					types.MRP("issue", "unknown_doc_type"),
					types.MRP("severity", "warning"),
					types.MRP("message", fmt.Sprintf("unknown docType: %s; %s", doc.DocType, vocabRemediation("docTypes"))),
					types.MRP("path", docPath),
				)
				_ = gp.AddRow(ctx, row)
				highestSeverity = maxInt(highestSeverity, 1)
				docmgr.RenderTaxonomy(ctx, docmgrctx.NewVocabularyUnknown(docPath, "DocType", doc.DocType, dv.docTypeList))
			}
		}
		if doc.Intent != "" {
			if _, ok := dv.intentSet[doc.Intent]; !ok {
				row := types.NewRow(
					types.MRP("ticket", doc.Ticket),
					types.MRP("issue", "unknown_intent"),
					types.MRP("severity", "warning"),
					types.MRP("message", fmt.Sprintf("unknown intent: %s; %s", doc.Intent, vocabRemediation("intent"))),
					types.MRP("path", docPath),
				)
				_ = gp.AddRow(ctx, row)
				highestSeverity = maxInt(highestSeverity, 1)
				docmgr.RenderTaxonomy(ctx, docmgrctx.NewVocabularyUnknown(docPath, "Intent", doc.Intent, dv.intentList))
			}
		}
		if doc.Status != "" {
			if _, ok := dv.statusSet[doc.Status]; !ok {
				statusValidText := strings.Join(dv.statusList, ", ")
				row := types.NewRow(
					types.MRP("ticket", doc.Ticket),
					types.MRP("issue", "unknown_status"),
					types.MRP("severity", "warning"),
					types.MRP("message", fmt.Sprintf("unknown status: %s (valid values: %s; list via 'docmgr vocab list --category status'); %s", doc.Status, statusValidText, vocabRemediation("status"))),
					types.MRP("path", docPath),
				)
				_ = gp.AddRow(ctx, row)
				highestSeverity = maxInt(highestSeverity, 1)
				docmgr.RenderTaxonomy(ctx, docmgrctx.NewVocabularyUnknown(docPath, "Status", doc.Status, dv.statusList))
			}
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
	parsedValues *values.Values,
) error {
	settings := &DoctorSettings{}
	if err := parsedValues.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
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
	if err := c.RunIntoGlazeProcessor(ctx, parsedValues, collector); err != nil {
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

	// Summary-first output: multi-ticket runs default to a one-line rollup per
	// ticket. Single-ticket (--ticket) and single-doc (--doc) runs keep the
	// full per-issue detail, as does --details.
	detailMode := settings.Details ||
		strings.TrimSpace(settings.Doc) != "" ||
		(strings.TrimSpace(settings.Ticket) != "" && !settings.All)

	if detailMode {
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
	} else {
		printDoctorRollup(order, grouped)
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

// printDoctorRollup prints the multi-ticket summary: one line per ticket with
// counts and top issue kinds, plus totals. Full detail stays behind --details.
func printDoctorRollup(order []string, grouped map[string][]types.Row) {
	type kindCount struct {
		kind  string
		count int
	}

	totalErrors := 0
	totalWarnings := 0
	ticketsWithErrors := 0
	ticketsWithWarnings := 0
	ticketsOK := 0
	ticketsChecked := 0

	// Global info-level findings (e.g. no_vocabulary) print once, first.
	for _, ticket := range order {
		for _, row := range grouped[ticket] {
			if strings.EqualFold(getRowString(row, "severity"), "info") {
				fmt.Printf("[INFO] %s\n", getRowString(row, "message"))
			}
		}
	}

	var lines []string
	for _, ticket := range order {
		entries := grouped[ticket]
		errors := 0
		warnings := 0
		hasNonInfo := false
		kinds := map[string]int{}
		kindOrder := []string{}
		for _, row := range entries {
			severity := strings.ToLower(getRowString(row, "severity"))
			issue := getRowString(row, "issue")
			switch severity {
			case "error":
				errors++
			case "warning":
				warnings++
			case "ok":
				hasNonInfo = true
				continue
			default:
				continue
			}
			hasNonInfo = true
			if _, ok := kinds[issue]; !ok {
				kindOrder = append(kindOrder, issue)
			}
			kinds[issue]++
		}
		// Groups that only carry global info findings (e.g. no_vocabulary)
		// are not tickets; they were already printed above.
		if !hasNonInfo {
			continue
		}
		ticketsChecked++
		if errors > 0 {
			ticketsWithErrors++
		} else if warnings > 0 {
			ticketsWithWarnings++
		} else {
			ticketsOK++
			continue
		}
		totalErrors += errors
		totalWarnings += warnings

		counts := make([]kindCount, 0, len(kindOrder))
		for _, k := range kindOrder {
			counts = append(counts, kindCount{kind: k, count: kinds[k]})
		}
		sort.SliceStable(counts, func(i, j int) bool { return counts[i].count > counts[j].count })
		top := make([]string, 0, 3)
		for i, kc := range counts {
			if i >= 3 {
				break
			}
			top = append(top, fmt.Sprintf("%s x%d", kc.kind, kc.count))
		}
		lines = append(lines, fmt.Sprintf("%s: %d error(s), %d warning(s) — top: %s", ticket, errors, warnings, strings.Join(top, ", ")))
	}

	for _, l := range lines {
		fmt.Println(l)
	}
	fmt.Printf("%d ticket(s) checked: %d with errors, %d with warnings, %d ok (%d errors, %d warnings total)\n",
		ticketsChecked, ticketsWithErrors, ticketsWithWarnings, ticketsOK, totalErrors, totalWarnings)
	if totalErrors > 0 || totalWarnings > 0 {
		fmt.Println("Run 'docmgr doctor --ticket <ID>' or add --details for full findings.")
	}
}

func getRowString(row types.Row, field string) string {
	if val, ok := row.Get(field); ok {
		return fmt.Sprint(val)
	}
	return ""
}

var _ cmds.BareCommand = &DoctorCommand{}
