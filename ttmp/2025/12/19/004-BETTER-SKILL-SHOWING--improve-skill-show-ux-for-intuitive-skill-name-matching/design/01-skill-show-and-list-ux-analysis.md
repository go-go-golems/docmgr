---
Title: "Analysis: Skill show/list UX + name resolution"
Ticket: 004-BETTER-SKILL-SHOWING
Status: active
Topics:
  - skills
  - ux
  - cli
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
  - Path: cmd/docmgr/cmds/skill/show.go
    Note: Cobra wiring for `docmgr skill show` (Glazed command wrapped into cobra)
  - Path: pkg/commands/skill_show.go
    Note: Current show logic (only title matching, requires --skill flag)
  - Path: cmd/docmgr/cmds/skill/list.go
    Note: Cobra wiring for `docmgr skill list` (+ flag completion)
  - Path: pkg/commands/skill_list.go
    Note: Current list output prints Path; will be updated to print a load command
  - Path: cmd/docmgr/cmds/common/common.go
    Note: Wrapper that builds cobra commands from Glazed command descriptions
  - Path: ../glazed/pkg/cmds/cmds.go
    Note: Glazed `WithArguments` helper for positional args
  - Path: ../glazed/pkg/cmds/parameters/gather-arguments.go
    Note: Source of "Too many arguments" error when no positional args are declared
  - Path: ../glazed/pkg/cmds/parameters/cobra.go
    Note: Glazed generates cobra `Use` / `Args` validation from argument definitions
  - Path: test-scenarios/testing-doc-manager/20-skills-smoke.sh
    Note: Existing end-to-end scenario test for skills list/show
  - Path: pkg/doc/using-skills.md
    Note: Docs that instruct how to load a skill (currently flag-only)
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-19T00:00:00Z
---

# Analysis: Skill show/list UX + name resolution

## Goal

Explain how `docmgr skill list` and `docmgr skill show` are implemented today, why the current UX fails in the reported ways, and where changes must land to make skill resolution resilient (title/slug/path, ticket scoping, ambiguity UX, and copy/pasteable load commands).

## Current wiring (how it fits together)

### 1) Cobra command tree

- `cmd/docmgr/cmds/skill/skill.go` registers the `docmgr skill` command group and attaches `list` and `show` as subcommands.
- `cmd/docmgr/cmds/skill/list.go` and `cmd/docmgr/cmds/skill/show.go` build cobra commands via `common.BuildCommand(...)`.

### 2) Glazed drives argument parsing (root cause of “Too many arguments”)

`common.BuildCommand` is a thin wrapper around `glazed/pkg/cli.BuildCobraCommand(...)`.

Glazed derives cobra `cmd.Args` validation + positional argument parsing from the command description’s **default layer argument definitions**:
- If a command has **no argument definitions**, any positional argument triggers `GatherArguments(...): "Too many arguments"`.
- This is exactly what’s happening today: `SkillShowCommand` defines only flags and declares **zero arguments**.

### 3) Data source: workspace index + QueryDocs

Both list and show use:
- `workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{RootOverride: ...})`
- `ws.InitIndex(...)`
- `ws.QueryDocs(...)` with `DocType: "skill"`

So the authoritative skill set is “all indexed documents with `DocType: skill` under docs root”.

## Current behavior vs bug report

### Issue: positional argument rejected

Bug report shows:

- `docmgr skill show test-driven-development` → `Error: Too many arguments`

This comes from Glazed’s argument parsing because `skill show` currently declares **no positional arguments**.

### Issue: filename (slug) matching doesn’t work

`pkg/commands/skill_show.go` matches only on `doc.Title` (case-insensitive exact or contains).

So queries like `test-driven-development` won’t match unless the title contains that substring (it usually doesn’t).

### Issue: “Skill: ” prefix makes matching confusing

Current matching is “title contains query”, which can work with partials (e.g. `"Test-Driven"`), but it’s fragile:
- Users naturally try `"Test-Driven Development"` (title without prefix)
- Or the filename slug
- Or a path they copied

## Desired improvements (what changes where)

### `docmgr skill show` improvements

We need to improve `pkg/commands/skill_show.go` in three dimensions:

1) **Input flexibility**
   - Support positional argument: `docmgr skill show <query>`
   - Keep flag support: `docmgr skill show --skill <query>`
   - Add `--ticket` to scope search (important for clashing skills)
   - Accept “path-ish” inputs: absolute path, repo-relative path, docs-root-relative path, directory, file basename with/without extension.

2) **Matching strategy (resilient resolution)**
   - Match against multiple identifiers:
     - title (full)
     - title normalized (strip `Skill:` prefix)
     - filename basename (strip `.md`)
     - filename basename normalized (strip leading numeric prefix like `01-`)
     - path variants (as indexed, repo-relative, docs-root-relative, absolute)
   - Use a scoring strategy to pick a best match when possible.

3) **Ambiguity UX**
   - If multiple skills match and no single best match can be chosen, print a set of **copy/pasteable commands** that load each candidate unambiguously (ideally including `--ticket` when relevant).

### `docmgr skill list` improvements

In `pkg/commands/skill_list.go`, the human-friendly output should not show raw paths as the primary “action”.

Instead, it should print the exact `docmgr skill show ...` invocation that loads the skill (including `--ticket` if needed). This enables workflow:

1) `docmgr skill list`
2) copy/paste the load command shown for the desired skill

Structured output can keep the `path` field, but should also expose a `load_command` field for scripts/LLMs.

## Testing impact

`test-scenarios/testing-doc-manager/20-skills-smoke.sh` already covers:
- list filters (`--ticket`, `--topics`, `--file`, `--dir`)
- show via `--skill` exact and partial matching

We should extend it to cover:
- positional show argument
- filename/slug matching
- path matching
- clashing skills (same title or same slug in different locations)
- `skill show --ticket ...` to disambiguate


