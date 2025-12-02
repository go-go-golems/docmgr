---
Title: Frontmatter healing, validation, and auto-fix guide
Ticket: DOCMGR-YAML-001
Status: active
Topics:
  - yaml
  - ux
  - errors
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
  - Path: pkg/commands/validate_frontmatter.go
    Note: CLI verb and fix engine
  - Path: internal/documents/frontmatter.go
    Note: Position-aware parser + preprocessing
  - Path: pkg/frontmatter/frontmatter.go
    Note: Quoting and preprocessing helpers
  - Path: pkg/diagnostics/docmgrctx/frontmatter.go
    Note: Taxonomy context with fixes
  - Path: pkg/diagnostics/docmgrrules/frontmatter_rules.go
    Note: Rule rendering with snippets and suggestions
  - Path: test-scenarios/testing-doc-manager/18-validate-frontmatter-smoke.sh
    Note: Smoke coverage for suggest/auto-fix
  - Path: pkg/commands/validate_frontmatter_test.go
    Note: Fix heuristic unit tests
ExternalSources: []
Summary: Detailed guide for interns on how docmgr heals, validates, and auto-fixes YAML frontmatter, including algorithms, file locations, and CLI usage with examples.
LastUpdated: 2025-12-02
---

# Frontmatter healing, validation, and auto-fix (intern guide)

This page explains exactly how docmgr detects, classifies, and heals YAML frontmatter issues. It maps every algorithm to its code, shows example inputs/outputs, and lists the CLI flows (including auto-fix) you should use and extend. Follow this when debugging or adding new fixes.

## Quick pointers
- Parser + diagnostics: `internal/documents/frontmatter.go` (`ReadDocumentWithFrontmatter`, `extractFrontmatter`, `buildSnippet`, `classifyYAMLError`) wraps failures as taxonomies and attaches snippets/line numbers.
- Quoting/preprocessing: `pkg/frontmatter/frontmatter.go` (`NeedsQuoting`, `QuoteValue`, `PreprocessYAML`) is used for both reads and writes.
- Validation verb: `pkg/commands/validate_frontmatter.go` implements `docmgr validate frontmatter` with `--suggest-fixes` and `--auto-fix` (writes `.bak`).
- Taxonomy context: `pkg/diagnostics/docmgrctx/frontmatter.go` adds `Fixes []string` to `FrontmatterParseContext` so rules can surface suggestions.
- Rule rendering: `pkg/diagnostics/docmgrrules/frontmatter_rules.go` prints line/col/snippet, the problem, and suggested fixes.
- Smokes/tests: `test-scenarios/testing-doc-manager/18-validate-frontmatter-smoke.sh` (fail → suggest → auto-fix → success); `pkg/commands/validate_frontmatter_test.go` (heuristics).

## Parsing and error capture
**Algorithm (internal/documents/frontmatter.go)**
- Manually extract frontmatter between `---` delimiters (`extractFrontmatter`), returning frontmatter bytes, body, and starting line offset.
- Preprocess frontmatter with `frontmatter.PreprocessYAML` to quote risky scalars before decoding.
- Decode with `yaml.Decoder` into a `yaml.Node` to preserve positions; decode to `models.Document`.
- On error: extract line/col (`extractLineCol`), classify the message (`classifyYAMLError`), build a snippet (`buildSnippet`), and wrap in `docmgrctx.NewFrontmatterParseTaxonomy`.
- Context carries `File`, `Line`, `Column`, `Snippet`, `Problem`, `Fixes` (Fixes filled later by the validation verb).

**Examples**
- Missing closing bracket:
  - Input: `Topics: [chat` → Problem: “did not find expected ',' or ']'”, Line/Col set, snippet shows offending line.
- Mapping values not allowed:
  - Input: `Summary: text: with colon` → Problem: “mapping values are not allowed…”; snippet shows line; suggested fix will be quoting.
- Missing delimiters:
  - Input missing closing `---` → Problem: “frontmatter delimiters '---' not found.”

## Quoting and preprocessing
**Module:** `pkg/frontmatter/frontmatter.go`
- `NeedsQuoting(string) bool`: checks for colon patterns, special leading chars, inline comments, tabs, template markers.
- `QuoteValue(string) string`: single-quotes and escapes internal single quotes.
- `PreprocessYAML([]byte) []byte`: walks top-level key/value lines, quotes unsafe scalars, skips nested/list structures and already quoted/complex values.

**Usage:**
- Read path: preprocessing runs before decoding to reduce parse failures.
- Write path: writer now preprocesses encoded YAML to enforce quoting.

## Validation CLI (`docmgr validate frontmatter`)
**File:** `pkg/commands/validate_frontmatter.go`
- Flags:
  - `--doc`: required path (absolute or relative to `--root`).
  - `--suggest-fixes`: attach generated fixes to taxonomy; rule prints them.
  - `--auto-fix`: generate fixes, write `.bak`, rewrite file, and re-parse; on success prints “Frontmatter auto-fixed: …” and exits 0.
