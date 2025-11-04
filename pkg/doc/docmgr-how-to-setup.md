---
Title: Tutorial — Setting up docmgr in a Codebase
Slug: how-to-setup
Short: Initialize vocabulary, scaffold templates and guidelines, and enforce health checks with doctor.
Topics:
- docmgr
- setup
- operations
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Tutorial
---

## 1. Purpose

This guide shows how to bootstrap and maintain the documentation system for a repository using `docmgr`. It explains not only the commands to run but also the rationale behind them, so new contributors understand the “why” as well as the “how”. You’ll set up vocabulary, scaffold templates/guidelines, and integrate `doctor` checks for ongoing health. For a deeper dive into writing experience and expectations, see:

- `docmgr help templates-and-guidelines` — how templates and guidelines work, and how to customize them
- `docmgr help cli-guide` — overview of core commands and concepts

## 2. Repository Conventions

- Docs live under `ttmp/` at the repository root. This keeps work-in-progress close to code and easy to review.
- Ticket workspaces are directories named `<TICKET>-<slug>/` under `ttmp/`. The slug is derived from the ticket title: lowercase; any non‑alphanumeric becomes `-`; consecutive `-` are collapsed; leading/trailing `-` are trimmed. For example, `go-go-mento: Webchat/Web hydration and integration reference` → `go-go-mento-webchat-web-hydration-and-integration-reference`.
- Vocabulary file lives at `ttmp/vocabulary.yaml` by default (configurable via `.ttmp.yaml` → `vocabulary`). It defines the allowed `Topics`, `DocType`, and `Intent` values.
- Scaffolding directories at root: `ttmp/_templates/`, `ttmp/_guidelines/`. Teams customize these to encode house style.

Per-ticket workspace contents created by `init`:

- `index.md`
- `design/`, `reference/`, `playbooks/`
- `scripts/`, `sources/`, `various/`, `archive/`
- `.meta/`

### 2.1 Repository Configuration (.ttmp.yaml)

Place a `.ttmp.yaml` at the repository root to configure defaults. The CLI searches for this file by walking up from the current directory until it finds the nearest `.ttmp.yaml`. When the config uses relative paths (for example, `root: ttmp`), they are interpreted relative to the directory that contains `.ttmp.yaml`.

```yaml
root: ttmp
defaults:
  owners: [manuel]
  intent: long-term
filenamePrefixPolicy: off
docTypeToggles:
  design-doc: true
  reference: true
  playbook: true
```

- `root`: default docs root (overrides the built-in `ttmp` when flags are not explicitly set)
- `defaults.owners` / `defaults.intent`: applied when initializing ticket index metadata
- `filenamePrefixPolicy`: reserved for future filename enforcement
- `docTypeToggles`: reserved for controlling allowed types (not enforced yet)

## 3. Seed Vocabulary

Start with a minimal, agreed vocabulary to prevent drift in metadata. These values become the shared language of your documentation; they drive search filters and keep frontmatter consistent. Keep the list short at first and evolve with consensus:

```bash
docmgr vocab add --category topics   --slug backend --description "Backend services"
docmgr vocab add --category topics   --slug frontend --description "Frontend app"
docmgr vocab add --category topics   --slug infrastructure
docmgr vocab add --category docTypes --slug index
docmgr vocab add --category docTypes --slug design-doc
docmgr vocab add --category docTypes --slug reference
docmgr vocab add --category docTypes --slug playbook
docmgr vocab add --category intent   --slug long-term
```

This writes the vocabulary file to `ttmp/vocabulary.yaml` by default (or to the path configured via `.ttmp.yaml:vocabulary`). Teams can evolve entries over time; `doctor` validates against this file.

To introduce a new document type:

```bash
docmgr vocab add --category docTypes --slug til --description "Today I Learned"
docmgr add --ticket MEN-XXXX --doc-type til --title "TIL — <topic>" --root ttmp
```

If there is a template at `ttmp/_templates/til.md`, it will be used; otherwise the file is created under `various/` with frontmatter `DocType: til` so it still participates in filters and validation.

