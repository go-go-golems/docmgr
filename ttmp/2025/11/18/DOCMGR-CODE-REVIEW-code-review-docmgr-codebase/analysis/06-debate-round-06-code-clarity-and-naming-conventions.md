---
Title: Debate Round 06 — Code Clarity and Naming Conventions
Ticket: DOCMGR-CODE-REVIEW
Status: active
Topics:
    - docmgr
    - code-review
DocType: analysis
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: "Good: Consistent Glazed patterns, clear function names. Issues: TTMPConfig vs TicketDirectory naming, abbreviations (doc/cfg). Consensus: Add package docs, rename TTMPConfig→WorkspaceConfig"
LastUpdated: 2025-11-18T11:30:00.000000000-05:00
---

# Debate Round 06 — Code Clarity and Naming Conventions

## Question

**"Are functions, variables, and types named clearly? Is the code self-documenting or needlessly obscure?"**

## Pre-Debate Research

### Type Naming Analysis

```go
// pkg/models/document.go
type Document struct { ... }           // ✅ Clear
type Vocabulary struct { ... }         // ✅ Clear
type VocabItem struct { ... }          // ✅ Abbreviated but acceptable
type RelatedFile struct { ... }        // ✅ Clear
type RelatedFiles []RelatedFile        // ✅ Clear custom type
type ExternalSource struct { ... }     // ✅ Clear
type TicketDirectory struct { ... }    // ⚠️ Misleading name (not a directory)

// pkg/commands/config.go
type TTMPConfig struct { ... }         // ⚠️ Acronym, unclear to newcomers
```

###Function Naming Patterns

```bash
# Good examples (self-documenting):
func NewAddCommand() (*AddCommand, error)
func readDocumentFrontmatter(path string) (*models.Document, error)
func splitFrontmatter(content string) []string
func writeDocumentWithFrontmatter(path string, doc *models.Document, content string, force bool) error

# Abbreviated examples:
func parsedLayers.GetSettings()  # "parsed" could be "parsedParameters"
```

### Variable Naming

```bash
# Good: Descriptive
settings, err := parsedLayers.GetSettings()
doc, body, err := ReadWithFrontmatter(path)
workspaces, err := DiscoverTicketWorkspaces(root)

# Acceptable abbreviations:
cfg := TTMPConfig{}    # Standard cfg for config
docs := []Document{}   # Standard plural
```

### Package Documentation

```bash
# Command: Check for package-level comments
grep -rn "^// Package" pkg/
```

**Result**: Only `pkg/doc/doc.go` and `pkg/utils/slug.go` have package docs.

**Missing:** `pkg/models`, `pkg/commands` have no package-level documentation.

---

## Opening Statements

### Casey (The New User)

*[Frustrated by terminology]*

I have **naming confusion**:

**1. "TTMP" — What does it mean?**

I see `TTMPConfig`, `.ttmp.yaml`, `--root` defaulting to "ttmp". 

**Nowhere** in the code is "TTMP" explained. I had to ask: "Temporary" or "To The Max Project" or what?

Eventually found it stands for **"Today/Tomorrow/Today/Tomorrow Mutable Project"** or something? Still unclear.

**Better name:** `WorkspaceConfig` (self-documenting)

**2. "TicketDirectory" — Not actually a directory**

```go
type TicketDirectory struct {
	Ticket   string
	Path     string
	Document *Document
}
```

This is **metadata about a ticket**, not a directory!

**Better name:** `TicketWorkspace` or `TicketMetadata`

**3. Package names vs. directory names**

- Directory: `pkg/commands/`
- But it's going away → `cmd/docmgr/cmds/`
- Confusing during migration

**4. Abbreviations I had to look up:**

- `cfg` — okay, standard
- `doc` — okay, domain-specific
- `vocab` — acceptable shortening of "vocabulary"
- `ttmp` — **not okay**, unexplained acronym

