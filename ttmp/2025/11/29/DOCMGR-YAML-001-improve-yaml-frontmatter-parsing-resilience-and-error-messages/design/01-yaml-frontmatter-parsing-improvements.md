---
Title: YAML Frontmatter Parsing Improvements
Ticket: DOCMGR-YAML-001
Status: active
Topics:
    - yaml
    - ux
    - errors
DocType: design
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: 'Design for improving YAML frontmatter parsing with better errors, auto-quoting, validation, and optional auto-fixing'
LastUpdated: 2025-11-29T18:30:00-05:00
---

# YAML Frontmatter Parsing Improvements

## Problem Statement

Users encounter cryptic YAML parsing errors when creating or editing document frontmatter, particularly when:

1. **Unquoted strings contain special characters** (colons, hashes, etc.)
   - Example: `Summary: Updated guide covering recent progress - position integration`
   - Error: `yaml: line 12: cannot unmarshal !!str into map[string]interface{}`
   
2. **Error messages lack context**
   - No field name shown
   - No line/column position highlighted
   - No suggestion for how to fix

3. **Manual quoting required**
   - Users must remember YAML quoting rules
   - Trial-and-error debugging

4. **No validation before commit**
   - Errors discovered at parse time, not during editing
   - No `validate-frontmatter` command

## Goals

### Primary Goals

1. **Better error messages** — Show field name, line/column, problematic value, and fix suggestion
2. **Auto-quoting on write** — When docmgr writes frontmatter, always quote strings with special chars
3. **Validation command** — `docmgr validate-frontmatter <file>` to check before parsing
4. **Optional auto-fix** — `--auto-fix-frontmatter` flag to attempt automatic repairs

### Secondary Goals

5. **Position-aware parsing** — Use `yaml.Node` API for accurate error locations
6. **Field-specific validation** — Custom validators with helpful hints per field

## Design

### 1. Enhanced Error Messages

#### Current Behavior

```
Error: yaml: line 12: cannot unmarshal !!str into map[string]interface {}
```

#### Proposed Behavior

```
Error parsing frontmatter in playbook/02-continuation-guide.md:

  Line 12, Column 10: Field 'Summary'
  
  11 | RelatedFiles: []
  12 | Summary: Updated guide covering recent progress - position integration
     |          ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
  13 | LastUpdated: 2025-11-29
  
  Problem: Value contains unquoted colon (:)
  
  YAML interprets colons as key-value separators. When your value contains
  a colon, you must quote the entire value.
  
  Fix (option 1): Add single quotes
    Summary: 'Updated guide covering recent progress - position integration'
  
  Fix (option 2): Add double quotes
    Summary: "Updated guide covering recent progress - position integration"
  
  Fix (option 3): Use block scalar (|)
    Summary: |
      Updated guide covering recent progress - position integration
  
  Hint: Run 'docmgr validate-frontmatter <file>' to check before committing.
```

#### Implementation

```go
package frontmatter

type ParseError struct {
    File      string
    Field     string
    Line      int
    Column    int
    Value     string
    Hint      string
    Fixes     []string
    Original  error
}

func (e *ParseError) Error() string {
    var buf strings.Builder
    
    buf.WriteString(fmt.Sprintf("Error parsing frontmatter in %s:\n\n", e.File))
    
    if e.Field != "" && e.Line > 0 {
        buf.WriteString(fmt.Sprintf("  Line %d, Column %d: Field '%s'\n\n", 
            e.Line, e.Column, e.Field))
    }
    
    // Show snippet with context (similar to compiler errors)
    if snippet := extractSnippet(e.File, e.Line); snippet != "" {
        buf.WriteString(snippet)
        buf.WriteString("\n\n")
    }
    
    buf.WriteString(fmt.Sprintf("  Problem: %s\n\n", detectProblemType(e.Original)))
    
    // Show fix suggestions
    if len(e.Fixes) > 0 {
        for i, fix := range e.Fixes {
            buf.WriteString(fmt.Sprintf("  Fix (option %d): %s\n", i+1, fix))
        }
        buf.WriteString("\n")
    }
    
    if e.Hint != "" {
        buf.WriteString(fmt.Sprintf("  Hint: %s\n", e.Hint))
    }
    
    return buf.String()
}

// Detect common error types and provide appropriate messages
func detectProblemType(err error) string {
    msg := err.Error()
    
    if strings.Contains(msg, "cannot unmarshal !!str") {
        return "Value contains unquoted colon (:)"
    }
    if strings.Contains(msg, "did not find expected key") {
        return "Invalid YAML structure (check indentation and colons)"
    }
    if strings.Contains(msg, "mapping values are not allowed") {
        return "Unexpected colon in value (needs quoting)"
    }
    
    return msg
}

// Generate fix suggestions based on the value
func suggestFixes(field, value string) []string {
    var fixes []string
    
    if needsQuoting(value) {
        // Single quotes (simplest)
        fixes = append(fixes, fmt.Sprintf("%s: '%s'", field, 
            strings.ReplaceAll(value, "'", "''")))
        
        // Double quotes (if value contains single quotes)
        if strings.Contains(value, "'") {
            fixes = append(fixes, fmt.Sprintf("%s: \"%s\"", field,
                strings.ReplaceAll(value, "\"", "\\\"")))
        }
        
        // Block scalar (for multi-line or very long values)
        if len(value) > 80 || strings.Contains(value, "\n") {
            fixes = append(fixes, fmt.Sprintf("%s: |\n  %s", field, value))
        }
    }
    
    return fixes
}
```

