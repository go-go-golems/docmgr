---
Title: docmgr tutorial validation run
Ticket: DOCMGR-DOC-VERBS
Status: active
Topics:
    - docmgr
    - documentation
    - cli
DocType: script
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-11-24T14:53:28.228927855-05:00
---

# docmgr tutorial validation run

## Purpose

Automate the tutorial validation workflow described in `docmgr help how-to-use` so we can repeatedly smoke-test the instructions. The script spins up a clean Git repo in `/tmp`, creates placeholder backend/frontend files, and executes the canonical `docmgr` commands (init, create-ticket, doc add, relate, tasks, changelog, doctor).

## Usage

```bash
$ ./docmgr/ttmp/2025/11/19/DOCMGR-DOC-VERBS-fix-how-to-use-tutorial-new-verb-structure/script/docmgr-tutorial-validation-run.sh \
    /tmp/test-git-repo

# Override docmgr binary if needed
$ DOCMGR_BIN=/home/manuel/.local/bin/docmgr \
    ./.../script/docmgr-tutorial-validation-run.sh /tmp/test-git-repo
```

## Implementation

- Executable: `script/docmgr-tutorial-validation-run.sh`
- Captures stdout/stderr into `"$TARGET/docmgr-run.log"` via `tee` (override with `LOG_PATH`).
- Steps:
  1. Recreates the target directory (default `/tmp/test-git-repo`) and seeds sample `backend/` + `web/` files.
  2. Runs `docmgr init --seed-vocabulary --root ttmp`.
  3. Creates `MEN-3083`, adds a design doc, relates code files, adds a task, writes changelog entry.
  4. Executes `docmgr doctor --ticket MEN-3083 --fail-on error`.
  5. Prints resulting `ttmp` contents for quick inspection.

## Related scripts

- `script/02-reset-and-recreate-repo.sh`: Thin wrapper that calls `docmgr-tutorial-validation-run.sh` `ITERATIONS` times (default 1) so you can repeatedly wipe and recreate `/tmp/test-git-repo`. Logs for each pass are stored under `/tmp/docmgr-validation-logs` unless you override `LOG_PATH_BASE`.

```bash
# Run the validation workflow three times in a row
$ ITERATIONS=3 script/02-reset-and-recreate-repo.sh /tmp/test-git-repo
```

## Notes

- Requires Git and a writable `/tmp`.
- Safe to rerun; the script wipes the target directory each time.
- Use `LOG_PATH=/custom/path.log` if you need to stash the transcript elsewhere.
- Use the wrapper (`02-reset-and-recreate-repo.sh`) when you want iterative practice or need a consistent repro sequence for usability tests.
