---
Title: 'Jury Deliberation: Contestant #3 Norma the Navigator (Path Normalization)'
Ticket: REFACTOR-TICKET-REPOSITORY-HANDLING
Status: active
Topics:
    - refactor
    - tickets
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/paths/resolver.go
      Note: |-
        Core normalization + fuzzy matching (Representations/Suffixes/MatchPaths/DirectoryMatch)
        Core normalization + fuzzy matching logic under debate
    - Path: internal/workspace/index_builder.go
      Note: Ingestion path that calls normalization and writes related_files rows
    - Path: internal/workspace/normalization.go
      Note: |-
        Workspace normalization wrapper used when persisting RelatedFiles
        Persisted normalization envelope
    - Path: test-scenarios/testing-doc-manager/14-path-normalization.sh
      Note: |-
        End-to-end CLI proof of path matching across multiple representations
        Integration evidence script
    - Path: test-scenarios/testing-doc-manager/run-all.sh
      Note: |-
        Scenario harness used to execute the show (creates mock workspace)
        Harness used to execute scenario
    - Path: ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/reference/21-how-to-judge-contestant-3-norma-the-navigator-path-normalization.md
      Note: How-to-judge guide used as judging rubric
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-12T20:54:47.548140722-05:00
---


# Jury Deliberation: Contestant #3 Norma the Navigator (Path Normalization)

## Goal

Judge Contestant #3 (“Norma the Navigator”) on **path normalization correctness** and **safe fuzzy matching**, grounded in real scenario evidence.

## Context

Norma is responsible for making different path spellings “mean the same thing” across:

- doc-relative (`../../../../../backend/chat/api/register.go`)
- docs-root-relative (`../backend/chat/api/register.go` from `ttmp/`)
- absolute (`/tmp/docmgr-scenario-norma/acme-chat-app/backend/chat/ws/manager.go`)
- basename-only (`register.go`)

This is critical for reverse lookup (`doc search --file ...`) and directory scoping (`--dir ...`) to work even when users provide different path forms than what was stored.

Judges: Murphy (robustness), Ockham (simplicity), Oracle (spec), Ada (craft).

## Quick Reference

### Evidence (what we ran)

1) Unit-stage sanity (normalization envelope asserted indirectly via Ingrid’s ingestion test):

```
go test ./internal/workspace -run TestWorkspaceInitIndex_IngestsDocsTopicsAndRelatedFiles -count=1 -v
```

2) Integration-stage show (includes scenario #14 path normalization):

```
go build -o /tmp/docmgr-local ./cmd/docmgr
DOCMGR_PATH=/tmp/docmgr-local bash test-scenarios/testing-doc-manager/run-all.sh /tmp/docmgr-scenario-norma
```

Observed excerpt (path-normalization searches; shortened):

```
... file=backend/chat/api/register.go note=Doc-relative path reference
... file=../../../../../backend/chat/api/register.go, ../backend/chat/api/register.go note=Doc-relative path reference (deep traversal) | Ttmp-relative path reference (shallower traversal)
... file=backend/chat/ws/manager.go note=Absolute path reference (host-specific)
```

### Key implementation moves (what judges care about)

- `Resolver.Normalize` tries a sequence of anchors: repo → doc → config → docs-root → docs-parent, returning the first *existing* path, else a deterministic fallback.
- `NormalizedPath.Representations()` returns a deduped set of strings used for matching.
- `MatchPaths` attempts:
  1) exact intersection of representation sets,
  2) suffix intersection (up to 3 segments),
  3) substring fallback (most controversial).

## Usage Examples

How to reproduce exactly:

```
cd /home/manuel/workspaces/2025-12-11/improve-yaml-frontmatter-handling-docmgr/docmgr
go build -o /tmp/docmgr-local ./cmd/docmgr
DOCMGR_PATH=/tmp/docmgr-local bash test-scenarios/testing-doc-manager/run-all.sh /tmp/docmgr-scenario-norma
```

Expected result: scenario completes successfully and the `doc search --file` calls in `14-path-normalization.sh` return the expected ticket doc for all of:

- doc-relative input
- docs-root-relative input
- absolute input
- basename-only input

## Related

- Judge rubric: `reference/21-how-to-judge-contestant-3-norma-the-navigator-path-normalization.md`
- Candidate roster: `reference/16-talent-show-candidates-code-performance-review.md`

---

## Deliberation Transcript (each judge speaks 3 times)

### Round 1 — First impressions (Turn 1 each)

#### 🔨 Murphy (Turn 1)

My nightmare is false positives: we start “matching” the wrong file because a string happens to contain another string. So I’m immediately suspicious of the `containsSubstring` fallback in `MatchPaths`.

That said, the scenario evidence is compelling: we successfully find the same doc by doc-relative, docs-root-relative, absolute, and even basename-only. That’s the feature users want.