---

### 2. Auto-Quoting Preprocessor

#### Implementation

```go
package frontmatter

// PreprocessFrontmatter adds quotes to values that need them
func PreprocessFrontmatter(raw []byte) []byte {
    lines := bytes.Split(raw, []byte("\n"))
    var result [][]byte
    
    inFrontmatter := false
    frontmatterCount := 0
    
    for _, line := range lines {
        trimmed := bytes.TrimSpace(line)
        
        // Track frontmatter delimiters (---)
        if bytes.Equal(trimmed, []byte("---")) {
            frontmatterCount++
            inFrontmatter = (frontmatterCount == 1)
            result = append(result, line)
            continue
        }
        
        // Only process lines within frontmatter
        if !inFrontmatter || frontmatterCount != 1 {
            result = append(result, line)
            continue
        }
        
        // Skip list items and nested structures
        if bytes.HasPrefix(trimmed, []byte("- ")) || 
           bytes.HasPrefix(trimmed, []byte("  ")) {
            result = append(result, line)
            continue
        }
        
        // Check if it's a key-value pair
        if idx := bytes.IndexByte(line, ':'); idx > 0 {
            key := bytes.TrimSpace(line[:idx])
            value := bytes.TrimSpace(line[idx+1:])
            
            // Skip if already quoted, empty, or is a list/object
            if len(value) == 0 || 
               value[0] == '"' || value[0] == '\'' ||
               value[0] == '[' || value[0] == '{' ||
               value[0] == '|' || value[0] == '>' {
                result = append(result, line)
                continue
            }
            
            // Check if value needs quoting
            if needsQuoting(value) {
                // Preserve indentation
                indent := line[:len(line)-len(bytes.TrimLeft(line, " \t"))]
                quoted := quoteValue(value)
                newLine := append(indent, key...)
                newLine = append(newLine, []byte(": ")...)
                newLine = append(newLine, quoted...)
                result = append(result, newLine)
                continue
            }
        }
        
        result = append(result, line)
    }
    
    return bytes.Join(result, []byte("\n"))
}

// needsQuoting checks if a value contains characters that require quoting
func needsQuoting(value []byte) bool {
    // Check for colon followed by space (YAML key-value separator)
    if bytes.Contains(value, []byte(": ")) {
        return true
    }
    
    // Check for trailing colon
    if bytes.HasSuffix(bytes.TrimSpace(value), []byte(":")) {
        return true
    }
    
    // Check for special chars at start
    if len(value) > 0 {
        first := value[0]
        if first == '@' || first == '`' || first == '#' ||
           first == '&' || first == '*' || first == '!' ||
           first == '|' || first == '>' || first == '%' {
            return true
        }
    }
    
    // Check for other problematic patterns
    problematic := []string{
        " #",   // Inline comment
        "\t",   // Tab character
        "{{",   // Template-like syntax
    }
    
    for _, pattern := range problematic {
        if bytes.Contains(value, []byte(pattern)) {
            return true
        }
    }
    
    return false
}

// quoteValue wraps a value in single quotes, escaping internal quotes
func quoteValue(value []byte) []byte {
    // Escape any single quotes by doubling them
    escaped := bytes.ReplaceAll(value, []byte("'"), []byte("''"))
    
    // Wrap in single quotes
    result := []byte("'")
    result = append(result, escaped...)
    result = append(result, '\'')
    
    return result
}
```

---

### 3. Validation Command

#### CLI Interface

```bash
# Validate single file
docmgr validate-frontmatter file.md

