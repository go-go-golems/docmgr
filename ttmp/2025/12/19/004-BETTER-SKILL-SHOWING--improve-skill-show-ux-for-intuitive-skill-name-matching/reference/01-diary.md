---
Title: Diary
Ticket: 004-BETTER-SKILL-SHOWING
Status: active
Topics:
  - skills
  - ux
  - cli
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
  - Path: ttmp/2025/12/19/004-BETTER-SKILL-SHOWING--improve-skill-show-ux-for-intuitive-skill-name-matching/reference/01-bug-report-skill-show-name-matching-issues.md
    Note: Primary bug report describing failures + desired UX
  - Path: pkg/commands/skill_show.go
    Note: Current `skill show` implementation to be made more resilient
  - Path: pkg/commands/skill_list.go
    Note: Current `skill list` output to be adjusted to show load commands
  - Path: test-scenarios/testing-doc-manager/20-skills-smoke.sh
    Note: Scenario tests to extend (clashes, --ticket, path/slug matching)
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-19T00:00:00Z
---

# Diary

## Goal

Improve `docmgr skill show` UX to be resilient to “LLM-ish” inputs (title variations, slug/filename, and paths), and improve `docmgr skill list` UX to print copy/pasteable load commands. This diary records the investigation, decisions, implementation, and testing outcomes.

## Step 1: Trace current behavior and identify the real root cause

This step focused on understanding why `docmgr skill show <name>` fails with “Too many arguments”, and where to implement resilient matching. The key outcome was confirming that argument parsing is controlled by Glazed command descriptions: since `skill show` declares **no positional arguments**, any argument triggers the Glazed “Too many arguments” error.

### What I did
- Read the ticket bug report capturing the UX failures and recommended matching strategies.
- Traced command wiring:
  - `cmd/docmgr/cmds/skill/*.go` → `common.BuildCommand(...)`
  - `common.BuildCommand` → `glazed/pkg/cli.BuildCobraCommand(...)`
- Located Glazed argument parsing implementation (`GatherArguments`) and confirmed where “Too many arguments” comes from.

### Why
- We need to add positional argument support in a way that works *with* Glazed, not around it.
- We need to implement matching improvements at the right layer (the command implementation), keeping cobra wiring minimal.

### What worked
- Confirmed the exact failure mode and its source (Glazed argument parsing with no declared arguments).
- Identified the exact files to change for:
  - positional argument support (`cmds.WithArguments(...)`)
  - matching logic (in `pkg/commands/skill_show.go`)
  - output UX (in `pkg/commands/skill_list.go`)
  - scenario tests (`test-scenarios/testing-doc-manager/20-skills-smoke.sh`)

### What didn't work
- N/A (this was a research-only step).

### What I learned
- In this repo, cobra positional arg validation is generated from Glazed “argument parameter definitions”.
- `cmds.WithArguments(...)` exists and sets `IsArgument = true`, which enables cobra arg acceptance and parsing.

### What was tricky to build
- The “either flag or positional” requirement cannot be expressed declaratively in Glazed, so it must be enforced in the command’s `Run()` by checking parsed settings after parsing.

### What warrants a second pair of eyes
- The final match-scoring/precedence rules in `skill show`: subtle ordering decisions can change which skill is selected when multiple candidates exist.

### What should be done in the future
- N/A (future work will be captured in later steps as implementation lands).

### Code review instructions
- Start with `ttmp/.../design/01-skill-show-and-list-ux-analysis.md` for the architecture + decision map.
- Then review changes in `pkg/commands/skill_show.go` (matching + ambiguity UX) and `pkg/commands/skill_list.go` (load command output).

### Technical details
- The “Too many arguments” error is raised by Glazed’s `ParameterDefinitions.GatherArguments(...)` when no arguments are declared but args are provided.

## Step 2: Implement resilient skill resolution + improved list UX + keep smoke output short

