---
Title: YAML Frontmatter Validation and Healing — Technical Reference
Slug: yaml-frontmatter-validation-reference
Short: Technical reference for YAML frontmatter parsing, error classification, and auto-fix algorithms.
Topics:
  - yaml
  - validation
  - frontmatter
  - implementation
IsTemplate: false
IsTopLevel: false
ShowPerDefault: false
SectionType: GeneralTopic
---

# YAML Frontmatter Validation and Healing — Technical Reference

This document describes the implementation details of docmgr's YAML frontmatter validation and auto-fix system. It covers parsing algorithms, error classification, fix heuristics, and the integration points between parsing, diagnostics, and the validation CLI verb. Use this reference when debugging parsing issues, extending fix heuristics, or understanding tradeoffs in the design.

For usage instructions, examples, and troubleshooting guidance, see:
```
docmgr help yaml-frontmatter-validation
```

## 1. Architecture Overview

The frontmatter validation system consists of four main components:

- **Parser** (`internal/documents/frontmatter.go`): Extracts frontmatter blocks, preprocesses YAML, decodes with position tracking, and wraps failures as diagnostics taxonomies.
- **Preprocessing** (`pkg/frontmatter/frontmatter.go`): Quoting helpers that identify and quote risky scalar values before YAML decoding to reduce parse failures.
- **Validation CLI** (`pkg/commands/validate_frontmatter.go`): Command that reads files, generates fix suggestions, applies auto-fix with backups, and re-parses to verify success.
- **Diagnostics Integration** (`pkg/diagnostics/docmgrctx/frontmatter.go`, `pkg/diagnostics/docmgrrules/frontmatter_rules.go`): Taxonomy context types and rule renderers that surface parse errors with line numbers, snippets, and fix suggestions.

The design separates parsing (which can fail) from fix generation (which operates on raw bytes), allowing the validation verb to attempt repairs even when parsing fails completely.

## 2. Parsing Pipeline

### 2.1. Frontmatter Extraction

The `extractFrontmatter` function in `internal/documents/frontmatter.go` manually scans for `---` delimiters to separate frontmatter from body content.

**Algorithm:**
1. Split file content into lines
2. Find first line that equals `---` (trimmed) → `start`
3. Find next line after `start` that equals `---` (trimmed) → `end`
4. If `start != 0` or `end <= start`, return error: "frontmatter delimiters '---' not found"
5. Extract frontmatter lines: `lines[start+1:end]`
6. Extract body lines: `lines[end+1:]`
7. Return frontmatter bytes, body bytes, and `fmStartLine = start + 2` (1-based, accounting for opening delimiter)

**Tradeoffs:**
- Manual scanning is faster than regex and gives precise line numbers
- Requires exact `---` match (no tolerance for `----` or `--- ` with trailing spaces)
- Delimiter detection happens before YAML parsing, so delimiter errors surface early

### 2.2. YAML Preprocessing

Before decoding, `PreprocessYAML` in `pkg/frontmatter/frontmatter.go` walks top-level key-value lines and quotes risky scalars.

**Algorithm:**
1. Split frontmatter into lines
2. For each line:
   - Skip blank lines and comments (lines starting with `#`)
   - Skip nested structures (lines starting with `- ` or `  ` after trimming)
   - Split on first `:` to get key and value
   - If value is already quoted (`"` or `'`) or complex (`[`, `{`, `|`, `>`), skip
   - If `NeedsQuoting(value)`, replace value with `QuoteValue(value)`
3. Rejoin lines

**What `NeedsQuoting` detects:**
- Leading special characters: `@`, `` ` ``, `#`, `&`, `*`, `!`, `|`, `>`, `%`, `?`
- Colon patterns: `: ` (colon-space) or trailing `:`
- Inline comments: ` #` (space-hash)
- Tabs: `\t`
- Template markers: `{{` or `}}`

**QuoteValue behavior:**
- Wraps value in single quotes: `'value'`
- Escapes internal single quotes: `'value'` → `''value''`