- Workflow:
  1) Read file; try normal parse.
  2) On failure, generate fixes (`generateFixes`):
     - Normalize delimiters (`normalizeDelimiters`): add missing closing `---`, wrap if no start delimiter, peel frontmatter/body, handle stray delimiter lines.
     - Scrub stray delimiter lines inside frontmatter (`scrubStrayDelimiters`).
     - Peel trailing non key/value lines out of frontmatter (`peelTrailingNonKeyLines`).
     - Quote unsafe scalars via `frontmatter.PreprocessYAML`.
  3) Attach `Fixes` to `FrontmatterParseContext` so rules render suggestions.
  4) If `--auto-fix`: write `.bak`, rewrite file, re-parse; on success suppress error taxonomy and print success; on failure render new taxonomy from the failed re-parse.

**Actions and rules:**
- Rules in `pkg/diagnostics/docmgrrules/frontmatter_rules.go`:
  - Headline: “YAML/frontmatter syntax error”
  - Body: file, line/col, problem, snippet, suggested fixes list (when present)
  - Action: `docmgr validate frontmatter --doc <file>`

## Writer behavior (healing on write)
**File:** `internal/documents/frontmatter.go` (`WriteDocumentWithFrontmatter`)
- Encodes `models.Document` via `yaml.Encoder`, then preprocesses with `frontmatter.PreprocessYAML` to ensure risky scalars get quoted before writing.
- Applies everywhere via command-level writers (add, meta update, create_ticket, doc_move, rename_ticket, ticket_close, import).

## Diagnostics integration
- Frontmatter parse errors propagate as taxonomies to:
  - doctor (`pkg/commands/doctor.go`)
  - list/search (`pkg/commands/list_docs.go`, `search.go`)
  - meta/relate/rename/import via `readDocumentFrontmatter` helpers
- Validation verb uses renderer/collector so fixes appear through the rule system (no ad-hoc printing).
- Context `Fixes` makes rules render suggestions alongside the standard problem/snippet.

## Examples (CLI)
**Parse failure (no fixes):**
```
docmgr validate frontmatter --doc ttmp/.../bad.md
1) [error] YAML/frontmatter syntax error
File: .../bad.md
Problem: frontmatter delimiters '---' not found
Actions:
- Validate frontmatter: docmgr validate frontmatter --doc .../bad.md
```

**Suggest fixes:**
```
docmgr validate frontmatter --doc .../bad.md --suggest-fixes
Suggested fixes:
  1. Normalize frontmatter delimiters (add missing closing ---)
```

**Auto-fix success:**
```
docmgr validate frontmatter --doc .../bad.md --auto-fix
Frontmatter auto-fixed: .../bad.md
Frontmatter OK: .../bad.md (Ticket=MEN-4242 DocType=reference)
```

**Auto-fix failure (will re-render new taxonomy):**
```
docmgr validate frontmatter --doc .../bad.md --auto-fix
1) [error] YAML/frontmatter syntax error
... (from re-parse)
Error: auto-fix applied but re-parse failed: ...
```

## Examples (input/output snippets)
- Broken delimiter + stray lines:
  - Input:
    ```
    ---
    Title: Broken
    Ticket: MEN-4242
    DocType: reference
    Summary: needs: quoting
    ----
    Body text
    ```
  - Auto-fix output:
    ```
    ---
    Title: Broken
    Ticket: MEN-4242
    DocType: reference
    Summary: 'needs: quoting'
    ---
    Body text
    ```

- Trailing non-key line inside frontmatter:
  - Input:
    ```
    ---
    Title: T
    Note: foo
    plain trailing line
    ---
    Body
    ```
  - Auto-fix moves `plain trailing line` into body and quotes where needed.

## Validations (what we detect, per case)

### Delimiter integrity
- **What:** Missing opening/closing `---` or stray delimiter lines.
- **Where:** `internal/documents/frontmatter.go` (`extractFrontmatter`).
- **Taxonomy:** `docmgr.frontmatter.parse` / `yaml_syntax` with problem “frontmatter delimiters '---' not found”.
- **Pseudocode:**
  ```
  scan lines for first '---' -> start
  scan for next '---' after start -> end
  if start <0 or end <= start: error("frontmatter delimiters '---' not found")
  ```

### YAML syntax errors
- **What:** Decoder errors such as “mapping values are not allowed”, “did not find expected key”, “could not find expected ':'”, bracket issues.
- **Where:** `internal/documents/frontmatter.go` (`ReadDocumentWithFrontmatter`, `classifyYAMLError`, `buildSnippet`).
- **Taxonomy:** `docmgr.frontmatter.parse` / `yaml_syntax` with line/col + snippet.
- **Pseudocode:**
  ```
  preprocess frontmatter (quote unsafe)
  if yaml decode fails:
    line,col := extractLineCol(err, fmStartLine)
    snippet := buildSnippet(lines, line, col)
    problem := classifyYAMLError(err)
    emit taxonomy(line,col,snippet,problem)
  ```