# Validate all docs in a ticket
docmgr validate-frontmatter --ticket TICKET-001

# Show fixes (don't apply)
docmgr validate-frontmatter file.md --suggest-fixes

# Apply fixes automatically
docmgr validate-frontmatter file.md --auto-fix
```

#### Implementation

```go
package cmd

func validateFrontmatterCmd() *cobra.Command {
    var (
        ticket       string
        suggestFixes bool
        autoFix      bool
    )
    
    cmd := &cobra.Command{
        Use:   "validate-frontmatter [FILE]",
        Short: "Validate YAML frontmatter syntax and suggest fixes",
        Long: `Validate YAML frontmatter in markdown files.
        
Checks for:
- Unquoted strings with special characters
- Invalid YAML syntax
- Unknown vocabulary values (warnings only)
- Missing required fields

Use --suggest-fixes to see recommended fixes without applying them.
Use --auto-fix to automatically repair issues (creates backup).`,
        
        RunE: func(cmd *cobra.Command, args []string) error {
            ctx := cmd.Context()
            
            if ticket != "" {
                return validateTicketFrontmatter(ctx, ticket, autoFix)
            }
            
            if len(args) == 0 {
                return fmt.Errorf("provide file path or --ticket")
            }
            
            return validateFileFrontmatter(ctx, args[0], suggestFixes, autoFix)
        },
    }
    
    cmd.Flags().StringVar(&ticket, "ticket", "", "Validate all docs in ticket")
    cmd.Flags().BoolVar(&suggestFixes, "suggest-fixes", false, "Show fix suggestions")
    cmd.Flags().BoolVar(&autoFix, "auto-fix", false, "Automatically fix issues")
    
    return cmd
}

func validateFileFrontmatter(ctx context.Context, path string, suggest, autoFix bool) error {
    raw, err := os.ReadFile(path)
    if err != nil {
        return err
    }
    
    // Try parsing
    fm, err := frontmatter.Parse(raw)
    if err != nil {
        // Parse error - check if fixable
        if parseErr, ok := err.(*frontmatter.ParseError); ok {
            if suggest || autoFix {
                fixes := suggestFixes(parseErr.Field, parseErr.Value)
                
                fmt.Println(parseErr.Error())
                fmt.Println("\nSuggested fixes:")
                for i, fix := range fixes {
                    fmt.Printf("  %d. %s\n", i+1, fix)
                }
                
                if autoFix && len(fixes) > 0 {
                    // Apply first fix
                    return applyFix(path, raw, parseErr, fixes[0])
                }
                
                return nil // Don't fail if showing suggestions
            }
        }
        
        return err
    }
    
    // Parse succeeded - check for warnings
    warnings := validateFrontmatterContent(fm)
    if len(warnings) > 0 {
        fmt.Printf("✓ Frontmatter is valid YAML\n\n")
        fmt.Printf("Warnings (%d):\n", len(warnings))
        for _, w := range warnings {
            fmt.Printf("  - %s\n", w)
        }
        return nil
    }
    
    fmt.Printf("✓ Frontmatter is valid with no warnings\n")
    return nil
}
```

#### Output Examples

**No errors:**
```bash
$ docmgr validate-frontmatter doc.md
✓ Frontmatter is valid YAML

Warnings (1):
  - Unknown topic 'custom-topic' (not in vocabulary.yaml)
    Run: docmgr vocab add --category topics --slug custom-topic
```

**With errors and suggestions:**
```bash
$ docmgr validate-frontmatter doc.md --suggest-fixes

Error parsing frontmatter in doc.md:

  Line 12, Column 10: Field 'Summary'
  
  Problem: Value contains unquoted colon (:)

Suggested fixes:
  1. Summary: 'Updated guide covering recent progress - position integration'
  2. Summary: "Updated guide covering recent progress - position integration"
  3. Summary: |
      Updated guide covering recent progress - position integration
```

**Auto-fix:**
```bash
$ docmgr validate-frontmatter doc.md --auto-fix

Backing up doc.md to doc.md.bak
Applying fix: Summary: 'Updated guide covering recent progress - position integration'
✓ Fixed frontmatter in doc.md

