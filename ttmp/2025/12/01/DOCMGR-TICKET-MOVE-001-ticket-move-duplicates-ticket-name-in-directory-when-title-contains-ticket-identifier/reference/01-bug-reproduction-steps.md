---
Title: Bug Reproduction Steps
Ticket: DOCMGR-TICKET-MOVE-001
Status: active
Topics:
    - bug
    - ticket-move
    - path-template
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-01T17:59:17.577211481-05:00
---

# Bug Reproduction Steps

## Goal

Reproduce the ticket name duplication bug in `docmgr ticket move` and `docmgr ticket create-ticket` commands.

## Context

The bug occurs when a ticket's title field contains the ticket identifier. The slug computation includes the ticket name from the title, and the path template combines `{{TICKET}}` with `{{SLUG}}`, causing duplication.

## Quick Reference

### Reproduction Steps

**Scenario 1: Create ticket with title containing ticket identifier**

```bash
# Create a ticket where the title includes the ticket identifier
docmgr ticket create-ticket --ticket TEST-9999 \
  --title "TEST-9999: Test ticket with ticket in title" \
  --topics test

# Check the resulting directory name
ls -la ttmp/2025/12/01/ | grep TEST-9999
# Output: TEST-9999-test-9999-test-ticket-with-ticket-in-title
#         ^^^^^^^^ ^^^^^^^^
#         ticket   slug (includes ticket from title)
```

**Scenario 2: Move ticket with title containing ticket identifier**

```bash
# Create a ticket with title containing ticket identifier
docmgr ticket create-ticket --ticket TEST-8888 \
  --title "TEST-8888: Another test" \
  --topics test

# Move the ticket directory to simulate legacy layout
mv ttmp/2025/12/01/TEST-8888-test-8888-another-test ttmp/TEST-8888-legacy

# Run ticket move
docmgr ticket move --ticket TEST-8888 --overwrite

# Check the resulting directory name
find ttmp -type d -name "*TEST-8888*"
# Output: ttmp/2025/12/01/TEST-8888-test-8888-another-test
#         ^^^^^^^^ ^^^^^^^^
#         ticket   slug (includes ticket from title)
```

**Scenario 3: Using test scenario 17**

The test scenario `17-ticket-move.sh` doesn't currently expose this bug because it uses a title that doesn't include the ticket identifier. However, if you modify the ticket creation to include the ticket in the title, the bug will manifest.

## Usage Examples

### Expected Behavior

When a ticket title contains the ticket identifier, the slug should strip or ignore the ticket identifier portion to avoid duplication:

- Ticket: `TEST-9999`
- Title: `TEST-9999: Description`
- Expected directory: `TEST-9999-description` (not `TEST-9999-test-9999-description`)

### Actual Behavior

- Ticket: `TEST-9999`
- Title: `TEST-9999: Description`
- Actual directory: `TEST-9999-test-9999-description` ❌

## Related

- See [Root Cause Analysis](../analysis/01-root-cause-analysis.md) for technical details
- Affected code: `pkg/commands/ticket_move.go` lines 126-131
- Affected code: `pkg/commands/create_ticket.go` (similar logic)
