---
Title: Intern code review verification questionnaire
Ticket: CLEANUP-LEGACY-WALKERS
Status: active
Topics:
    - refactor
    - tickets
    - docmgr-internals
DocType: playbook
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/workspace/query_docs.go
      Note: QueryDocs implementation (questions reference this)
    - Path: internal/workspace/skip_policy.go
      Note: Canonical ingest skip policy (question q19)
    - Path: pkg/commands/status.go
      Note: Phase 1.1 example migration (questions reference this)
    - Path: ttmp/2025/12/13/CLEANUP-LEGACY-WALKERS--cleanup-legacy-walkers-and-discovery-helpers-after-workspace-refactor/design/01-cleanup-overview-and-migration-guide.md
      Note: Design spec with migration patterns and no-compat policy
    - Path: ttmp/2025/12/13/CLEANUP-LEGACY-WALKERS--cleanup-legacy-walkers-and-discovery-helpers-after-workspace-refactor/reference/01-diary.md
      Note: Primary reading material for questionnaire
ExternalSources: []
Summary: Questionnaire and experiments for an intern reviewing the CLEANUP-LEGACY-WALKERS cleanup work
LastUpdated: 2025-12-13T12:45:00-05:00
---


# Intern Code Review Verification Questionnaire

## Purpose

This playbook verifies that an intern (or new reviewer) understands the **CLEANUP-LEGACY-WALKERS** refactor: why we migrated commands from legacy walkers (`CollectTicketWorkspaces`, `findTicketDirectory`, `filepath.Walk`) to the centralized `Workspace.QueryDocs` API, and what changed as a result.

## Prerequisites

- Read the cleanup ticket **diary** (`reference/01-diary.md`) through at least **Step 2.2**.
- Read the cleanup **design spec** (`design/01-cleanup-overview-and-migration-guide.md`).
- Have the repo checked out with Phase 1 + Phase 2.1–2.2 migrations complete.
- Be able to run `go test ./...` and build a local `docmgr` binary.

## Verification Format (YAML DSL)

The questions below are structured in YAML for machine readability (and future automation). Answer them in a separate document or in comments here.