Run 'docmgr doctor --doc doc.md' to verify.
```

---

### 4. Optional Auto-Fix Mode

#### Modes

1. **Manual (default)** — Parse fails, show error with suggestions
2. **Suggest (`--suggest-fixes`)** — Parse fails, show suggestions, don't modify file
3. **Auto-fix (`--auto-fix`)** — Parse fails, automatically apply first fix, create backup

#### Safety Features

- Creates `.bak` backup before modifying
- Only applies simplest fix (single quotes)
- Reports what was changed
- Suggests validation after fix

#### When to Use

- **Manual**: Normal workflow (fail fast)
- **Suggest**: Learning YAML syntax, uncertain about fix
- **Auto-fix**: Bulk operations, CI pipelines, legacy doc cleanup

---

### 5. Ensure Quoting on Write

When docmgr writes frontmatter (via `meta update`, `doc add`, `changelog`, etc.), always quote strings appropriately.

#### Current Behavior (Hypothetical)

```go
// May write unquoted values
fmt.Fprintf(w, "Summary: %s\n", summary)  // ❌ No quoting
```

#### Proposed Behavior

```go
// Always quote values properly
func writeFrontmatterField(w io.Writer, key string, value interface{}) error {
    quotedValue := quoteIfNeeded(value)
    _, err := fmt.Fprintf(w, "%s: %s\n", key, quotedValue)
    return err
}

func quoteIfNeeded(value interface{}) string {
    s, ok := value.(string)
    if !ok {
        // Marshal non-string types with yaml.Marshal (handles quoting)
        out, _ := yaml.Marshal(value)
        return strings.TrimSpace(string(out))
    }
    
    if s == "" {
        return `""`
    }
    
    // Check if quoting needed
    if needsQuotingStr(s) {
        // Use single quotes, escape internal quotes
        escaped := strings.ReplaceAll(s, "'", "''")
        return fmt.Sprintf("'%s'", escaped)
    }
    
    return s
}

func needsQuotingStr(s string) bool {
    // Same logic as needsQuoting() but for strings
    return strings.Contains(s, ": ") ||
           strings.HasSuffix(strings.TrimSpace(s), ":") ||
           len(s) > 0 && isSpecialFirstChar(s[0])
}

func isSpecialFirstChar(c byte) bool {
    return c == '@' || c == '`' || c == '#' || c == '&' || 
           c == '*' || c == '!' || c == '|' || c == '>' || c == '%'
}
```

---

### 6. Position-Aware Parsing

Use `yaml.Node` API to preserve line/column information:

```go
func ParseWithPositions(raw []byte) (*Frontmatter, map[string]Position, error) {
    var root yaml.Node
    
    if err := yaml.Unmarshal(raw, &root); err != nil {
        // Extract line/column from error
        line, col := extractLineColumn(err)
        return nil, nil, &ParseError{
            Line:     line,
            Column:   col,
            Original: err,
        }
    }
    
    // Build position map
    positions := make(map[string]Position)
    fm := &Frontmatter{}
    
    if err := parseNodeWithPositions(&root, fm, positions, ""); err != nil {
        return nil, nil, err
    }
    
    return fm, positions, nil
}

type Position struct {
    Line   int
    Column int
}

func parseNodeWithPositions(node *yaml.Node, target interface{}, positions map[string]Position, path string) error {
    if node.Kind != yaml.MappingNode {
        return fmt.Errorf("expected mapping, got %v", node.Kind)
    }
    
    for i := 0; i < len(node.Content); i += 2 {
        keyNode := node.Content[i]
        valNode := node.Content[i+1]
        
        key := keyNode.Value
        fieldPath := key
        if path != "" {
            fieldPath = path + "." + key
        }
        
        // Record position
        positions[fieldPath] = Position{
            Line:   keyNode.Line,
            Column: keyNode.Column,
        }
        
        // Parse value and populate target struct
        // (delegate to reflection-based unmarshal for simplicity)
    }
    
    return nil
}
```

---

### 7. Field-Specific Validation

Define validation rules per field with custom hints:

```go
type FieldValidator struct {
    Name      string
    Required  bool
    Type      string
    MaxLength int
    Pattern   *regexp.Regexp
    Validator func(any) error
    Hint      string
}

