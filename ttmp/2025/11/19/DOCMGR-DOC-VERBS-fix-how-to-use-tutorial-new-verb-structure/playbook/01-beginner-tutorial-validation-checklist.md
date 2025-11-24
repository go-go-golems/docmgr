---
Title: Beginner tutorial validation checklist
Ticket: DOCMGR-DOC-VERBS
Status: active
Topics:
    - docmgr
    - documentation
    - cli
DocType: playbook
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-11-24T20:22:00-05:00
---

# Beginner tutorial validation checklist

## Purpose

Ensure `docmgr help how-to-use` is understandable by having a newcomer follow the official tutorial, recreate the sample repo, and answer a few sanity-check questions without prior context.

## Environment Assumptions

- You are on the repo root: `/home/manuel/workspaces/2025-11-24/review-docmgr-how-to-use-doc/docmgr`.
- `docmgr` binary is available at `/home/manuel/.local/bin/docmgr` (already on PATH for manuel).
- Bash shell available.
- `/tmp` is writable (script wipes `/tmp/test-git-repo` on every run).

## Safety guardrails (avoid common mistakes)

- Do NOT manually delete `/tmp/test-git-repo` between steps. The reset script wipes the target for you. If you need a fresh state, re-run Step 2.
- Do NOT chain commands with `&&` or `;`. Run commands one per line, verify output, then proceed.
- After Step 2, verify the repo exists before continuing:

```bash
test -d /tmp/test-git-repo/ttmp || { echo "missing /tmp/test-git-repo/ttmp"; exit 1; }
ls -R /tmp/test-git-repo/ttmp | head -50
/home/manuel/.local/bin/docmgr status --root /tmp/test-git-repo/ttmp --summary-only
```

- When changing directories, run `cd` alone and then `pwd` to confirm before the next command.

## Commands / Checklist

1. **Skim the tutorial once:** Run `docmgr help how-to-use` and spend 5 minutes reading Part 1 (Essentials). While reading, note any words you don't understand.
2. **Reset the practice repo:**  
   ```bash
   cd /home/manuel/workspaces/2025-11-24/review-docmgr-how-to-use-doc
   docmgr/ttmp/2025/11/19/DOCMGR-DOC-VERBS-fix-how-to-use-tutorial-new-verb-structure/script/02-reset-and-recreate-repo.sh /tmp/test-git-repo
   ```  
   (Set `ITERATIONS=3` to repeat the workflow back-to-back.)
   Note: The script wipes the target each run; you do not need (and should not) `rm -rf` the directory yourself.
3. **Walk through the tutorial manually:** After the script run, cd into `/tmp/test-git-repo` and re-run each command from the tutorial yourself (init, ticket create, doc add, relate, task, changelog, doctor). Compare the CLI output to the tutorial screenshots/snippets. Ensure `docmgr doc relate` is invoked with **multiple** `--file-note` flags in a single command so they experience the batching workflow.
   Before continuing, confirm the repo exists and print status:
   ```bash
   test -d /tmp/test-git-repo/ttmp || { echo "missing /tmp/test-git-repo/ttmp"; exit 1; }
   cd /tmp/test-git-repo
   pwd
   /home/manuel/.local/bin/docmgr status --root ttmp --summary-only
   ```
   Then run the tutorial commands one-by-one (no chaining):
   ```bash
   /home/manuel/.local/bin/docmgr init --seed-vocabulary --root ttmp
   /home/manuel/.local/bin/docmgr ticket create-ticket --ticket MEN-3083 --title "Tutorial validation ticket" --topics test,backend
   /home/manuel/.local/bin/docmgr doc add --ticket MEN-3083 --doc-type design-doc --title "Placeholder design context"
   /home/manuel/.local/bin/docmgr doc relate --ticket MEN-3083 \
     --file-note "backend/api/register.go:Registers API routes (normalization logic)" \
     --file-note "web/src/store/api/chatApi.ts:Frontend integration"
   /home/manuel/.local/bin/docmgr task add --ticket MEN-3083 --text "Update API docs for /chat/v2"
   /home/manuel/.local/bin/docmgr changelog update --ticket MEN-3083 \
     --entry "Initial tutorial validation pass" \
     --file-note "backend/api/register.go:Source implementation for normalization"
   /home/manuel/.local/bin/docmgr doctor --root ttmp --ticket MEN-3083 --stale-after 30 --fail-on error
   ```
4. **Answer the questions:** Without looking at this repo, record your answers in `ttmp/2025/11/24/MEN-3083-.../README.md`:  
   - Describe the on-disk layout for a newly created ticket.  
   - Name the command that creates a design doc and explain where the resulting file appears.  
   - Explain how to relate more than one file (with notes) in a single CLI invocation.  
   - Describe the CLI verbs you reach for when tracking to-dos versus recording notable progress.  
   - Summarize the warning produced by `docmgr doctor` during the sample run and outline the follow-up action.  
   - Outline the steps to append another changelog entry that includes file notes.  
   - Describe how to relate files to a specific subdocument rather than the ticket index.  
   - Explain how to learn which topic/status values are acceptable when doctor reports an unknown value.
5. **Record confusion:** Every time something feels unclear, append a dated bullet to `docmgr/ttmp/2025/11/19/DOCMGR-DOC-VERBS-fix-how-to-use-tutorial-new-verb-structure/working-note/01-tutorial-clarity-findings.md`. Include the tutorial section or workflow step where it happened (e.g., “Step 3 – relating files with multiple notes was unclear”) so we can trace fixes later. Call out when the tutorial repeats itself unnecessarily or when an instruction is simply wrong; flag those explicitly so we can prioritize fixes.

## Exit Criteria

- `/tmp/test-git-repo/ttmp/2025/11/24/MEN-3083-tutorial-validation-ticket/` exists with design doc, tasks, changelog populated.
- You can answer the questions above without re-reading the tutorial.
- Any confusion, errors, or mismatched outputs are logged so we can fix the docs.

## Notes

- Expected warning: `docmgr doctor` will emit `unknown_topics: [test]` because the seeded vocabulary doesn't include `test`; that's fine for now.
- Use `/tmp/docmgr-validation-logs/docmgr-run-*.log` if you need to attach raw command output to a bug report.
- The reset script resets the target directory each run; prefer re-running it over manually deleting `/tmp/test-git-repo`.
