---
Title: 'Debate Round 3 — Q8: How do we keep vocabulary/config concerns from leaking everywhere?'
Ticket: REFACTOR-TICKET-REPOSITORY-HANDLING
Status: active
Topics:
    - refactor
    - tickets
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-12T15:04:32.744313555-05:00
---

# Debate Round 3 — Q8: How do we keep vocabulary/config concerns from leaking everywhere?

## Goal

Debate **Question 8** from `reference/02-debate-questions-repository-lookup-api-design.md`:

> How do we keep vocabulary/config concerns from leaking everywhere?

We want a clean boundary between:

- “repository lookup” (find tickets, find docs, normalize paths)
- “validation/policy” (vocabulary checks, required fields, ignore rules)
- “UI/formatting” (human output, templates, schema printing)

## Context

Today, many commands resolve:

- docs root (`workspace.ResolveRoot`)
- config path (`workspace.FindTTMPConfigPath`)
- vocabulary path (`workspace.ResolveVocabularyPath`)

But only some commands actually *use* vocabulary/config semantically (e.g. doctor validation). Others just print the resolved paths in output.

This round is about **semantics + structure**, not performance/security/backcompat.

## Pre-Debate Research (codebase evidence)

### Evidence A — Vocabulary is loaded in the commands layer

- `pkg/commands/vocabulary.go`:
  - `LoadVocabulary()` calls `workspace.ResolveVocabularyPath()` and parses YAML into `models.Vocabulary`.
  - If not found, returns an empty vocabulary (not an error).

### Evidence B — Doctor uses vocabulary deeply (validation policy), not lookup

- `pkg/commands/doctor.go`:
  - loads `.docmgrignore` patterns (repo root + docs root)
  - calls `LoadVocabulary()` and builds sets for topics/docTypes/intent/status
  - uses those sets to produce warnings like `unknown_topics`, `unknown_status`, etc.

Interpretation: doctor is a **policy/validation** layer that consumes docs/tickets; it shouldn’t define repository lookup semantics.

### Evidence C — Many commands resolve vocab/config but don’t use them for behavior

- `pkg/commands/add.go` and `pkg/commands/meta_update.go` both resolve `cfgPath` and `vocabPath`, primarily for display/context.
- `add` pulls defaults from the ticket index doc (topics/owners/status/intent), not from vocabulary.
- `meta update` updates fields without vocabulary validation.

### Evidence D — Root/config resolution is already centralized, but scattered at call sites

Most commands do:
- `settings.Root = workspace.ResolveRoot(settings.Root)`
- then optionally `FindTTMPConfigPath()` to compute `configDir` (e.g. `search`, `relate`)

Interpretation: the “context object” exists conceptually but not as a single reusable value.

## Debate (Question 8)

### Opening Statements (Round 1)

#### Mara (Staff Engineer) — “Separate lookup from policy; make policy opt-in”

I want repository lookup to be vocabulary-agnostic. Lookup should return docs/tickets (and parse errors) and maybe provide the resolved context (root/configDir/repoRoot), but it should not decide what topics/status are “valid”.

Validation belongs in separate components:
- `Validator` (vocabulary-aware)
- `Doctor` (uses validators + emits diagnostics)

The repository API should expose stable primitives; doctor builds policies on top.

#### Jon (Senior Engineer) — “Centralize context, not rules”

We can reduce leakage by creating one object that owns resolved paths:

- `RepoContext{Root, ConfigPath, ConfigDir, RepoRoot, VocabularyPath}`

Then commands can depend on it without each re-resolving things. But I agree that vocabulary rules belong outside the lookup API.

#### `workspace.ResolveRoot` — “I’m configuration discovery; don’t duplicate me”

I already unify root/config discovery. The new repository API should accept a `RepoContext` (or build one) so commands don’t call me + `FindTTMPConfigPath` + `ResolveVocabularyPath` repeatedly.

#### `workspace.CollectTicketWorkspaces` — “Give me ignore rules, but keep vocab out”

Ticket discovery might need ignore rules (underscore dirs, `.docmgrignore`, command flags). That’s orthogonal to vocabulary. Keep vocabulary checks out of discovery; expose enough info so policy layers can decide what to do.

#### `documents.WalkDocuments` — “Parsing is neutral; validation is layered”

I’m a neutral walker: I return parse errors, I don’t judge. The repository layer should keep this neutrality. If a caller wants “only valid docs”, that’s a filter on top.

#### `paths.Resolver` — “ConfigDir and RepoRoot belong in context, not scattered”

Path normalization needs anchors (docs root, config dir, repo root). That’s configuration, but it’s part of lookup correctness. So the repository layer should own constructing resolvers consistently. That’s not “vocabulary leakage”; it’s “path context”.

#### `pkg/commands/*` (bloc) — “We want fewer knobs; keep validation separate”

Commands want a single way to get:
- root + configDir
- doc/ticket handles
- consistent path normalization

But we don’t want every command to become vocabulary-aware. Keep vocabulary checks in doctor (and maybe in `validate frontmatter`), not in base lookup.

### Rebuttals (Round 2)

#### Mara → `paths.Resolver`

Agree that configDir/repoRoot are needed for normalization. That’s “lookup context” and belongs in the repository object. Vocabulary is different: it’s policy, not context.

#### Jon → Mara

Let’s make the boundary concrete: “lookup produces handles + context; policy consumes handles + context.” That’s easy to explain and implement incrementally.

### Moderator Summary

**Consensus direction**
- Split into at least two layers:
  - **Lookup layer**: root/config context, ticket discovery, doc enumeration, path normalization.
  - **Policy layer**: vocabulary validation, required-field validation, ignore-rule selection, diagnostics output.

**Key design artifact to draft next**
- A `RepoContext` struct and which fields belong there.

**Immediate next step suggestion**
- Pick one command that currently mixes concerns (`doctor`) and one that mostly does lookup (`search`), and outline how each would consume the new layers.

## Usage Examples

If we run a follow-up round, we can draft a “layering diagram” and then test it mentally against `doctor`, `search`, `relate`, and `meta update`.

## Related

- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/02-debate-questions-repository-lookup-api-design.md`
- `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/analysis/01-ticket-discovery-document-lookup-codebase-analysis.md`
