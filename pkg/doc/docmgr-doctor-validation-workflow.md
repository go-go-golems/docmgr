---
Title: Doctor and validation workflow (algorithmic walkthrough)
Slug: doctor-validation-workflow
Short: Deep dive into how doctor and validate frontmatter detect, classify, and report issues (with links to code and rules).
Topics:
- docmgr
- diagnostics
- validation
- yaml
IsTemplate: false
IsTopLevel: false
ShowPerDefault: true
SectionType: GeneralTopic
---

# Doctor and validation workflow (algorithmic walkthrough)

This page walks through the end-to-end validation pipeline used by `docmgr doctor` and `docmgr validate frontmatter`. It explains how these commands discover documents, parse frontmatter, validate schema, generate fix suggestions, and render diagnostics through the rule system. Each step is tied to specific code locations and algorithms so you can reason about behavior, extend the system, or debug issues.

For user-facing guidance on using these commands, see:
```
docmgr help yaml-frontmatter-validation
```

For implementation details on parsing and fix heuristics, see:
```
docmgr help yaml-frontmatter-validation-reference
```

For the diagnostics taxonomy system architecture, see:
```
docmgr help diagnostics-taxonomy-and-rules
```

## 1. High-Level Flow

The validation pipeline follows a consistent pattern across both `doctor` and `validate frontmatter`: discover inputs → parse frontmatter → validate schema → generate fixes (optional) → render diagnostics. The key difference is scope: `doctor` scans entire workspaces and checks multiple document types, while `validate frontmatter` focuses on a single file with optional auto-fix capabilities.

### 1.1. Input Discovery

Input discovery determines which files to validate. The strategy differs by command:

**Doctor workspace scan (`--all` or `--ticket`):**
- Discovers the workspace (docs root + config + repo root) via `workspace.DiscoverWorkspace` (see `internal/workspace/workspace.go`)
- Builds a temporary in-memory workspace index via `Workspace.InitIndex` (see `internal/workspace/index_builder.go`)
  - Ingestion applies the canonical skip policy (for example `.meta/` and `_*/` like `_templates/` / `_guidelines/`) — see `internal/workspace/skip_policy.go`
- Queries the indexed doc set via `Workspace.QueryDocs` (see `internal/workspace/query_docs.go`)
  - `doctor` requests `IncludeErrors=true` so parse-error docs are available as `DocHandle{ReadErr: ...}` for repair workflows
  - `doctor` requests `IncludeDiagnostics=true` so QueryDocs can emit structured diagnostics (for example parse-skip and normalization fallback)
- Detects “ticket scaffolds missing index.md” separately via `workspace.FindTicketScaffoldsMissingIndex` (see `internal/workspace/discovery.go`)
- Respects `.docmgrignore` and `--ignore-glob` as a compatibility layer:
  - the index uses the canonical ingestion policy, and
  - `doctor` applies ignore globs/dirs as a **post-filter** over QueryDocs results so behavior matches legacy expectations (see `pkg/commands/doctor.go`)

**Doctor single-file mode (`--doc`):**
- Validates exactly one file specified by `--doc` path
- Path can be absolute or relative to `--root` (defaults to `ttmp/`)
- Implemented in `validateSingleDoc` function (`pkg/commands/doctor.go`)

**Validate frontmatter (`--doc`):**
- Validates exactly one file (required `--doc` parameter)
- Same path resolution as doctor single-file mode
- Focuses solely on frontmatter syntax and structure

### 1.2. Parsing Phase

All commands use the same frontmatter parser: `ReadDocumentWithFrontmatter` (`internal/documents/frontmatter.go`). This function extracts the YAML block between `---` delimiters, preprocesses risky scalars to quote colons and special characters, decodes with `yaml.Node` to preserve position information, and wraps any failures in a `FrontmatterParseTaxonomy` with line numbers, column positions, code snippets, and user-friendly problem descriptions.