I’m going to hunt for guardrails: do we at least normalize to lower-case and slash? Yes: `normalizeForCompare` lowercases and `filepath.ToSlash`es.

#### 🗡️ Ockham (Turn 1)

This is a nice “one place decides” implementation. I like that `Resolver.Normalize` has a single, explicit base order. That makes it predictable.

The interface is simple: call `Normalize`, get `Canonical`, `RepoRelative`, `DocsRelative`, `DocRelative`, `Abs`, `Anchor`, `Exists`.

The danger with normalization code is it becomes a magical black box. This one is “readable magic”: a fixed list of anchors and a deterministic fallback.

#### 📜 Oracle (Turn 1)

Oracle cares about spec intent: store multiple normalized keys so SQL matching later doesn’t need to reconstruct anchors.

We see that explicitly in `internal/workspace/normalization.go`: it persists an envelope of representations (norm_canonical, norm_repo_rel, norm_docs_rel, norm_doc_rel, norm_abs, norm_clean, anchor).

Scenario #14 demonstrates the user-visible guarantee: “search works regardless of how paths were spelled.”

#### 💎 Ada (Turn 1)

Craft note: the `Resolver` is surprisingly clean. The anchor order and the `buildResult` fields align with what we store in sqlite. That’s coherence.

I also like the “exists-first, fallback-second” approach: choose the first anchor that points to a real file. If nothing exists, still return something deterministic so the system behaves predictably.

The one scary bit is substring matching. It’s hard to reason about long-term, so we’ll discuss that in Round 2.

### Round 2 — Cross-examination (Turn 2 each)

#### 🔨 Murphy (Turn 2)

Let’s talk substring fallback. If a user searches for `register.go`, substring matching makes that succeed even when we don’t have an exact representation intersection.

But substring matching also risks: searching for `api.go` might match `graphqlapi.go` or something dumb.

Do we have mitigation? Some: we try exact representations and suffixes first; substring is last resort. Also `Suffixes(3)` is a middle ground that feels safer than substring.

My preference: keep substring, but constrain it further (maybe only basename-only queries), or document it explicitly so we accept the trade-off with eyes open.

#### 🗡️ Ockham (Turn 2)

I’m with Murphy on “document the trade-off.” The simplest code that solves the real problem is acceptable, but substring matching is where “simple” can become “surprising.”

However, the scenario suite explicitly tests basename-only search; without substring fallback you might lose that convenience.

So my position: keep it, but make it very explicit in comments and tests that this is a deliberate convenience feature.

#### 📜 Oracle (Turn 2)

Oracle observes that the scenario suite enshrines the behavior: basename-only lookup is a supported capability.

Given that, the code is aligned with expected behavior, but we should ensure the intent is recorded as “best-effort matching” not “exact identity.”

I’d also like to see, long term, a spec note about when substring matching is allowed (maybe only after suffix checks fail).

#### 💎 Ada (Turn 2)

From maintainability: the anchor ordering is the “core contract” here. It should have a small “WHY this order” comment:

- repo-root is the most stable canonical,
- doc-relative is needed for docs that refer to nearby files,
- config/docs-root/docs-parent are fallbacks that make tooling resilient.

Also, substring matching should have a comment explaining the user-facing payoff (basename queries) and the risk (possible false positives).

### Round 3 — Final verdict (Turn 3 each)

#### 🔨 Murphy (Turn 3)

Verdict: **Ship**, with a caution label.

The feature is demonstrably useful and works end-to-end. My only concern is false positives from substring matching. I want that risk clearly documented and (ideally) narrowed later.

Score: **8.75/10**.

#### 🗡️ Ockham (Turn 3)

Verdict: **Ship**.

The implementation is compact and understandable. It solves a real pain point. Add a few “WHY” comments around anchor order and matching tiers and it’s golden.

Score: **9.0/10**.

#### 📜 Oracle (Turn 3)

Verdict: **Ship**.

The envelope strategy and the scenario evidence meet the intent: cross-form path matching works for users.

Score: **9.25/10** (request: document substring matching intent explicitly).

#### 💎 Ada (Turn 3)

Verdict: **Ship**.

This is one of those subsystems where a tiny comment prevents a future refactor from “simplifying away” critical behavior. A couple targeted comments would protect the design.

Score: **9.0/10**.

### Final aggregate + follow-ups

**Aggregate score:** 9.0/10  
**Final verdict:** ✅ **SHIP** (minor documentation + future tightening of substring matching)

**Follow-ups:**

1) Add short “WHY” comments in `Resolver.Normalize` describing anchor order rationale.  
2) Add a comment in `MatchPaths` explaining the tiered matching strategy and why substring fallback exists (basename convenience).  
3) Consider constraining substring fallback in the future (e.g., only for basename-only queries) to reduce false positives.
