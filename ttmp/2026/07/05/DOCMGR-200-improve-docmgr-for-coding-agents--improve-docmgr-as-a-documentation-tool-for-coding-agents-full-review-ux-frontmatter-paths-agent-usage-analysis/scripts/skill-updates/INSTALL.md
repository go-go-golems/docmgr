# Staged skill updates (docmgr DOCMGR-200 docs refresh)

The skill directories under `~/.claude/skills/` are read-only in the sandbox, so
the refreshed skills are staged here. To install:

```bash
cp skill-updates/docmgr/SKILL.md /home/manuel/.claude/skills/docmgr/SKILL.md
cp skill-updates/diary/SKILL.md  /home/manuel/.claude/skills/diary/SKILL.md
```

## What changed

### docmgr/SKILL.md
- Canonical spellings: `ticket create` / `ticket rename` / `ticket list` (old
  `create-ticket`, `rename-ticket`, `ticket tickets` noted as aliases).
- `docmgr init` now seeds the vocabulary by default (`--seed-vocabulary=false` to skip).
- New `docmgr ticket show <ref>` in the get-oriented workflow.
- Forgiving `--ticket` refs (ID / unique prefix / pasted directory path) and
  `--doc` refs (absolute / cwd / repo / docs-root / duplicated ttmp/ prefix /
  workspace-unique suffix) documented.
- Anchored paths: `relate` writes `repo://` / `ws://` / `docs://` / `abs://`;
  the skill keeps instructing ALWAYS-absolute paths in `--file-note` (they get
  anchored automatically). Note about commas/colons in notes and exit 1 on
  malformed values.
- Output contract section: one-line mutation successes, `--verbose`-gated
  banner/coaching output, non-zero exits on failure (empty `--entry`,
  malformed `--file-note`, failed `meta update`, unknown task ID).
- Stable task IDs (`<!-- t:xxxx -->` markers), `task migrate`, unknown-ID
  behavior (exit 1 + task table).
- Doctor: per-ticket rollup default + `--details`, `--fix` (frontmatter
  auto-repair + anchor migration, `.bak` backups), `--fix-anchors`, `sources/`
  skipped unless `--include-sources`, built-in vocabulary always recognized.
- Dual-mode (`--with-glaze-output`) available on all verbs including mutations.

### diary/SKILL.md
- `task check --id N` example updated to stable task IDs (positions still work).
- Note that related-file paths are stored in anchored form; keep passing
  absolute paths in `--file-note`; malformed values exit 1.

The `references/docmgr.md` long-form reference in the docmgr skill was NOT
restaged; consider regenerating it via `docmgr skill export` from the updated
embedded help if it drifts.