Guidance:
- Topics should reflect your architecture or domains (for example, `backend`, `frontend`, `websocket`, `observability`)
- `DocType` is how readers will approach the doc (for example, `design-doc`, `reference`, `playbook`)
- `Intent` reflects longevity (for example, `long-term` for docs meant to persist)

## 4. Scaffold Templates and Guidelines

The `_templates/` and `_guidelines/` directories are created automatically the first time you run `init`. If you don't want to create a real ticket yet, initialize a temp workspace and delete it afterwards — the scaffolds remain:

```bash
docmgr init --ticket TMP-BOOTSTRAP --title "Bootstrap doc system" --root ttmp
rm -rf ttmp/TMP-BOOTSTRAP-bootstrap-doc-system
```

Now place house-style templates and guidelines under:

- `ttmp/_templates/<docType>.md`
- `ttmp/_guidelines/<docType>.md`

Developers can preview guidelines:

```bash
docmgr guidelines --doc-type design-doc --output markdown
```

Recommendations:
- Use templates for copy-ready section scaffolds (executive summary, decisions, plan)
- Use guidelines to explain intent and quality bars (what reviewers look for)
- See `docmgr help templates-and-guidelines` for examples and best practices

## 5. Enrich RelatedFiles with Rationale (optional best practice)

Encourage contributors to add a brief note explaining why each related file matters. This improves review speed and LLM context quality.

```bash
# Add files and attach notes
docmgr relate --ticket MEN-4242 \
  --files backend/chat/api/register.go,web/src/store/api/chatApi.ts \
  --file-note "backend/chat/api/register.go:Registers chat routes (source of path normalization)" \
  --file-note "web/src/store/api/chatApi.ts=Frontend API integration; must align with backend paths"

# When applying suggestions, reasons are stored as notes unless overridden
docmgr relate --ticket MEN-4242 --suggest --apply-suggestions --query WebSocket
```

`meta update --field RelatedFiles` still works and creates entries without notes; prefer `relate` for adding context.

## 6. Enforce Health with Doctor

Run `doctor` locally and in automation to keep the system healthy. It catches stale docs, missing fields, unknown vocabulary, and broken file references. Customize ignore patterns to reduce noise without hiding real issues:

```bash
# Local checks (ignore scaffolding and raise on errors)
docmgr doctor --root ttmp --ignore-dir _templates --ignore-dir _guidelines --stale-after 30 --fail-on error

# Ignore known duplicate index (example)
docmgr doctor --root ttmp --ignore-glob "ttmp/*/design/index.md" --fail-on warning
```

Doctor checks include:

- Missing/invalid `index.md`
- Stale documents (configurable threshold)
- Required fields (Title, Ticket, Status, Topics)
- Unknown Topics/DocType/Intent (validated against `doc/vocabulary.yaml`)
- Missing `RelatedFiles` on disk

Tip: Set `--stale-after` high initially (for example, 30–45 days) while adoption ramps up, then lower it as your cadence stabilizes.

Repository ignores:

Create a `.docmgrignore` at the repository root (or inside the docs root, e.g., `ttmp/.docmgrignore`) to exclude paths from validation (comments with `#`; use globs or names). Example:

```gitignore
# VCS and build artifacts
.git/
node_modules/
dist/

# Suppress noisy nested index for a specific directory layout
ttmp/*/design/index.md
```

## 7. CI Integration

Add a job to fail fast on documentation regressions. Make CI strict over time (for example, start with `--fail-on error`, later consider `--fail-on warning` if noise is low):

```yaml
- name: Validate docs
  run: |
    docmgr doctor --root ttmp \
      --ignore-dir _templates --ignore-dir _guidelines \
      --stale-after 30 --fail-on error
```

## 8. Operational Tips

- Keep vocabulary small and stable; socialize changes via PRs
- Encourage `Owners`, `Summary`, and `RelatedFiles` on `index.md` for every ticket
- Use `search` regularly (content, reverse lookup, external sources) during reviews
- Treat `doctor` warnings as actionable tech-debt where appropriate; track follow-ups
- Revisit templates/guidelines quarterly to reflect lessons learned
- Manage ticket checklists with `docmgr tasks` (list/add/check/uncheck/edit/remove)
- Append decisions and progress to `changelog.md` with `docmgr changelog update`