**My position:** Names should be **self-explanatory** or **documented**. Acronyms need explanation.

---

### Alex Rodriguez (The Architect)

*[Analyzing patterns]*

Let me do a **naming consistency audit**:

**Good patterns:**

```go
// Commands use consistent "New[X]Command" pattern
func NewAddCommand() (*AddCommand, error)
func NewSearchCommand() (*SearchCommand, error)
func NewCreateTicketCommand() (*CreateTicketCommand, error)
```

✅ **Excellent**: Predictable, discoverable pattern across 20 commands.

**Inconsistent patterns:**

```go
// Some commands use "CommandDescription", others embed it
type AddCommand struct {
	*cmds.CommandDescription
}

// Settings structs use inconsistent naming
type AddSettings struct { ... }        // Suffix
type ChangelogUpdateSettings struct { ... }  // Infix
```

⚠️ **Acceptable** but could be more consistent.

**Naming conventions I observe:**

1. **Command types**: `[Action]Command` (good)
2. **Settings types**: `[Action]Settings` (good)
3. **Constructor functions**: `New[Action]Command()` (good)
4. **Method names**: `RunIntoGlazeProcessor()` (Glazed convention, long but clear)

**My recommendations:**

**1. Add package documentation:**

```go
// Package models defines the core data structures for docmgr documents.
// The Document type represents a managed markdown document with YAML frontmatter,
// and Vocabulary defines allowed values for document metadata fields.
package models
```

**2. Rename unclear types:**

- `TTMPConfig` → `WorkspaceConfig`
- `TicketDirectory` → `TicketWorkspace`

**3. Add godoc comments for exported types:**

```go
// WorkspaceConfig holds configuration for the documentation workspace.
// It specifies the root directory where ticket workspaces are stored
// and the path to the vocabulary file.
type WorkspaceConfig struct {
	Root       string `yaml:"root"`
	Vocabulary string `yaml:"vocabulary"`
}
```

---

### Sarah Chen (The Pragmatist)

*[Pragmatic assessment]*

Let me separate **real problems** from **nitpicks**:

**Real problems (fix these):**

1. ✅ **TTMPConfig** — Rename to `WorkspaceConfig`
2. ✅ **TicketDirectory** — Rename to `TicketWorkspace`
3. ✅ **Missing package docs** — Add to `pkg/models`, `pkg/commands`

