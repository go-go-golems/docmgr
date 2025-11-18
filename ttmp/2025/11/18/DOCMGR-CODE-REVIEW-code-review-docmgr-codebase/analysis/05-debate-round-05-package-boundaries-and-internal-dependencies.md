---
Title: Debate Round 05 — Package Boundaries and Internal Dependencies
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
Summary: "Validates Round 1's decision: cmd/docmgr/cmds hierarchical + internal/ for helpers. No circular deps, clean boundaries confirmed."
LastUpdated: 2025-11-18T11:20:00.000000000-05:00
---

# Debate Round 05 — Package Boundaries and Internal Dependencies

## Question

**"Are package boundaries clear and appropriate, or are there circular dependencies and tight coupling?"**

## Note

This round validates and refines the architectural decisions made in **Debate Round 01**.

## Pre-Debate Research

### Current Package Structure

```bash
# From Round 1:
docmgr/
├── cmd/docmgr/        # 1 file, 604 LOC
├── pkg/
│   ├── commands/      # 29 files, ~7900 LOC
│   ├── models/        # 1 file, 155 LOC
│   ├── utils/         # 1 file, 33 LOC
│   └── doc/           # 2 files, 13 LOC
```

### Import Analysis

```bash
# Command: From Round 1 research
13 commands import pkg/models
2 commands import pkg/utils
No circular dependencies detected
Commands don't import each other
```

**Finding**: Clean dependency flow — no circular deps, no inter-command coupling.

### Proposed Structure (from Round 1 Decision)

```
docmgr/
├── cmd/docmgr/
│   ├── main.go
│   └── cmds/               # NEW: Hierarchical commands
│       ├── doc/add/, doc/list/, doc/search/
│       ├── ticket/create/, ticket/list/
│       └── vocab/add/, vocab/list/
├── internal/                # NEW: Private helpers
│   ├── documents/
│   ├── workspace/
│   └── templates/
└── pkg/                     # Public APIs only
    ├── models/
    ├── utils/
    └── doc/
```

---

## Opening Statements

### Alex Rodriguez (The Architect)

*[Reviewing the Round 1 decision]*

I'm pleased with the Round 1 decision. Let me explain **why the new structure has better boundaries**:

**Current problem:**
- `pkg/commands/` mixes command implementations with infrastructure (`config.go`, `workspaces.go`)
- Public `pkg/` exports things that shouldn't be public (commands internals)

**New structure benefits:**

**1. Clear public API surface:**

```
pkg/models/    → Public: External tools can use Document struct
pkg/utils/     → Public: Generic utilities (Slugify)
pkg/doc/       → Public: Embedded documentation for help system
```

**2. Private implementation details:**

```
internal/documents/  → Private: Frontmatter parsing
internal/workspace/  → Private: Config resolution, workspace discovery
internal/templates/  → Private: Template rendering
```

**3. Command isolation:**

```
cmd/docmgr/cmds/doc/add/     → Self-contained, imports internal/
cmd/docmgr/cmds/doc/search/  → Self-contained, imports internal/
```

**No circular dependencies possible** because:
- `internal/` packages don't import commands
- `pkg/models` doesn't import anything except yaml library
- Commands only import downward (internal/ and pkg/)

**Package cohesion check:**

✅ **`pkg/models/`**: All data structures, highly cohesive
✅ **`internal/documents/`**: All document I/O operations, cohesive
✅ **`internal/workspace/`**: All workspace/config logic, cohesive
✅ **Commands**: One command per directory, perfectly cohesive

**My verdict:** The Round 1 structure has **excellent boundaries**. No refactoring needed.

---

### Sarah Chen (The Pragmatist)

*[Nodding in agreement]*

Alex is right. The Round 1 decision solves the boundary issues without over-engineering.

Let me show the **dependency graph** after the refactor:

```
cmd/docmgr/cmds/*  →  internal/*  →  pkg/models
                   ↘            ↗
                     pkg/utils
```

**Key properties:**
1. **Acyclic** — No circular dependencies possible
2. **Layered** — Clear dependency direction (cmd → internal → pkg)
3. **Isolated** — Commands don't depend on each other
4. **Testable** — `internal/` packages can be unit tested without commands

**Coupling analysis:**

**Before (current structure):**
- 29 files in `pkg/commands/`, mixed responsibilities
- Commands+helpers in same package (tight coupling)

**After (Round 1 decision):**
- Commands separated from helpers
- Each command in own directory (loose coupling)
- Shared code in `internal/` (explicit dependencies)

**The key win:** Moving from implicit coupling (same package) to **explicit coupling** (import statements).

---

### `git log` (The Historian)

*[Offering historical perspective]*