### Unsafe scalars
- **What:** Colons/hashes/tabs/special leading chars in unquoted scalars.
- **Where:** `pkg/frontmatter/frontmatter.go` (`NeedsQuoting`, `PreprocessYAML`).
- **Taxonomy:** surfaces as YAML syntax if unquoted; prevention via preprocessing.
- **Pseudocode:**
  ```
  for each top-level key/value:
    if value starts with special or contains ": " or " #":
      quote(value)
  ```

### Trailing non key/value lines
- **What:** Plain text without `:` inside frontmatter causes “expected ':'”.
- **Where:** Healed in `peelTrailingNonKeyLines` during fix generation.
- **Taxonomy:** surfaces as YAML syntax if not healed.
- **Pseudocode:**
  ```
  while last line has no ":":
    move line from fmLines to bodyLines
  ```

### Schema (future)
- **What:** Field-level checks (missing/invalid values).
- **Where:** To be added via validators (emit `schema_violation` taxonomies).

## Fix heuristics (how healing is applied, per heuristic)

### Normalize delimiters
- **Where:** `pkg/commands/validate_frontmatter.go` (`normalizeDelimiters`).
- **Goal:** Add missing closing `---`, wrap when absent, treat variants, split fm/body.
- **Pseudocode:**
  ```
  find start '---' (or assume none)
  find end '---' after start; if none, set end to first blank or EOF
  fm := lines[start+1:end]; body := lines[end+1:]
  if missing start: wrap everything
  return "---\n" + fm + "\n---\n" + body
  ```

### Scrub stray delimiter lines
- **Where:** `scrubStrayDelimiters` (same file).
- **Goal:** Remove delimiter-like lines inside fm to avoid parser confusion.
- **Pseudocode:**
  ```
  for each line in fm:
    if trimmed starts with '---': skip
    else keep
  ```

### Peel trailing non-key lines
- **Where:** `peelTrailingNonKeyLines` (same file).
- **Goal:** Move trailing non `:` lines from fm into body before rewrite.
- **Pseudocode:**
  ```
  while last line in fm does not contain ":":
    peel = append(peel, last)
    remove last from fm
  body = peel + body
  ```

### Quote unsafe scalars
- **Where:** `pkg/frontmatter/frontmatter.go` (`PreprocessYAML`).
- **Goal:** Quote risky values to avoid YAML syntax errors.
- **Pseudocode:**
  ```
  for each top-level key/value:
    if NeedsQuoting(value): value = QuoteValue(value)
  ```

### Auto-fix orchestration
- **Where:** `generateFixes` + `applyAutoFix` in `pkg/commands/validate_frontmatter.go`.
- **Goal:** Chain heuristics, write `.bak`, rewrite, re-parse; on success suppress error taxonomy.
- **Pseudocode:**
  ```
  fm, body = SplitFrontmatter(raw) or normalizeDelimiters(raw)
  fm = scrubStrayDelimiters(fm)
  peel trailing non-key lines into body
  fm = PreprocessYAML(fm)
  write .bak; rewrite file with fm + body
  re-parse; if ok -> success message; else emit new taxonomy
  ```

## Tests and smokes
- Unit tests: `pkg/commands/validate_frontmatter_test.go` (delimiter normalization, stray delimiter cleanup); add more for peel/quote as needed.
- Smoke: `test-scenarios/testing-doc-manager/18-validate-frontmatter-smoke.sh` covers fail → suggest → auto-fix → success (creates `.bak`).
- Diagnostics smoke: `test-scenarios/testing-doc-manager/15-diagnostics-smoke.sh` exercises frontmatter parse taxonomy (line/col/snippet) via doctor/list/search.

## How to extend
1) New fix heuristic: add to `generateFixes` (keep reversible and safe), unit-test it, and ensure `Fixes` text is helpful.
2) Schema validation: add validators, emit `FrontmatterSchema` taxonomies via doctor/validate, and add rules/actions.
3) Docs: update `pkg/doc/docmgr-diagnostics-and-rules.md` and CLI guides when adding flags or behaviors.

## File/symbol map
- `internal/documents/frontmatter.go`: `ReadDocumentWithFrontmatter`, `extractFrontmatter`, `buildSnippet`, `classifyYAMLError`, `WriteDocumentWithFrontmatter`, `SplitFrontmatter`.
- `pkg/frontmatter/frontmatter.go`: `NeedsQuoting`, `QuoteValue`, `PreprocessYAML`.
- `pkg/commands/validate_frontmatter.go`: `ValidateFrontmatterCommand`, `generateFixes`, `normalizeDelimiters`, `scrubStrayDelimiters`, `peelTrailingNonKeyLines`, `applyAutoFix`.
- `pkg/diagnostics/docmgrctx/frontmatter.go`: `FrontmatterParseContext` (with `Fixes`), `NewFrontmatterParseTaxonomy`.
- `pkg/diagnostics/docmgrrules/frontmatter_rules.go`: `FrontmatterSyntaxRule`.
- `test-scenarios/testing-doc-manager/18-validate-frontmatter-smoke.sh`: smoke sequence.
- `pkg/commands/validate_frontmatter_test.go`: unit tests for fix heuristics.