**Key parsing steps:**
1. Extract frontmatter block (find `---` delimiters)
2. Preprocess YAML (quote unsafe scalars via `frontmatter.PreprocessYAML`)
3. Decode with `yaml.Decoder` into `yaml.Node` (preserves line/col)
4. On error: extract line/col, build snippet, classify error, emit taxonomy
5. Decode node into `models.Document` struct

### 1.3. Schema Validation Phase

After successful parsing, schema validation checks that required fields are present and optionally warns about missing recommended fields or vocabulary mismatches.

**Required fields (all commands):**
- `models.Document.Validate()` checks Title, Ticket, DocType
- Failures emit `FrontmatterSchema` taxonomies with field names and details
- Both `doctor` and `validate frontmatter` use this validation

**Optional checks (doctor only):**
- Missing Status or Topics fields (warnings, not errors)
- Vocabulary mismatches: Topics, DocType, Intent, Status values not in `vocabulary.yaml`
- Implemented in `pkg/commands/doctor.go` after successful parse

### 1.4. Fix Generation Phase (validate frontmatter only)

The `validate frontmatter` command can generate fix suggestions and optionally apply them automatically. Fix generation operates on raw file bytes (not parsed structure), allowing repairs even when parsing fails completely.

**Fix heuristics (`generateFixes` in `pkg/commands/validate_frontmatter.go`):**
1. Normalize delimiters (add missing closing `---`, handle stray delimiter lines)
2. Scrub stray delimiter lines inside frontmatter content
3. Peel trailing non-key lines (move plain text without `:` from frontmatter to body)
4. Quote unsafe scalars via `PreprocessYAML`

**Fix application (`--auto-fix`):**
1. Generate fixes and fixed content
2. Write backup file (`<path>.bak`)
3. Rewrite file with fixed content
4. Re-parse the file
5. On success: suppress error taxonomy, print "Frontmatter auto-fixed" message
6. On failure: emit new taxonomy from failed re-parse

### 1.5. Reporting Phase

All validation errors flow through docmgr's diagnostics taxonomy system. Taxonomies are rendered via the default rule registry (`pkg/diagnostics/docmgrrules`), which matches taxonomies to rules based on stage and symptom codes. Rules produce `RuleResult` objects with headlines, bodies, severities, and suggested actions.

**Frontmatter rule output:**
- Headline: "YAML/frontmatter syntax error"
- Body: File path, line/col, problem description, code snippet, suggested fixes (if present)
- Actions: Link to `docmgr validate frontmatter --doc <path>`

## 2. Detailed Algorithms

### 2.1. Frontmatter Parsing

**File:** `internal/documents/frontmatter.go` (`ReadDocumentWithFrontmatter`)

**Algorithm:**
```
raw = read file bytes
fm, body, fmStartLine = extractFrontmatter(raw)
  // Scans for --- delimiters
  // Returns frontmatter bytes, body bytes, starting line number (1-based)

fm = frontmatter.PreprocessYAML(fm)
  // Walks top-level key-value lines
  // Quotes risky scalars (colons, special chars, etc.)

decoder = yaml.NewDecoder(bytes.NewReader(fm))
node = yaml.Node{}
if err := decoder.Decode(&node); err != nil {
  // Extract line number from error message (regex: "line ([0-9]+)")
  yamlLine = extractLineNumber(err)
  absoluteLine = fmStartLine + yamlLine - 1
  
  // Build snippet: 3 lines of context with line numbers
  snippet = buildSnippet(fileLines, absoluteLine, column)
  
  // Classify error into user-friendly problem text
  problem = classifyYAMLError(err)
  
  // Wrap in taxonomy
  taxonomy = docmgrctx.NewFrontmatterParse(path, absoluteLine, column, snippet, problem, err)
  return nil, "", core.WrapWithCause(err, taxonomy)
}

// Decode node into Document struct
doc = models.Document{}
if err := node.Decode(&doc); err != nil {
  // Same error handling as above (extract line/col, build snippet, emit taxonomy)
}

return &doc, body, nil
```

**Outputs:**
- Success: `*models.Document`, body string, `nil` error
- Failure: `nil`, empty string, taxonomy-wrapped error with line/col/snippet/problem