This step implemented the feature work described in the bug report: `docmgr skill show` now accepts a positional query and resolves skills using multiple strategies (title/prefix/slug/path), and `docmgr skill list` now prints copy/pasteable load commands rather than raw paths. We also adjusted the skills smoke test to keep output short while still validating the new behavior (including clashing skills and `--ticket`).

### What I did
- Updated `skill show`:
  - Added positional argument support (`docmgr skill show <query>`) while keeping `--skill` as a legacy flag.
  - Added `--ticket` scoping to disambiguate clashes.
  - Implemented resilient matching across:
    - title (with/without `Skill:` prefix),
    - filename slug (stripping `.md` and numeric prefixes),
    - explicit path / directory matching,
    - contains-based fallbacks.
  - If ambiguous, print “Load:” commands for each candidate instead of picking arbitrarily.
- Updated `skill list`:
  - Human output prints `Load:` commands, using shortest-unambiguous identifier order:
    - filename slug → title (no `Skill:`) → full path (if duplicates).
  - Structured output adds `load_command` column.
  - Omit `--root` in the load command when it matches the default resolved root.
- Updated scenario test `test-scenarios/testing-doc-manager/20-skills-smoke.sh`:
  - Added a clashing workspace-level `api-design.md` skill to force ambiguity.
  - Added checks for positional show, `--ticket` show, slug/path show.
  - Reduced output: print only `[ok] Test N` lines and dump command output only on failure.
- Fixed an over-eager path scoring edge case that made generic words (like `websocket`) tie due to directory-name substring matches; path scoring is now gated behind “query looks like a path”.

### Why
- LLMs (and humans) will naturally try different identifiers (slug, title without prefix, or a file path); the command should be resilient to those fluctuations.
- When multiple skills match, selecting “first match” is dangerous; better to make the ambiguity explicit with copy/pasteable commands.
- Skill list should be action-oriented: show users exactly how to load a skill.
- Smoke tests should validate behavior without drowning the user in output.

### What worked
- `docmgr skill show <query>` works for slugs (e.g. `systematic-debugging`), titles (with/without `Skill:`), and explicit paths.
- `docmgr skill show "API Design"` now correctly reports ambiguity when two skills exist, and `--ticket MEN-4242` disambiguates.
- `docmgr skill list` now prints short `Load:` commands (and uses full path only when necessary).
- The updated skills smoke scenario passes end-to-end with concise output.

### What didn't work
- Initially, path scoring was too permissive and could create ambiguity for generic word queries; fixed by requiring the query to look like a path before path scoring is applied.

### What I learned
- It’s easy to accidentally treat arbitrary strings as “path-ish” when using a resolver that always produces a canonical representation; guarding path scoring is important for correctness.
- Generating “best” commands requires global knowledge of the listing (detecting duplicates).

### What was tricky to build
- Balancing resilience with determinism: we want flexible matching, but we must avoid “surprising” best-match selection when multiple candidates are plausible.
- Producing minimal load commands required understanding the workspace root resolution chain and when `--root` is truly necessary.

### What warrants a second pair of eyes
- The match scoring/precedence rules in `pkg/commands/skill_show.go` (especially around ambiguous-but-scored cases).
- The load command choice rules in `pkg/commands/skills_helpers.go` (ensuring they remain intuitive as more skills are added).

### What should be done in the future
- Consider exposing a `--strict` mode for `skill show` that errors unless the match is unambiguous (even when scoring would pick a winner).
- Consider adding a dedicated “skill id” (stable) if collisions become common at scale.

### Code review instructions
- Start in `pkg/commands/skill_show.go` and review the matcher + ambiguity UX.
- Then `pkg/commands/skill_list.go` and `pkg/commands/skills_helpers.go` for load-command generation.
- Finally run the smoke test:
  - `DOCMGR_PATH=/tmp/docmgr-local bash test-scenarios/testing-doc-manager/20-skills-smoke.sh /tmp/docmgr-scenario`