**Tradeoffs:**
- Only processes top-level key-value pairs (heuristic-based, doesn't parse YAML structure)
- Skips nested structures to avoid breaking valid YAML
- Single-quote escaping follows YAML spec but may be surprising (double single quotes)

### 2.3. Position-Aware Decoding

The parser uses `yaml.Decoder` with `yaml.Node` to preserve position information, then decodes the node into `models.Document`.

**Error handling:**
1. On decode error, extract line number from error message using regex: `line ([0-9]+)`
2. Map YAML line number to absolute file line: `absoluteLine = fmStartLine + yamlLine - 1`
3. Build snippet: 3 lines of context (line-1 to line+1) with line numbers and optional caret
4. Classify error message into user-friendly problem text
5. Wrap in `FrontmatterParseTaxonomy` with file, line, col, snippet, problem

**Line number extraction:**
- Regex: `yamlLineRe = regexp.MustCompile(\`line ([0-9]+)\`)`
- YAML parser reports lines relative to frontmatter start (line 1 = first line after opening `---`)
- Absolute line = `fmStartLine + yamlLine - 1`

**Snippet building:**
- Shows 3 lines: `max(1, line-1)` to `min(len(lines), line+1)`
- Format: `%4d | %s\n` with optional caret: `     | %s^\n` (spaces for column position)
- Column position is usually 0 (not reliably extracted from YAML errors)

**Error classification:**
- "mapping values are not allowed" → "mapping values are not allowed (missing quotes before ':' or bad indentation)"
- "did not find expected key" → "did not find expected key (check colons and indentation)"
- "found character that cannot start any token" → "invalid character (likely needs quoting or escaping)"
- Others pass through unchanged

## 3. Fix Generation Heuristics

The `generateFixes` function in `pkg/commands/validate_frontmatter.go` applies multiple heuristics in sequence to repair common issues.

### 3.1. Delimiter Normalization

`normalizeDelimiters` handles missing or malformed delimiters.

**Algorithm:**
1. Find first `---` line → `start`
2. Find next `---`-like line after `start` → `end`
3. If no `start`, treat entire file as frontmatter (wrap it)
4. If no `end`, scan from `start+1` to first blank line or EOF → `end`
5. Extract frontmatter lines: `lines[start+1:end]`
6. Extract body lines: `lines[end+1:]`
7. Peel trailing non-key lines from frontmatter (lines without `:`)
8. Reconstruct: `---\n` + frontmatter + `\n---\n` + body

**Tradeoffs:**
- Heuristic-based (doesn't parse YAML), so may misidentify boundaries in edge cases
- Assumes frontmatter ends at first blank line if no closing delimiter (reasonable default)
- Peels trailing non-key lines to handle common mistake (plain text in frontmatter)

### 3.2. Stray Delimiter Cleanup

`scrubStrayDelimiters` removes lines that look like delimiters but appear inside frontmatter content.

**Algorithm:**
1. Split frontmatter into lines
2. For each line, if trimmed starts with `---`, skip it
3. Rejoin remaining lines

**Tradeoffs:**
- Very aggressive: removes any line starting with `---`, even if it's valid content
- Prevents parser confusion from delimiter-like lines
- May remove legitimate content (rare edge case)

### 3.3. Trailing Non-Key Line Peeling

`peelTrailingNonKeyLines` moves plain text lines (without `:`) from the end of frontmatter into the body.

**Algorithm:**
1. While last line in frontmatter doesn't contain `:`, move it to body
2. Prepend peeled lines to body (preserve order)

**Tradeoffs:**
- Heuristic: assumes lines without `:` are body content, not YAML
- May misidentify valid YAML values that don't contain colons (rare)
- Preserves order by prepending (peeled lines appear before original body)

### 3.4. Scalar Quoting

Reuses `PreprocessYAML` to quote risky scalars (same logic as read-path preprocessing).

**Tradeoffs:**
- Consistent with read-path behavior (same quoting rules)
- Only processes top-level key-value pairs (doesn't handle nested structures)

### 3.5. Fix Orchestration

`generateFixes` chains heuristics:

1. Try `SplitFrontmatter` (normal extraction)
2. If that fails, call `normalizeDelimiters` (fallback)
3. Apply `scrubStrayDelimiters`
4. Apply `peelTrailingNonKeyLines`
5. Apply `PreprocessYAML` (quoting)
6. Reconstruct file: `---\n` + fixed frontmatter + `\n---\n` + body (with peeled lines prepended)
7. Return fix descriptions and fixed content

**Fix descriptions:**
- "Normalize frontmatter delimiters (add missing closing ---)" if delimiter normalization was used
- "Remove stray delimiter lines inside frontmatter" if any were removed
- "Move non key/value lines out of frontmatter" if lines were peeled
- "Quote unsafe scalars (colons, hashes, special leading chars)" if quoting was applied

## 4. Auto-Fix Application

The `applyAutoFix` function writes a backup (`.bak`) and rewrites the file.

**Algorithm:**
1. Write original content to `path + ".bak"`
2. Write fixed content to `path`
3. Re-parse the file
4. If re-parse succeeds, suppress error taxonomy and print success message
5. If re-parse fails, render new taxonomy from the failed re-parse

**Tradeoffs:**
- Always creates backup (even if fix fails) for safety
- Re-parsing verifies fix worked (catches cases where heuristics made things worse)
- Suppresses original error taxonomy on success (avoids confusing "fixed but still broken" messages)
- Shows new error taxonomy on failure (helps debug why fix didn't work)

## 5. Diagnostics Integration

### 5.1. Taxonomy Context

`FrontmatterParseContext` in `pkg/diagnostics/docmgrctx/frontmatter.go` carries:

- `File`: Document path
- `Line`: Absolute line number (1-based)
- `Column`: Column number (usually 0, not reliably extracted)
- `Snippet`: Code snippet with line numbers
- `Problem`: User-friendly problem description
- `Fixes`: Array of suggested fix descriptions (populated by validation verb)

**Constructor:**
- `NewFrontmatterParse`: Creates taxonomy with `StageFrontmatterParse`, `SymptomYAMLSyntax`
- Wraps original error in cause chain via `core.WrapWithCause`

### 5.2. Rule Rendering

`FrontmatterSyntaxRule` in `pkg/diagnostics/docmgrrules/frontmatter_rules.go` renders taxonomies:

**Output format:**
```
YAML/frontmatter syntax error
File: <path>
Line: <line> Col: <col>
Problem: <problem>

Snippet:
<line numbers and code>

Suggested fixes:
  1. <fix description>
  2. <fix description>

Actions:
- Validate frontmatter: docmgr validate frontmatter --doc <path>
```

**When fixes are present:**
- Fixes are attached to context by validation verb (`tryAttachFixes`)
- Rule renders them as a numbered list
- Actions point to validation command (for re-running with `--auto-fix`)

### 5.3. Integration with Diagnostics System

Frontmatter validation is fully integrated with docmgr's diagnostics taxonomy system. Parse errors are wrapped as `FrontmatterParseTaxonomy` objects that flow through the same rendering pipeline as other diagnostics (vocabulary warnings, missing files, stale docs, etc.). This means frontmatter errors appear consistently in `docmgr doctor`, `docmgr list docs`, `docmgr doc search`, and other commands that emit diagnostics.

The validation verb (`docmgr validate frontmatter`) can attach fix suggestions to the taxonomy context, which the rule renderer then surfaces to users. This design allows the same error taxonomy to be used both for reporting (via doctor/list/search) and for interactive fixing (via the validation verb with `--suggest-fixes` or `--auto-fix`).

For details on how to extend the diagnostics system, add new rules, or understand the taxonomy architecture, see:
```
docmgr help diagnostics-taxonomy-and-rules
```

## 6. Write-Path Hardening

All docmgr commands that write frontmatter (`doc add`, `meta update`, `create_ticket`, `doc_move`, `rename_ticket`, `ticket_close`, `import`) use `WriteDocumentWithFrontmatter` in `internal/documents/frontmatter.go`, which:

1. Encodes `models.Document` via `yaml.Encoder`
2. Preprocesses encoded YAML with `PreprocessYAML` to quote risky scalars
3. Writes `---\n` + preprocessed frontmatter + `\n---\n\n` + body

**Tradeoffs:**
- Ensures all CLI-generated frontmatter is valid (prevents errors at write time)
- Preprocessing happens after encoding (may quote values that YAML encoder already quoted, but safe)
- Consistent quoting rules across read and write paths

## 7. Testing Strategy

### 7.1. Unit Tests

- `pkg/commands/validate_frontmatter_test.go`: Tests fix heuristics (delimiter normalization, stray delimiter cleanup)
- `pkg/frontmatter/frontmatter_test.go`: Tests quoting helpers (`NeedsQuoting`, `QuoteValue`, `PreprocessYAML`)
- `internal/documents/frontmatter_test.go`: Tests parsing and error extraction

### 7.2. Smoke Tests

- `test-scenarios/testing-doc-manager/18-validate-frontmatter-smoke.sh`: Exercises validation verb (fail → suggest → auto-fix → success, verifies `.bak` creation)
- `test-scenarios/testing-doc-manager/15-diagnostics-smoke.sh`: Exercises frontmatter parse taxonomy via doctor/list/search

## 8. Known Limitations and Tradeoffs

### 8.1. Preprocessing Limitations

- **Top-level only**: `PreprocessYAML` only processes top-level key-value pairs. Nested structures are skipped, so colons in nested values may still cause errors.
- **Heuristic-based**: Doesn't parse YAML structure, so may misidentify nested content as top-level.

### 8.2. Fix Heuristics Limitations

- **Delimiter normalization**: Assumes frontmatter ends at first blank line if no closing delimiter (may be wrong if body starts with blank lines).
- **Stray delimiter cleanup**: Removes any line starting with `---`, even if it's valid content (rare edge case).
- **Trailing line peeling**: Assumes lines without `:` are body content (may misidentify valid YAML values).

### 8.3. Error Message Limitations

- **Column numbers**: Usually 0 (YAML parser doesn't reliably report column positions).
- **Complex errors**: Some YAML errors are hard to classify (fall back to raw error message).

### 8.4. Design Tradeoffs

- **Separation of parsing and fixing**: Allows fix generation even when parsing fails completely, but fix heuristics operate on raw bytes (less precise than parsed structure).
- **Backup always created**: Safe but may accumulate `.bak` files if auto-fix runs multiple times.
- **Re-parse verification**: Catches cases where heuristics made things worse, but adds latency.

## 9. Extension Points

### 9.1. Adding New Fix Heuristics

1. Add heuristic function to `pkg/commands/validate_frontmatter.go`
2. Call it in `generateFixes` (order matters: delimiters → cleanup → quoting)
3. Add fix description to returned `fixes` array
4. Add unit test in `validate_frontmatter_test.go`
5. Update smoke test if behavior changes significantly

### 9.2. Improving Error Classification

1. Add pattern to `classifyYAMLError` in `internal/documents/frontmatter.go`
2. Test with real error messages from YAML parser
3. Ensure user-friendly description explains the problem clearly

### 9.3. Enhancing Preprocessing

1. Extend `NeedsQuoting` to detect new risky patterns
2. Update `PreprocessYAML` to handle nested structures (requires YAML parsing)
3. Add tests for edge cases

## 10. File Map

**Core parsing:**
- `internal/documents/frontmatter.go`: `ReadDocumentWithFrontmatter`, `extractFrontmatter`, `buildSnippet`, `classifyYAMLError`, `WriteDocumentWithFrontmatter`, `SplitFrontmatter`

**Preprocessing:**
- `pkg/frontmatter/frontmatter.go`: `NeedsQuoting`, `QuoteValue`, `PreprocessYAML`

**Validation CLI:**
- `pkg/commands/validate_frontmatter.go`: `ValidateFrontmatterCommand`, `generateFixes`, `normalizeDelimiters`, `scrubStrayDelimiters`, `peelTrailingNonKeyLines`, `applyAutoFix`

**Diagnostics:**
- `pkg/diagnostics/docmgrctx/frontmatter.go`: `FrontmatterParseContext`, `NewFrontmatterParseTaxonomy`
- `pkg/diagnostics/docmgrrules/frontmatter_rules.go`: `FrontmatterSyntaxRule`

**Tests:**
- `pkg/commands/validate_frontmatter_test.go`: Fix heuristic unit tests
- `test-scenarios/testing-doc-manager/18-validate-frontmatter-smoke.sh`: Validation verb smoke test