**Error extraction details:**
- YAML parser reports line numbers relative to frontmatter start (line 1 = first line after opening `---`)
- Absolute file line = `fmStartLine + yamlLine - 1`
- Column numbers are usually 0 (not reliably extracted from YAML errors)
- Snippet shows 3 lines: `max(1, line-1)` to `min(len(lines), line+1)` with optional caret

### 2.2. Schema Validation

**Required fields:** `models.Document.Validate()` (`pkg/models/document.go`)

**Algorithm:**
```
missing := []string{}
if doc.Title == "" {
  missing = append(missing, "Title")
}
if doc.Ticket == "" {
  missing = append(missing, "Ticket")
}
if doc.DocType == "" {
  missing = append(missing, "DocType")
}
if len(missing) > 0 {
  return fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
}
return nil
```

**Taxonomy emission:**
- On validation failure, `doctor` and `validate frontmatter` emit `FrontmatterSchema` taxonomies
- Context includes file path, field name, and detail message
- Rule renderer (`FrontmatterSchemaRule`) suggests `docmgr meta update --field <field> --value <value>`

**Optional doctor checks:**
- Missing Status/Topics: Emits warnings (not errors) via `docmgr.RenderTaxonomy`
- Vocabulary mismatches: Checks Topics, DocType, Intent, Status against `vocabulary.yaml`
- Unknown values emit `VocabularyUnknown` taxonomies with known values list and `vocab add` suggestion

### 2.3. Fix Generation (validate frontmatter)

**File:** `pkg/commands/validate_frontmatter.go` (`generateFixes`)

**Algorithm:**
```
// Try normal frontmatter extraction
fm, body, _, err := documents.SplitFrontmatter(raw)
if err != nil {
  // Fallback: normalize delimiters heuristically
  fixed, fixErr := normalizeDelimiters(raw)
  if fixErr == nil {
    fixes = []string{"Normalize frontmatter delimiters (add missing closing ---)"}
    return fixes, fixed, nil
  }
  return nil, nil, err
}

// Clean up frontmatter content
cleaned := scrubStrayDelimiters(fm)
  // Remove lines that look like --- inside frontmatter

fmLines := bytes.Split(cleaned, []byte("\n"))
peeledBody := peelTrailingNonKeyLines(&fmLines)
  // Move trailing lines without ":" from frontmatter to body

cleaned = bytes.Join(fmLines, []byte("\n"))

// Quote unsafe scalars
quoted := frontmatter.PreprocessYAML(cleaned)
  // Quotes colons, special chars, etc.

// Build fix descriptions
fixes := []string{}
if !bytes.Equal(quoted, cleaned) {
  fixes = append(fixes, "Quote unsafe scalars (colons, hashes, special leading chars)")
}
if !bytes.Equal(cleaned, fm) {
  fixes = append(fixes, "Remove stray delimiter lines inside frontmatter")
}
if len(peeledBody) > 0 {
  fixes = append(fixes, "Move non key/value lines out of frontmatter")
}

// Reconstruct file
buf := bytes.Buffer{}
buf.WriteString("---\n")
buf.Write(quoted)
buf.WriteString("\n---\n")
if len(peeledBody) > 0 {
  buf.Write(peeledBody)
  buf.WriteByte('\n')
}
buf.Write(body)

return fixes, buf.Bytes(), nil
```

**Fix application (`applyAutoFix`):**
```
// Write backup
backupPath := path + ".bak"
os.WriteFile(backupPath, original, 0644)

// Write fixed content
os.WriteFile(path, fixed, 0644)

// Re-parse to verify
doc, _, err := documents.ReadDocumentWithFrontmatter(path)
if err == nil {
  // Success: suppress original error taxonomy, print success message
  fmt.Printf("Frontmatter auto-fixed: %s\n", path)
  return doc, nil
}

// Failure: emit new taxonomy from failed re-parse
if tax, ok := core.AsTaxonomy(err); ok {
  docmgr.RenderTaxonomy(ctx, tax)
}
return nil, fmt.Errorf("auto-fix applied but re-parse failed: %w", err)
```

