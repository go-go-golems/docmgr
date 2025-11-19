---
Title: How to Work on Any Ticket with docmgr
Slug: how-to-work-on-any-ticket
Short: Step-by-step checklist for taking over any ticket workspace and keeping docmgr metadata aligned.
Topics:
- docmgr
- workflow
- onboarding
- tickets
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Tutorial
---

# How to Work on Any Ticket with docmgr

Whenever you inherit ticket `<TICKET-ID>` inside repository `<REPO-PATH>`, follow this playbook to get oriented, understand the current context, and keep the workspace in a compliant state. Every section builds on the previous one so you can ramp up quickly without dropping important metadata.

## Step 0: Confirm the Workspace and Refresh docmgr Basics

Before touching ticket files, confirm that you can run docmgr at the repository root. This ensures the help system is available and that the ticket metadata already exists.

```bash
cd <REPO-PATH>

docmgr help how-to-use
docmgr ticket list --ticket <TICKET-ID>
docmgr doc list --ticket <TICKET-ID>
docmgr task list --ticket <TICKET-ID>
```

If any command fails, fix the repository setup (see `docmgr help how-to-setup`) before proceeding.

## Step 1: Review the Ticket Source Material

Read all existing documentation in order so you understand why the ticket exists and what has already been attempted.

1. Open the ticket index (typically `ttmp/<TICKET-ID>/index.md`) for the canonical summary.
2. Inspect implementation diaries under `log/`, reading entries chronologically to catch historical context.
3. Review the current `tasks.md` and `changelog.md` to see outstanding work and completed changes.
4. Skim any background docs referenced by `docmgr doc list --ticket <TICKET-ID>` and note prerequisites or dependencies.

## Step 2: Start with the Highest-Priority Tasks

Always begin with the next unchecked task so progress stays orderly. Use the CLI to see and update task status as you work.

```bash
docmgr task list --ticket <TICKET-ID>
docmgr task check --ticket <TICKET-ID> --id <TASK-ID>
```

Update the list as soon as you complete a meaningful unit of work and capture any new subtasks that emerge.

## Step 3: Keep Files and the Changelog in Sync

Every modification must be traceable. Relate files immediately after edits and log the change so future maintainers know what happened and why.

```bash
docmgr doc relate --ticket <TICKET-ID> \
  --file-note "/ABS/PATH/TO/FILE:Why this file matters right now"

docmgr changelog update --ticket <TICKET-ID> \
  --entry "What changed and why" \
  --file-note "/ABS/PATH/TO/FILE:Reason"
```

Use absolute paths for clarity, and group related changes into a single changelog entry with multiple `--file-note` values if needed.

## Step 4: Maintain an Implementation Diary

After each significant step, jot down what you tried, what succeeded or failed, and what to do next. Append to the active diary in `log/` or create a new note under `log/various/` if no diary exists yet. These entries become the institutional memory for the ticket.

## Step 5: Capture Repo-Specific Intelligence

Document any local setup, build commands, unusual comparison steps, or environment switches that apply to this repository. Add these notes to the ticket workspace (often `various/` or a dedicated reference doc) so the next person can reproduce your environment without guesswork.

## Step 6: Track Known Issues and Immediate Focus

Keep a running list of blockers, gaps, and temporary workarounds. Update it whenever you discover a new risk so planning conversations have an up-to-date source of truth. Mention follow-up tasks if they cannot be addressed immediately.

## Step 7: Close the Ticket When Done

When all tasks are complete and work is ready for review or deployment, use `ticket close` to atomically update status, changelog, and timestamps:

```bash
# Check if all tasks are done
docmgr task list --ticket <TICKET-ID>

# Close with defaults (status=complete)
docmgr ticket close --ticket <TICKET-ID>

# Or close with custom status
docmgr ticket close --ticket <TICKET-ID> --status review --changelog-entry "Implementation complete, ready for review"
```

**What `ticket close` does:**
- Updates Status (default: `complete`, override with `--status`)
- Optionally updates Intent (via `--intent`)
- Appends a changelog entry
- Updates LastUpdated timestamp
- Warns if tasks aren't all done (doesn't fail)

**Pro tip:** When you check off the last task with `docmgr task check`, it automatically suggests running `ticket close`.

## docmgr Helpers at a Glance

Re-run the status and metadata commands whenever you context-switch to ensure nothing drifted:

```bash
docmgr status --summary-only
docmgr ticket list
docmgr meta update --ticket <TICKET-ID> --field Status --value active
docmgr ticket close --ticket <TICKET-ID>  # When done
```

## Where to Go Next

After finishing this checklist, return to the task list, confirm priorities with the ticket owner, and continue iterating through tasks, file relations, changelog entries, and diary updates. When all tasks are complete, use `docmgr ticket close` to finalize the work. This loop keeps every ticket workspace healthy and auditable.