**Nitpicks (don't bother):**

1. ❌ `cfg` abbreviation — Standard in Go
2. ❌ `doc` abbreviation — Domain-specific, clear in context
3. ❌ `vocab` abbreviation — Acceptable shortening
4. ❌ Long function names — Self-documenting is good

**Casey's "TTMP" confusion** is valid. But the fix isn't to explain "TTMP"—it's to **stop using it**.

**Migration path:**

**Phase 1: Aliases (non-breaking)**

```go
// Deprecated: Use WorkspaceConfig instead
type TTMPConfig = WorkspaceConfig

type WorkspaceConfig struct {
	Root       string `yaml:"root"`
	Vocabulary string `yaml:"vocabulary"`
}
```

**Phase 2: Update docs**

- README: Change "TTMP root" → "workspace root"
- Help text: Change "ttmp directory" → "documentation workspace"

**Phase 3: Rename default**

- Flag default: "ttmp" → "workspace" (or keep "ttmp" for backward compat)
- Config file: `.ttmp.yaml` → `.docmgr.yaml` (support both)

**Low-effort, high-impact** naming fixes.

---

## Rebuttals

### Casey (The New User) — Rebuttal

*[Grateful for the acknowledgment]*

Thank you for taking my confusion seriously!

Sarah's migration path is perfect. Aliases for backward compatibility, gradual rename, users barely notice.

One more request: **Add a glossary to README.md**

```markdown
## Glossary

- **Workspace**: Root directory containing ticket documentation (default: `ttmp`)
- **Ticket**: Work item identifier (e.g., `MEN-3475`)
- **Ticket Workspace**: Directory structure for a single ticket's documentation
- **Doc Type**: Category of document (analysis, design-doc, playbook, reference, til)
- **Vocabulary**: Controlled vocabulary for topics, doc types, and intents
```

This would have saved me hours of confusion.

---

### Alex Rodriguez (The Architect) — Rebuttal

*[Agreeing with Sarah's phased approach]*

Sarah's migration path is solid. Let me add **godoc examples** to make the API clearer:

```go
// WorkspaceConfig holds configuration for the documentation workspace.
//
// Example config file (.docmgr.yaml):
//
//	root: ~/projects/myapp/docs
//	vocabulary: ~/projects/myapp/docs/vocabulary.yaml
//
// The root directory contains ticket workspaces organized by date:
//
//	root/
//	  2025/
//	    11/
//	      18/
//	        MEN-3475-add-feature/
//	          analysis/
//	          design-doc/
//	          playbook/
type WorkspaceConfig struct {
	Root       string `yaml:"root"`
	Vocabulary string `yaml:"vocabulary"`
}
```

**Godoc examples** are **self-documenting** and show up in IDE tooltips.

---

## Moderator Summary

### Key Findings

**Good Naming:**
- ✅ Command types: `[Action]Command` (consistent)
- ✅ Constructor functions: `New[Action]Command()` (predictable)
- ✅ Function names are self-documenting
- ✅ Standard abbreviations (`cfg`, `doc`, `vocab`) are fine

**Problematic Naming:**
- 🔥 `TTMPConfig` — Unexplained acronym
- 🔥 `TicketDirectory` — Misleading (not a directory, it's metadata)
- ⚠️ Missing package documentation (`pkg/models`, `pkg/commands`)

### Consensus

**Everyone agrees:**
1. ✅ Rename `TTMPConfig` → `WorkspaceConfig`
2. ✅ Rename `TicketDirectory` → `TicketWorkspace`
3. ✅ Add package-level documentation
4. ✅ Add godoc comments with examples for exported types
5. ✅ Add glossary to README.md

**Disagreement:** None (rare unanimous agreement)

### Action Items

**High priority:**
1. ✅ Rename `TTMPConfig` → `WorkspaceConfig` (with type alias for compat)
2. ✅ Rename `TicketDirectory` → `TicketWorkspace`
3. ✅ Add package docs to `pkg/models`, `internal/*`

**Medium priority:**
4. ✅ Add godoc examples to key types
5. ✅ Add glossary to README.md
6. ✅ Update help text: "ttmp" → "workspace"

**Low priority:**
7. ❓ Consider renaming `.ttmp.yaml` → `.docmgr.yaml` (support both)
8. ❓ Change default root from "ttmp" to "workspace" (breaking change)

### Migration Strategy (Sarah's Phased Approach)

**Phase 1: Aliases**
```go
type TTMPConfig = WorkspaceConfig  // Backward compat
type TicketDirectory = TicketWorkspace
```

**Phase 2: Update docs**
- README, godoc, help text

**Phase 3: Optional breaking changes**
- Rename `.ttmp.yaml` (support both)
- Change default root (if desired)

### Connection to Other Rounds

- **Round 1**: Package names (`internal/documents`, `cmd/docmgr/cmds`) are clear
- **Round 4**: Error messages should use new names ("workspace" not "ttmp")
- **Round 9**: Documentation will use updated terminology

### Moderator's Observation

- **Casey's confusion is a real user experience issue** — Acronyms hurt onboarding
- **Quick wins available** — Renaming 2 types + adding docs is low-effort, high-impact
- **Backward compatibility is easy** — Type aliases mean no breaking changes
- **Consensus is strong** — Everyone agrees on the fixes

**Recommendation:** Implement naming changes in next refactor pass (alongside Round 1 restructuring).