### 2.4. Doctor Single-File Path

**File:** `pkg/commands/doctor.go` (`validateSingleDoc`)

**Algorithm:**
```
// Parse frontmatter
doc, _, err := documents.ReadDocumentWithFrontmatter(path)
if err != nil {
  // Render taxonomy if it's a taxonomy-wrapped error
  if tax, ok := core.AsTaxonomy(err); ok {
    docmgr.RenderTaxonomy(ctx, tax)
  }
  return severityError, err
}

// Validate required fields
if err := doc.Validate(); err != nil {
  // Emit FrontmatterSchema taxonomy for each missing field
  // Render via rule system
  return severityError, err
}

// Optional checks (warnings)
if doc.Status == "" {
  docmgr.RenderTaxonomy(ctx, docmgrctx.NewFrontmatterSchema(path, "Status", "missing recommended field"))
}
if len(doc.Topics) == 0 {
  docmgr.RenderTaxonomy(ctx, docmgrctx.NewFrontmatterSchema(path, "Topics", "missing recommended field"))
}

// Vocabulary checks
checkVocabularyMismatches(ctx, doc, topicSet, docTypeSet, intentSet, statusSet)

// Success: emit success row
row := types.NewRow(
  types.MRP("doc", path),
  types.MRP("status", "ok"),
)
gp.AddRow(ctx, row)
return severityOK, nil
```

**Note:** Related file checks (missing files in `RelatedFiles`) are currently only performed in workspace scan mode (for ticket index files), not in single-file mode.

## 3. Command Behaviors

### 3.1. Doctor Workspace Scan

**Command:** `docmgr doctor --all` or `docmgr doctor --ticket <TICKET>`

**Behavior:**
1. Discovers the workspace (`workspace.DiscoverWorkspace`) and resolves the effective docs root/config/repo anchors.
2. Builds the in-memory workspace index once (`Workspace.InitIndex`), applying the canonical skip policy during ingestion.
3. Checks for ticket scaffolds missing `index.md` (`workspace.FindTicketScaffoldsMissingIndex`) and emits `missing_index` findings (scoped to `--ticket` when provided).
4. Queries the indexed doc set via `Workspace.QueryDocs` (typically with `IncludeErrors=true` and `IncludeDiagnostics=true`) and groups findings by ticket.
5. Applies validation and checks:
   - frontmatter parse/schema issues (as taxonomies)
   - optional field and vocabulary warnings
   - `RelatedFiles` existence checks (doc-anchored path normalization)
   - stale docs (`LastUpdated` older than `--stale-after` days)
6. Applies `.docmgrignore`/`--ignore-glob`/`--ignore-dir` as a post-filter over QueryDocs results (compatibility behavior).
7. Optionally writes diagnostics JSON (`--diagnostics-json`)
8. Exits with error code if `--fail-on` threshold is met

**Output:**
- Human-readable: Grouped by ticket, shows issue type, severity, message, path
- JSON (with `--with-glaze-output --output json`): Array of row objects
- Diagnostics JSON (`--diagnostics-json`): Array of `RuleResult` objects

### 3.2. Doctor Single-File Mode

**Command:** `docmgr doctor --doc <file>`

**Behavior:**
1. Validates exactly one file
2. Parses frontmatter (renders taxonomy on failure)
3. Validates required fields
4. Checks optional fields (Status, Topics) as warnings
5. Validates vocabulary
6. Emits success row if no issues

**Limitations:**
- No auto-fix (use `validate frontmatter --auto-fix` instead)
- No related file checks (only in workspace scan)
- No stale doc checks (only in workspace scan)

### 3.3. Validate Frontmatter

**Command:** `docmgr validate frontmatter --doc <file> [--suggest-fixes] [--auto-fix]`

