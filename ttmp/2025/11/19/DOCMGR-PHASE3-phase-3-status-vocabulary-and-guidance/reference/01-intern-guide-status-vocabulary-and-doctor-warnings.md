---
Title: Intern Guide — Status Vocabulary and Doctor Warnings
Ticket: DOCMGR-PHASE3
Status: active
Topics:
    - tools
    - ux
    - documentation
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-11-19T16:29:53.424402678-05:00
---

# Intern Guide — Status Vocabulary and Doctor Warnings

## Goal

Help you implement and use the Phase 3 workflow improvements:
- Add/extend `status` vocabulary (team‑extensible)
- Make `status` values discoverable
- Warn on unknown `status` (doctor), not fail
- Document suggested (non‑enforced) transitions

## Context

From the debate synthesis and Round 3:
- Status becomes vocabulary‑guided with warnings (not errors)
- Teams can customize via `vocabulary.yaml` or `docmgr vocab add`
- Suggested transitions guide usage, but are not enforced
- `ticket close` typically sets `status=complete`; teams can override

## Quick Reference

Status vocabulary (seeds):

```yaml
# ttmp/vocabulary.yaml
status:
  - slug: draft
    description: Initial draft state
  - slug: active
    description: Active work in progress
  - slug: review
    description: Ready for review
  - slug: complete
    description: Work completed
  - slug: archived
    description: Archived/completed work
```

Suggested transitions (not enforced):
- draft → active → review → complete → archived
- review → active (send back)
- complete → active (reopen) — warn as unusual

CLI discovery:

```bash
# List status vocabulary
docmgr vocab list --category status --with-glaze-output --output table

# Add a custom status
docmgr vocab add --category status --slug on-hold --description "Work paused"
```

Doctor warnings:

```bash
# Warn on unknown status (does not fail unless --fail-on warning)
docmgr doctor --ticket DOCMGR-PHASE3 --with-glaze-output --output table

# Typical warning message:
# unknown_status  warning  unknown status: done (valid values: draft, active, review, complete, archived; list via 'docmgr vocab list --category status')
```

Ticket close relationship:

```bash
# Close a ticket with defaults (status=complete)
docmgr ticket close --ticket YOUR-123

# Override explicitly
docmgr ticket close --ticket YOUR-123 --status archived
```

## Usage Examples

End‑to‑end setup for a new repo:

```bash
# Initialize and seed vocabulary
docmgr init --seed-vocabulary

# Inspect seeded statuses
docmgr vocab list --category status --with-glaze-output --output table

# Create ticket and add docs
docmgr ticket create-ticket --ticket FEAT-100 --title "Feature 100"
docmgr doc add --ticket FEAT-100 --doc-type reference --title "Status Vocabulary Cheatsheet"

# Validate (expect warnings only for unknowns)
docmgr doctor --ticket FEAT-100 --fail-on error
```

## Related

- Synthesis: `docmgr/ttmp/2025/11/18/.../reference/06-debate-synthesis-closing-workflow-status-intent-and-llm-ux.md`
- Round 3: `docmgr/ttmp/2025/11/18/.../analysis/14-debate-round-3-status-and-intent-lifecycle-transitions.md`
- `docmgr help how-to-use` (Closing Tickets, Vocabulary, Doctor sections)