var frontmatterSchema = []FieldValidator{
    {
        Name:      "Title",
        Required:  true,
        Type:      "string",
        MaxLength: 100,
        Validator: func(v any) error {
            s, ok := v.(string)
            if !ok {
                return fmt.Errorf("must be a string")
            }
            if len(s) < 3 {
                return fmt.Errorf("must be at least 3 characters")
            }
            if len(s) > 100 {
                return fmt.Errorf("must be under 100 characters (got %d)", len(s))
            }
            return nil
        },
        Hint: "Short, descriptive title for the document.",
    },
    {
        Name:      "Summary",
        Required:  true,
        Type:      "string",
        MaxLength: 200,
        Validator: func(v any) error {
            s, ok := v.(string)
            if !ok {
                return fmt.Errorf("must be a string")
            }
            if len(s) > 200 {
                return fmt.Errorf("must be under 200 characters (got %d)", len(s))
            }
            return nil
        },
        Hint: "Brief description. Use quotes if it contains colons or special chars.",
    },
    {
        Name:     "Topics",
        Required: true,
        Type:     "[]string",
        Validator: func(v any) error {
            topics, ok := v.([]string)
            if !ok {
                return fmt.Errorf("must be a list of strings")
            }
            if len(topics) == 0 {
                return fmt.Errorf("must have at least one topic")
            }
            return nil
        },
        Hint: "List of topics (e.g., - backend, - frontend). Check vocabulary with 'docmgr vocab list'.",
    },
    {
        Name:     "Status",
        Required: false,
        Type:     "string",
        Pattern:  regexp.MustCompile(`^(draft|active|review|complete|archived)$`),
        Hint:     "One of: draft, active, review, complete, archived. Check vocabulary with 'docmgr vocab list --category status'.",
    },
}

func validateFrontmatterContent(fm *Frontmatter) []ValidationWarning {
    var warnings []ValidationWarning
    
    for _, validator := range frontmatterSchema {
        value := getFieldValue(fm, validator.Name)
        
        // Check required
        if validator.Required && value == nil {
            warnings = append(warnings, ValidationWarning{
                Field:   validator.Name,
                Message: "Required field missing",
                Hint:    validator.Hint,
            })
            continue
        }
        
        // Run custom validator
        if validator.Validator != nil && value != nil {
            if err := validator.Validator(value); err != nil {
                warnings = append(warnings, ValidationWarning{
                    Field:   validator.Name,
                    Message: err.Error(),
                    Hint:    validator.Hint,
                })
            }
        }
        
        // Check pattern
        if validator.Pattern != nil {
            if s, ok := value.(string); ok {
                if !validator.Pattern.MatchString(s) {
                    warnings = append(warnings, ValidationWarning{
                        Field:   validator.Name,
                        Message: fmt.Sprintf("Invalid value: %s", s),
                        Hint:    validator.Hint,
                    })
                }
            }
        }
    }
    
    return warnings
}
```

---

## Implementation Plan

### Phase 1: Enhanced Errors (1-2 days)

- [ ] Create `pkg/frontmatter/errors.go` with `ParseError` type
- [ ] Add `detectProblemType()` and `suggestFixes()` functions
- [ ] Update all frontmatter parsing to wrap errors with `ParseError`
- [ ] Add snippet extraction with line highlighting
- [ ] Test with common error scenarios

### Phase 2: Auto-Quoting on Write (1 day)

- [ ] Create `pkg/frontmatter/quote.go` with quoting utilities
- [ ] Add `needsQuoting()` and `quoteValue()` functions
- [ ] Update `meta update`, `doc add`, `changelog` to use auto-quoting
- [ ] Add tests for edge cases (already-quoted, empty, special chars)

### Phase 3: Preprocessing Layer (1-2 days)

- [ ] Create `pkg/frontmatter/preprocess.go`
- [ ] Implement `PreprocessFrontmatter()` function
- [ ] Add opt-in preprocessing to parse flow
- [ ] Test with various frontmatter formats
- [ ] Document when preprocessing is enabled

### Phase 4: Validation Command (1 day)

- [ ] Add `docmgr validate-frontmatter` command
- [ ] Support single file and --ticket modes
- [ ] Add --suggest-fixes and --auto-fix flags
- [ ] Create backup files before auto-fixing
- [ ] Integration tests

### Phase 5: Position-Aware Parsing (Optional, 1-2 days)

- [ ] Refactor to use `yaml.Node` API
- [ ] Build position map during parsing
- [ ] Use positions in error messages
- [ ] Update error snippets to use actual positions

### Phase 6: Field Validation (Optional, 1 day)

- [ ] Define `frontmatterSchema` with validators
- [ ] Implement `validateFrontmatterContent()`
- [ ] Add hints per field
- [ ] Integrate with `validate-frontmatter` command

---

## Testing Strategy

### Unit Tests

```go
func TestNeedsQuoting(t *testing.T) {
    tests := []struct {
        value    string
        expected bool
    }{
        {"simple", false},
        {"with: colon", true},
        {"ends:", true},
        {"@special", true},
        {"# comment", true},
        {"already 'quoted'", false},
        {`"already quoted"`, false},
    }
    
    for _, tt := range tests {
        result := needsQuotingStr(tt.value)
        if result != tt.expected {
            t.Errorf("needsQuoting(%q) = %v, want %v", 
                tt.value, result, tt.expected)
        }
    }
}