**Behavior:**
1. Validates exactly one file
2. Parses frontmatter (renders taxonomy on failure)
3. If `--suggest-fixes` or `--auto-fix`:
   - Generates fix suggestions
   - Attaches suggestions to `FrontmatterParseContext.Fixes`
   - Rule renderer prints suggestions
4. If `--auto-fix`:
   - Applies fixes (creates `.bak` backup)
   - Re-parses to verify
   - Suppresses error taxonomy on success
   - Emits new taxonomy on failure
5. Emits success row if validation passes

**Output:**
- Human-readable: Error taxonomy with line/col/snippet/suggestions, or success message
- JSON (with `--with-glaze-output --output json`): Row object with doc path, title, ticket, docType, status

## 4. Diagnostics Integration

Frontmatter validation errors flow through docmgr's diagnostics taxonomy system. This design allows consistent error reporting across all commands (`doctor`, `list docs`, `search`, `meta update`, etc.) and enables fix suggestions to be attached to taxonomies and rendered by rules.

### 4.1. Taxonomy Flow

**Parse errors:**
1. `ReadDocumentWithFrontmatter` wraps failures in `FrontmatterParseTaxonomy`
2. Taxonomy carries: file, line, col, snippet, problem, original error
3. Commands call `docmgr.RenderTaxonomy(ctx, taxonomy)` to render

**Schema errors:**
1. `doc.Validate()` returns error listing missing fields
2. Commands emit `FrontmatterSchema` taxonomies per missing field
3. Taxonomy carries: file, field name, detail message

**Fix suggestions:**
1. `validate frontmatter --suggest-fixes` generates fixes
2. Attaches `Fixes []string` to `FrontmatterParseContext`
3. `FrontmatterSyntaxRule` renders fixes as numbered list

### 4.2. Rule Rendering

**FrontmatterSyntaxRule** (`pkg/diagnostics/docmgrrules/frontmatter_rules.go`):
- Matches: `StageFrontmatterParse` + `SymptomYAMLSyntax`
- Renders: File path, line/col, problem, snippet, suggested fixes (if present)
- Actions: Link to `docmgr validate frontmatter --doc <path>`

**FrontmatterSchemaRule**:
- Matches: `StageFrontmatterParse` + `SymptomSchemaViolation`
- Renders: File path, field name, issue detail
- Actions: Link to `docmgr meta update --doc <path> --field <field> --value <value>`

### 4.3. Integration Points

**Commands that emit frontmatter taxonomies:**
- `doctor`: Workspace scan and single-file mode
- `validate frontmatter`: Single-file validation
- `list docs` / `search`: Skip docs with parse errors, emit `ListingSkip` taxonomies
- `meta update` / `relate` / `rename_ticket`: Wrap parse errors in taxonomies

**Helper functions:**
- `internal/documents/frontmatter.go`: `ReadDocumentWithFrontmatter` wraps parse errors
- `internal/workspace/index_builder.go`: ingestion stores `parse_ok` / `parse_err` so parse failures can be surfaced consistently via diagnostics and `IncludeErrors`
- `internal/workspace/discovery.go`: missing-index scaffold detection (`FindTicketScaffoldsMissingIndex`)

## 5. Files and Symbols Reference

**Parser:**
- `internal/documents/frontmatter.go`: `ReadDocumentWithFrontmatter`, `extractFrontmatter`, `classifyYAMLError`, `buildSnippet`, `SplitFrontmatter`, `WriteDocumentWithFrontmatter`

**Fix engine:**
- `pkg/commands/validate_frontmatter.go`: `ValidateFrontmatterCommand`, `generateFixes`, `normalizeDelimiters`, `scrubStrayDelimiters`, `peelTrailingNonKeyLines`, `applyAutoFix`, `tryAttachFixes`

**Preprocessing:**
- `pkg/frontmatter/frontmatter.go`: `NeedsQuoting`, `QuoteValue`, `PreprocessYAML`