```yaml
verification:
  ticket: CLEANUP-LEGACY-WALKERS
  reviewer: <your-name>
  date: <YYYY-MM-DD>
  
  sections:
    - id: understanding-intent
      title: "Understanding the Refactor Intent"
      questions:
        - id: q1
          type: multiple_choice
          question: "What was the PRIMARY goal of the CLEANUP-LEGACY-WALKERS refactor?"
          options:
            - "A) Make docmgr faster by caching document metadata"
            - "B) Remove duplicated discovery/walking logic and centralize semantics in Workspace.QueryDocs"
            - "C) Add new filtering capabilities to doc search"
            - "D) Fix bugs in the legacy CollectTicketWorkspaces walker"
          correct: B
          explanation: |
            The refactor's goal is **consistency and de-duplication**: before this work, each command
            implemented slightly different ticket discovery, skip rules, and frontmatter parsing.
            By migrating all commands to use the same Workspace+QueryDocs backend, we ensure
            "what is a doc" and "how filters behave" are defined in ONE place.
        
        - id: q2
          type: multiple_choice
          question: "The cleanup spec explicitly states a policy on backwards compatibility. What is it?"
          options:
            - "A) Preserve all legacy behavior using flags/shims to avoid breaking scripts"
            - "B) Document behavior changes but make no compatibility shims; QueryDocs semantics win"
            - "C) Maintain compatibility for 2 major versions before changing semantics"
            - "D) Use feature flags to toggle between old and new discovery paths"
          correct: B
          explanation: |
            The spec (design/01-cleanup-overview-and-migration-guide.md §3 "No Backwards Compatibility")
            explicitly states: **no shims, no flags, no legacy fallbacks**. When QueryDocs semantics differ
            from legacy behavior, we document the change and update tests/docs—but we don't add compatibility layers.

    - id: understanding-architecture
      title: "Understanding the New Architecture"
      questions:
        - id: q3
          type: free_form
          question: |
            Explain in 2–3 sentences: what is the role of `workspace.Workspace` and how does it relate to `QueryDocs`?
          guidance: |
            A strong answer mentions:
            - Workspace is the "front door" for document discovery and metadata queries
            - It owns: docs root discovery, config loading, path normalization (via paths.Resolver), and the in-memory SQLite index
            - QueryDocs is the primary query API exposed by Workspace; it translates structured filters into SQL
        
        - id: q4
          type: multiple_choice
          question: "When a command like `docmgr list tickets` runs, when/where is the in-memory SQLite index built?"
          options:
            - "A) Once at process startup, shared across all commands"
            - "B) Lazily on the first query, then cached for subsequent queries"
            - "C) Per CLI invocation: each command calls ws.InitIndex, rebuilding from scratch"
            - "D) The index is pre-built offline and loaded from a .docmgr/ directory"
          correct: C
          explanation: |
            The index is rebuilt from scratch per CLI invocation (Decision Q16 in the original refactor spec).
            Commands call ws.InitIndex(ctx, opts) explicitly after discovering the workspace.
            This keeps the design simple and avoids stale-cache issues at the cost of a small upfront walk+parse.
        
        - id: q5
          type: code_reading
          question: "Open `pkg/commands/status.go` and find the function `computeStatusTickets`. Why does it need to scan `QueryDocs` results twice (once for index docs, once for non-index docs)?"
          guidance: |
            The answer is: `status` needs **per-ticket metadata** (from index.md) and **per-ticket doc counts**
            (from non-index docs). We could do this in one QueryDocs call, but splitting into two queries makes
            the logic clearer: index docs give us ticket metadata + staleness, non-index docs give us counts.
        
        - id: q6
          type: free_form
          question: |
            In the legacy code, `status.go` used `CollectTicketWorkspaces` + `filepath.Walk`. What were the **semantic differences** between that approach and the new QueryDocs-based approach?
          guidance: |
            Strong answers mention:
            - **Skip rules**: legacy used ad-hoc "skip if starts with underscore", now uses canonical skip_policy.go
            - **Parse-error docs**: legacy silently skipped broken frontmatter, now broken docs are indexed with parse_ok=0 (and can be surfaced with IncludeErrors=true)
            - **Path tags**: legacy had no concept of "control doc" or "archived path" tagging, now those are explicit
    
    - id: understanding-filters
      title: "Understanding QueryDocs Filters and Semantics"
      questions:
        - id: q7
          type: multiple_choice
          question: "If you call `ws.QueryDocs(ctx, workspace.DocQuery{Scope: workspace.Scope{Kind: workspace.ScopeTicket, TicketID: \"MEN-1234\"}, Filters: workspace.DocFilters{DocType: \"design-doc\"}})`, what documents will be returned?"
          options:
            - "A) All design-docs in the entire repository"
            - "B) All design-docs under ticket MEN-1234"
            - "C) Only the ticket index.md for MEN-1234 (Scope overrides Filters)"
            - "D) Error: contradictory query (cannot use Scope and Filters together)"
          correct: B
          explanation: |
            Scope and Filters compose: ScopeTicket constrains the result set to docs under that ticket,
            then Filters further narrows by DocType. The SQL compiled query will have:
            `WHERE d.ticket_id = 'MEN-1234' AND d.doc_type = 'design-doc'`
        
        - id: q8
          type: experiment
          question: "Run this command and observe the output"
          command: |
            docmgr list tickets --ticket CLEANUP
          expected_behavior: |
            The command should return **zero results** (or just tickets containing "CLEANUP" as a substring if legacy behavior leaked back in).
            
            **Why**: After Phase 1.2, ticket filtering is **exact match** (via `ticket_id = ?`), not substring.
            The ticket ID is `CLEANUP-LEGACY-WALKERS`, so `--ticket CLEANUP` doesn't match.
          follow_up: |
            Now try: `docmgr list tickets --ticket CLEANUP-LEGACY-WALKERS`
            This should return exactly one ticket (the cleanup ticket).
        
        - id: q9
          type: code_reading
          question: "Open `internal/workspace/query_docs_sql.go` and find `compileDocQuery`. Trace how a `RelatedFile` filter becomes a SQL WHERE clause. What normalization strategy is used?"
          guidance: |
            The answer: `buildQueryPathKeySet` normalizes the user-provided path using the Workspace's resolver,
            yielding multiple representations (canonical, repo-relative, docs-relative, doc-relative, absolute, clean).
            The SQL then does `EXISTS (SELECT 1 FROM related_files WHERE ... AND (norm_canonical IN (...) OR norm_repo_rel IN (...) ...))`.
            This is the "fallback matching" strategy: we match against ANY stored representation.
    
    - id: understanding-tricky-parts
      title: "Understanding the Tricky Parts"
      instructions: |
        These questions focus on the "sharp edges" documented in the diary. Read the diary steps for context.
      
      questions:
        - id: q10
          type: free_form
          question: |
            Phase 1.1 (status.go migration) had to distinguish between "ticket index docs" and "non-index docs" to compute per-ticket counts. How did we identify "index docs" in the QueryDocs results?
          guidance: |
            Answer: we check `filepath.Base(h.Path) == "index.md"` OR `h.Doc.DocType == "index"`.
            This is documented in the diary Step 1 "What was tricky" section.
        
        - id: q11
          type: experiment
          question: "Verify parse-error doc handling"
          setup: |
            1. Create a markdown file with broken YAML frontmatter:
               ```bash
               echo -e "---\nBroken: [\n---\n\n# Test" > /tmp/broken-doc.md
               ```
            2. Copy it into a ticket dir:
               ```bash
               TICKET_DIR=$(docmgr list tickets --ticket CLEANUP-LEGACY-WALKERS --with-glaze-output --select path --output json | jq -r '.[0].path')
               cp /tmp/broken-doc.md "$TICKET_DIR/broken-test.md"
               ```
            3. Query with IncludeErrors=false (the default):
               ```bash
               docmgr list docs --ticket CLEANUP-LEGACY-WALKERS
               ```
            4. Query with diagnostics enabled (simulated by doctor):
               ```bash
               docmgr doctor --ticket CLEANUP-LEGACY-WALKERS --fail-on none | grep -i broken
               ```
          expected_behavior: |
            - `list docs` should NOT show the broken doc (parse_ok=0 docs are excluded by default).
            - `doctor` SHOULD report it as a finding (it uses IncludeErrors=true + IncludeDiagnostics=true).
          cleanup: |
            ```bash
            rm "$TICKET_DIR/broken-test.md"
            ```
        
        - id: q12
          type: multiple_choice
          question: "Why did we need to update the scenario harness (run-all.sh) to require pinned DOCMGR_PATH?"
          options:
            - "A) The refactor changed the CLI flag names and an old binary would fail"
            - "B) The system-installed docmgr was older and didn't support new verbs like workspace export-sqlite"
            - "C) QueryDocs requires a newer SQLite version not available on the system"
            - "D) go run ./cmd/docmgr doesn't work in CI environments"
          correct: B
          explanation: |
            The original failure (documented in analysis/01-phase-1-integration-suite-failure-analysis-wrong-docmgr-binary.md)
            was: the scenario ran the **system** docmgr (older) which didn't recognize `--out` flag.
            By requiring `DOCMGR_PATH` to be explicitly set, we force test runs to pin the binary under test.
    
    - id: understanding-migration-patterns
      title: "Understanding Migration Patterns"
      instructions: |
        These questions test whether you can apply the same migration patterns to future commands.
      
      questions:
        - id: q13
          type: code_reading
          question: "Open `pkg/commands/add.go` (Phase 2.1 migration). Find `findTicketDirectoryViaWorkspace`. Explain in 1–2 sentences: what does this helper replace, and why was it kept as a separate function?"
          guidance: |
            Answer: it replaces the legacy `findTicketDirectory` call for ticket discovery.
            It was kept as a helper because `add.go` only needs the **ticket directory path**
            (to mkdir subdirs and write files), not the full QueryDocs hydration.
        
        - id: q14
          type: experiment
          question: "Test exact-match ticket filtering"
          command: |
            # Try substring match (should NOT work anymore after Phase 1.2):
            docmgr list tickets --ticket CLEANUP
            
            # Try exact match (should work):
            docmgr list tickets --ticket CLEANUP-LEGACY-WALKERS
          expected_behavior: |
            - First command returns **zero results** (ticket ID is `CLEANUP-LEGACY-WALKERS`, not `CLEANUP`).
            - Second command returns **one ticket** (exact match).
            
            This demonstrates that ticket filtering is now **exact match via `ticket_id = ?`**, not substring.
        
        - id: q15
          type: free_form
          question: |
            Look at `pkg/commands/meta_update.go` (Phase 2.2). When `--ticket` and `--doc-type` are both provided, which QueryDocs parameters are used to enumerate target docs?
          guidance: |
            Answer: `ScopeTicket` with `TicketID=<ticket>`, and `Filters.DocType=<doc-type>`.
            This returns all docs matching that type under the ticket, which meta_update then iterates to apply the field update.
    
    - id: testing-and-validation
      title: "Testing and Validation"
      instructions: |
        These exercises verify you can validate the migrations yourself.
      
      questions:
        - id: q16
          type: experiment
          question: "Build and test a local binary"
          commands:
            - step: "Build the refactored binary"
              command: "go build -o /tmp/docmgr-cleanup-verify ./cmd/docmgr"
            - step: "Run integration suite with pinned binary"
              command: "DOCMGR_PATH=/tmp/docmgr-cleanup-verify bash test-scenarios/testing-doc-manager/run-all.sh /tmp/docmgr-verify-scenario"
            - step: "Verify completion"
              command: "echo $?"
          expected_behavior: |
            - The suite should complete successfully: `[ok] Scenario completed at .../acme-chat-app`
            - Exit code should be 0
            - No "unknown flag" or "command not found" errors
        
        - id: q17
          type: free_form
          question: |
            Suppose you're reviewing Phase 3 (which migrates write-path commands like `doc_move.go`, `ticket_close.go`).
            What part of those commands should migrate to QueryDocs, and what part should remain filesystem-based?
          guidance: |
            Answer: **Discovery** (finding the ticket dir, resolving the target doc) should migrate to QueryDocs.
            **Write operations** (moving files, rewriting frontmatter in bulk) should stay filesystem-based
            because they're actually modifying disk, not querying metadata.
            
            This pattern is documented in the cleanup migration guide §4 "Risk Notes" → "Write-Path Commands".
    
    - id: regression-awareness
      title: "Regression Awareness"
      instructions: |
        These questions test whether you'd catch a future regression or semantic drift.
      
      questions:
        - id: q18
          type: multiple_choice
          question: "A developer adds a new command `docmgr doc archive --ticket X` and implements ticket discovery using `filepath.Walk`. What should you ask them to do instead?"
          options:
            - "A) Nothing; filepath.Walk is fine for write-path commands"
            - "B) Use Workspace.DiscoverWorkspace + ws.QueryDocs for ticket discovery instead"
            - "C) Use CollectTicketWorkspaces instead of filepath.Walk"
            - "D) Add a --legacy-discovery flag so both paths are supported"
          correct: B
          explanation: |
            Post-cleanup, ALL ticket discovery should go through Workspace.QueryDocs.
            CollectTicketWorkspaces is deprecated (Phase 4), and adding a --legacy-discovery flag
            violates the "no backwards compatibility" policy.
        
        - id: q19
          type: code_reading
          question: "Open `internal/workspace/skip_policy.go` and read `DefaultIngestSkipDir`. Which directories are ALWAYS skipped during index ingestion?"
          expected_answer: |
            - `.meta/` (always)
            - Directories starting with underscore: `_templates/`, `_guidelines/`, etc. (always)
          follow_up: |
            If a command still uses `filepath.Walk` without calling `DefaultIngestSkipDir`, what's the risk?
            (Answer: that command will see docs that the index doesn't know about, causing "found in X but not in Y" mismatches.)
        
        - id: q20
          type: experiment
          question: "Verify that broken frontmatter docs are indexed but hidden by default"
          setup: |
            1. Create a temp ticket and a doc with invalid frontmatter:
               ```bash
               docmgr ticket create-ticket --ticket INTERN-TEST --title "Intern verification test"
               TICKET_DIR=$(docmgr list tickets --ticket INTERN-TEST --with-glaze-output --select path --output json | jq -r '.[0].path')
               echo -e "---\nBroken: [\n---\n\n# Test" > "$TICKET_DIR/broken.md"
               ```
            2. Query without IncludeErrors:
               ```bash
               docmgr list docs --ticket INTERN-TEST | grep -i broken
               ```
            3. Query via doctor (which uses IncludeErrors=true):
               ```bash
               docmgr doctor --ticket INTERN-TEST --fail-on none | grep -i broken
               ```
          expected_behavior: |
            - `list docs` should NOT show broken.md
            - `doctor` SHOULD report it as a finding (parse error)
          cleanup: |
            ```bash
            rm -rf "$TICKET_DIR"
            ```
          explanation: |
            This demonstrates the Spec §10.6 contract: broken docs are indexed (`parse_ok=0`), but excluded
            from default results unless `IncludeErrors=true`. This lets repair workflows (like doctor) surface
            them without polluting normal queries.

    - id: future-work-understanding
      title: "Understanding Future Work"
      instructions: |
        These questions test whether you understand what should happen next (based on diary "future work" sections).
      
      questions:
        - id: q21
          type: free_form
          question: |
            The diary Step 2 mentions: "If users observe that `--ticket X` no longer works because they were relying on substring matching, what should we do?"
          expected_answer: |
            **Do not add a compatibility shim.** Instead:
            - Document the exact-match semantics clearly in help text and docs
            - Add tests that codify the new contract
            - If substring matching is genuinely needed, treat it as a new feature request with a dedicated flag (e.g. `--ticket-contains`) that has explicit semantics
        
        - id: q22
          type: code_reading
          question: "Open `internal/workspace/index_builder.go` and find the `inferTicketIDFromPath` function. Why does it exist, and when is it called?"
          expected_answer: |
            It's called when a doc's **frontmatter fails to parse** (`readErr != nil` or `doc == nil`).
            It tries to extract a ticket ID from the directory structure (ttmp/YYYY/MM/DD/<TICKET--slug>/...)
            so that broken docs can still be discovered by ticket-scoped queries (useful for doctor/repair flows).
          follow_up: |
            What happens if a broken doc lives OUTSIDE the standard ticket layout?
            (Answer: ticket_id will be NULL/empty, so ScopeTicket queries won't find it unless we query ScopeRepo.)
```

## Exit Criteria

The intern/reviewer should be able to:

- **Explain** the refactor's intent (de-duplication + centralized semantics).
- **Trace** a QueryDocs call from a command through SQL compilation to result hydration.
- **Identify** when a new command should use QueryDocs vs when filesystem ops are appropriate.
- **Run** the integration suite successfully with a pinned local binary.
- **Spot** future regressions (commands bypassing Workspace, compatibility shims, duplicated skip rules).

## Related Resources

- Diary: `reference/01-diary.md`
- Design spec: `design/01-cleanup-overview-and-migration-guide.md`
- Inventory: `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/analysis/09-cleanup-inventory-report-task-18.md`
- Original refactor spec: `ttmp/2025/12/12/REFACTOR-TICKET-REPOSITORY-HANDLING--refactor-ticket-repository-handling/design/01-workspace-sqlite-repository-api-design-spec.md`
- Scenario harness: `test-scenarios/testing-doc-manager/run-all.sh`

---

**Use this playbook to onboard new reviewers or to self-verify understanding before continuing Phase 3.**
