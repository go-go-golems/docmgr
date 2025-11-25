# Tutorial validation review

## Summary

- Followed the checklist in `playbook/01-beginner-tutorial-validation-checklist.md` step by step: read `docmgr help how-to-use`, reset the practice repo, reran the tutorial commands manually, answered the guided questions, recorded clarifications, and captured doubts in the working note.
- Confirmed the workspace now contains the expected design doc, tasks, and changelog entries in `ttmp/2025/11/24/MEN-3083-tutorial-validation-ticket/`.

## What worked well

- The tutorial spells out every CLI verb with precise arguments, so it was easy to copy the commands (one per line) and observe matching output—the docs for `docmgr init`, `ticket create-ticket`, `doc add`, `task add`, and `changelog update` behave as advertised.
- Running `docmgr doctor` reliably surfaced the expected `unknown_topics` warning, which reinforces the purpose of the doctor command and helped validate the sample workflow.
- The checklist prompts you to “experience” multiple `--file-note` flags in one command; re-running `docmgr doc relate` with the two example file notes demonstrated the batching UX exactly as envisioned.

## Issues encountered

- `docmgr doc relate --ticket ...` on the ticket index immediately after the reset script returns `Error: no changes specified.` The script already adds those related files, so the command looks like a failure even though nothing needs to change. Clarify in the tutorial that the helper script seeds these relationships, or show an alternative such as relating the files to a subdocument (which is a valuable learning point itself).
- The CLI reference text embedded in `docmgr help how-to-use` is verbose and sometimes feels like it repeats the same lesson (e.g., multiple sections reiterate why related files need notes). A shorter “Part 1: Essentials” summary with a handful of targeted examples would help newcomers stay focused during their first 5-minute skim .
- The expected `unknown_topics` warning for the `test` topic is noted in the checklist notes, but there is no immediate pointer on how to address it if someone wants to eliminate the warning. Recommend adding a quick “Fix it” section showing `docmgr vocab add --category topics --slug test --description "...“` or similar.

## Improvements

- Explicitly remind readers that they should run `cd /tmp/test-git-repo` then `pwd` before issuing commands (the checklist already says this, but the tutorial text could mirror it so the instructions are easier to discover without jumping back to the checklist).
- When describing file relations, add a short side-by-side example showing how to target a subdocument (`docmgr doc relate --doc ...`) versus the ticket index; this reinforces the best practice of keeping `index.md` minimal.
- Encourage readers to append their answers to the tutorial questions inside `README.md` (as requested by step 4) and mention that the answers are later used to gauge whether they “get” the workflow without rereading the repo.

## Follow-up actions taken

- Logged the confusing `docmgr doc relate` interaction in `working-note/01-tutorial-clarity-findings.md` alongside the “root vs CWD” reminder so the doc team can prioritize clarifying the helper script’s behavior.
- Answered the checklist questions directly in `README.md` (the new “Tutorial Review Questions” section).
- Left the raw terminal outputs (each command run) in `/tmp/docmgr-validation-logs/docmgr-run-*.log` should someone want to inspect full command feedback.