**Diagnostics:**
- `pkg/diagnostics/docmgrctx/frontmatter.go`: `FrontmatterParseContext` (with `Fixes` field), `NewFrontmatterParse`, `NewFrontmatterSchema`
- `pkg/diagnostics/docmgrrules/frontmatter_rules.go`: `FrontmatterSyntaxRule`, `FrontmatterSchemaRule`

**Doctor:**
- `pkg/commands/doctor.go`: `DoctorCommand`, `validateSingleDoc`, workspace scan logic
- `internal/workspace/workspace.go`: `DiscoverWorkspace`, `Workspace`, `InitIndex`, `QueryDocs`
- `internal/workspace/discovery.go`: `FindTicketScaffoldsMissingIndex`

**Models:**
- `pkg/models/document.go`: `Document`, `Validate()`, `RelatedFiles`

**Documentation:**
- `pkg/doc/docmgr-diagnostics-and-rules.md`: Diagnostics taxonomy system architecture
- `pkg/doc/docmgr-yaml-frontmatter-validation.md`: User guide for validation and auto-fix
- `pkg/doc/docmgr-yaml-frontmatter-validation-reference.md`: Technical reference for parsing and fix algorithms
- Ticket reference: `ttmp/2025/11/29/DOCMGR-YAML-001.../reference/02-frontmatter-healing-and-validation-guide.md`

## 6. Extending the System

### 6.1. Adding Doctor Auto-Fix

To add auto-fix capabilities to `doctor`:

1. Reuse fix engine from `validate_frontmatter.go` (`generateFixes`, `applyAutoFix`)
2. Add `--auto-fix` flag to `DoctorSettings`
3. In workspace scan loop, when frontmatter parse fails:
   - Call `generateFixes` to get suggestions and fixed content
   - If `--auto-fix`, call `applyAutoFix` (creates `.bak`, rewrites, re-parses)
   - On success: suppress error taxonomy, emit success row
   - On failure: emit new taxonomy from failed re-parse
4. Follow same pattern as `validate frontmatter` for backup handling and error suppression

### 6.2. Adding Schema Rules

To add new schema validation rules:

1. Define validator function (e.g., check field format, value ranges)
2. Emit `FrontmatterSchema` taxonomies on violations
3. Update `FrontmatterSchemaRule` if rendering needs changes
4. Add actions pointing to appropriate fix command (usually `docmgr meta update`)

### 6.3. Adding Fix Heuristics

To add new fix heuristics:

1. Implement heuristic function in `pkg/commands/validate_frontmatter.go`
2. Call it in `generateFixes` (order matters: delimiters → cleanup → quoting)
3. Add fix description to returned `fixes` array
4. Add unit test in `validate_frontmatter_test.go`
5. Update smoke test if behavior changes significantly

### 6.4. Changing Rules

When modifying rule rendering:

1. Update rule in `pkg/diagnostics/docmgrrules/frontmatter_rules.go`
2. Ensure actions point to correct verb and help page
3. Test with real taxonomies (use smoke test scenarios)
4. Update documentation if output format changes

## 7. Testing and Validation

**Unit tests:**
- `pkg/commands/validate_frontmatter_test.go`: Fix heuristic tests (delimiter normalization, stray cleanup)
- `pkg/frontmatter/frontmatter_test.go`: Quoting helper tests (`NeedsQuoting`, `QuoteValue`, `PreprocessYAML`)
- `internal/documents/frontmatter_test.go`: Parsing and error extraction tests
- `pkg/commands/doctor_test.go`: Doctor command tests

**Smoke tests:**
- `test-scenarios/testing-doc-manager/18-validate-frontmatter-smoke.sh`: Validation verb smoke (fail → suggest → auto-fix → success, verifies `.bak` creation)
- `test-scenarios/testing-doc-manager/15-diagnostics-smoke.sh`: Diagnostics smoke (exercises frontmatter parse taxonomy via doctor/list/search, writes diagnostics JSON)

**Manual testing:**
- Run `docmgr doctor --all` on a workspace with known issues
- Verify taxonomies render correctly
- Test `validate frontmatter --auto-fix` on broken files
- Check that `.bak` files are created and fixes are applied correctly