func TestPreprocessFrontmatter(t *testing.T) {
    input := []byte(`---
Title: Simple Title
Summary: Complex summary with: colon and - dash
Topics:
  - yaml
  - ux
---

# Document body
`)
    
    result := PreprocessFrontmatter(input)
    
    // Should quote Summary but not Title
    if !bytes.Contains(result, []byte("Summary: 'Complex summary")) {
        t.Error("Summary should be quoted")
    }
    if bytes.Contains(result, []byte("Title: '")) {
        t.Error("Title should not be quoted")
    }
}
```

### Integration Tests

```bash
# Test validation command
docmgr validate-frontmatter testdata/bad-frontmatter.md
# Should show error with suggestions

# Test auto-fix
docmgr validate-frontmatter testdata/bad-frontmatter.md --auto-fix
# Should create .bak and fix file

# Test with ticket
docmgr validate-frontmatter --ticket TEST-001 --suggest-fixes
# Should validate all docs in ticket
```

---

## Backward Compatibility

### No Breaking Changes

- Existing valid frontmatter continues to work
- Auto-quoting only affects new writes
- Preprocessing is opt-in (via flag or config)
- Enhanced errors provide more info (still failures)

### Migration Path

1. **Phase 1**: Deploy enhanced errors (no behavior change, just better messages)
2. **Phase 2**: Deploy auto-quoting on write (gradual improvement)
3. **Phase 3**: Add validation command (new feature)
4. **Phase 4**: Add auto-fix mode (optional, for cleanup)

Users can adopt incrementally.

---

## Alternatives Considered

### Alternative 1: Switch to goccy/go-yaml

**Pros:**
- Better error messages out of the box
- Built-in validator support
- More lenient by default

**Cons:**
- Breaking change (different API)
- Migration effort across entire codebase
- Less mature than yaml.v3

**Decision:** Not worth migration. Enhance current parser instead.

### Alternative 2: Always Use Block Scalars

Force all string values to use `|` or `>` notation:

```yaml
Summary: |
  Updated guide covering recent progress - position integration
```

**Pros:**
- Never need quoting
- Clear boundaries

**Cons:**
- Verbose for short strings
- Unfamiliar to users
- Harder to edit inline

**Decision:** Reject. Block scalars are better for long/multi-line content, not short summaries.

### Alternative 3: Do Nothing

Rely on users learning YAML syntax.

**Pros:**
- No implementation work
- "Industry standard" (everyone deals with YAML quirks)

**Cons:**
- Poor UX
- Friction in workflows
- Wastes user time on debugging

**Decision:** Reject. We can do better.

---

## Open Questions

1. **Should auto-quoting be enabled by default or opt-in?**
   - Recommendation: Enabled by default when writing, opt-in when reading (preprocessing)

2. **Should validation warnings fail the command or just warn?**
   - Recommendation: Warnings only (exit 0). Use `--fail-on warning` flag if needed.

3. **What should auto-fix do with multiple possible fixes?**
   - Recommendation: Apply simplest (single quotes), log others as alternatives

4. **Should we support YAML 1.2 vs 1.1?**
   - Current: yaml.v3 uses YAML 1.1
   - Recommendation: Stay with 1.1 (most compatible)

---

## Success Metrics

1. **Error resolution time** — Users should fix frontmatter errors in < 30 seconds (down from ~5 minutes)
2. **Error recurrence** — Same user shouldn't hit same error twice (auto-quoting prevents repeat mistakes)
3. **Support requests** — Fewer "YAML syntax error" support requests
4. **Adoption** — Users actually run `validate-frontmatter` before commits

---

## References

- `gopkg.in/yaml.v3` documentation: https://pkg.go.dev/gopkg.in/yaml.v3
- `goccy/go-yaml` comparison: https://github.com/goccy/go-yaml
- YAML spec (colons): https://yaml.org/spec/1.2/spec.html#id2788859
- Related work: YAPP-ERROR-FEEDBACK-001 (similar error enhancement pattern)

---

## Next Steps

1. Review this design document
2. Prioritize phases (start with Phase 1 & 2)
3. Implement enhanced errors
4. Implement auto-quoting on write
5. Add validation command
6. Document new commands in help system