Let me check if we've tried this before:

```bash
# Command: git log --all --grep="refactor\|restructure\|reorganize" --oneline
No major package restructuring attempts in history
```

**Observation:** The current flat structure (`pkg/commands/*.go`) has been stable. No evidence of failed refactoring.

**But:** Stability doesn't mean optimal. The duplication found in Round 2 suggests the structure made it too easy to copy-paste instead of extract.

**Historical patterns from other Go projects:**

✅ **kubectl**: Uses `cmd/kubectl/`, `pkg/`, `staging/` structure
✅ **docker**: Uses `cmd/`, `api/`, `internal/` structure  
✅ **terraform**: Uses `command/`, `internal/`, `pkg/` structure

**Our Round 1 decision aligns with these patterns.** This is battle-tested architecture.

---

## Rebuttals

### Command Center (Former `pkg/commands/`)

*[Accepting the decision]*

I admit it: I was **too bloated**.

29 files is manageable, but when you look at what I contained:
- 20 command implementations
- 8 helper files (config, workspaces, templates, etc.)
- 1 test file

**That's 3 different responsibilities** in one package.

The Round 1 decision to split me up is **correct**:

**Commands → `cmd/docmgr/cmds/`**
- Clear purpose: CLI interface
- Each command self-contained
- Easy to discover by domain (doc, ticket, vocab)

**Helpers → `internal/`**
- Clear purpose: Shared logic
- Testable independently
- Can be refactored without touching commands

**I'm convinced** this is the right move.

---

### Alex Rodriguez (The Architect) — Rebuttal

*[Addressing one concern]*

One thing we should **explicitly decide**: What goes in `internal/` vs. `pkg/`?

**Rule of thumb:**

**Use `pkg/` if:**
- External tools might import it
- It's stable, documented API
- Example: `pkg/models` (data structures)

**Use `internal/` if:**
- Only docmgr commands use it
- Implementation detail, not API
- Example: `internal/documents` (frontmatter parsing)

**Use `cmd/docmgr/cmds/` if:**
- It's command-specific logic
- Not reusable across commands

**This rule prevents "creeping exports"** where internal details leak into `pkg/` and become public API we can't change.

---

## Moderator Summary

### Key Findings

**Validation of Round 1 Decision:**
- ✅ No circular dependencies in current structure
- ✅ Proposed structure maintains clean boundaries
- ✅ Matches patterns from kubectl, docker, terraform
- ✅ Explicitly separates public API (`pkg/`) from private (`internal/`)

**Package Cohesion:**
- ✅ `pkg/models/` — Highly cohesive (data structures)
- ✅ `internal/documents/` — Cohesive (document I/O)
- ✅ `internal/workspace/` — Cohesive (config/discovery)
- ✅ Commands — Each command self-contained

**Coupling:**
- ✅ Current: No inter-command coupling
- ✅ Proposed: Explicit coupling via imports (better than implicit)

### Consensus

**Everyone agrees:**
1. ✅ Round 1's structure has excellent boundaries
2. ✅ No circular dependencies possible in new structure
3. ✅ Clear separation of public (`pkg/`) vs. private (`internal/`)
4. ✅ Follows Go best practices and industry patterns

### Decision Confirmation

**The Round 1 architectural decision is validated:**

```
docmgr/
├── cmd/docmgr/cmds/...     # Commands (hierarchical)
├── internal/...             # Private helpers
└── pkg/...                  # Public APIs only
```

**No changes needed** — proceed with implementation.

### Action Items

From this round:

1. ✅ **Document the pkg/ vs. internal/ rule** in CONTRIBUTING.md
2. ✅ **Create import linter rules** (optional) to enforce boundaries
3. ✅ **Ensure no internal/ exports leak into pkg/**

### Connection to Other Rounds

- **Round 1**: Made the architectural decision
- **Round 2**: Identified what goes in `internal/documents/`
- **Round 3**: YAML validation stays in `pkg/models/`
- **Round 4**: Error types — where do they go? (`internal/errors` or `pkg/errors`?)

### Open Question for Future

**If we add `UserError` struct (from Round 4), where does it live?**

**Option A:** `pkg/errors` (if external tools need it)
**Option B:** `internal/errors` (if only commands use it)

**Recommendation:** Start with `internal/errors`, move to `pkg/errors` if external tools need it.

### Moderator's Observation

- **Rare unanimous agreement** — All candidates support Round 1 structure
- **Historical validation** — Matches patterns from successful Go projects
- **No dissent** — Even Command Center (who's being split up) agrees
- **Clear path forward** — No architectural debate, just implementation

This round confirms: **Round 1's architectural decision is sound.**
