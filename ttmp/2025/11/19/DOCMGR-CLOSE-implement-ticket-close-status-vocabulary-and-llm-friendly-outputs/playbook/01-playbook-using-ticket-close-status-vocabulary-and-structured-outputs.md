---
Title: 'Playbook: Using Ticket Close, Status Vocabulary, and Structured Outputs'
Ticket: DOCMGR-CLOSE
Status: active
Topics:
    - docmgr
    - workflow
    - ux
    - automation
DocType: playbook
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-11-19T15:41:26.127223608-05:00
---

# Playbook: Using Ticket Close, Status Vocabulary, and Structured Outputs

**Ticket:** DOCMGR-CLOSE  
**Context:** New features added in November 2025 to streamline ticket workflows, status management, and automation.

## Purpose

This playbook shows you how to use the new ticket lifecycle features implemented in DOCMGR-CLOSE:
- Close tickets atomically with `ticket close`
- Manage status vocabulary
- Use structured outputs for automation
- Leverage task completion suggestions

## Prerequisites

- docmgr 0.0.3+ with ticket close support
- Initialized workspace (`docmgr init --seed-vocabulary`)
- Understanding of basic docmgr commands (`ticket create-ticket`, `task add`)

## Feature 1: Closing Tickets Atomically

### The Old Way (3-5 commands)
```bash
# Update status
docmgr meta update --ticket MEN-4242 --field Status --value complete

# Update intent
docmgr meta update --ticket MEN-4242 --field Intent --value long-term

# Add changelog entry
docmgr changelog update --ticket MEN-4242 --entry "Ticket closed"

# Remember to update LastUpdated (handled automatically above)
```

### The New Way (1 command)
```bash
# Close with defaults (status=complete)
docmgr ticket close --ticket MEN-4242

# Close with custom status
docmgr ticket close --ticket MEN-4242 --status archived

# Close with everything specified
docmgr ticket close --ticket MEN-4242 \
  --status complete \
  --intent long-term \
  --changelog-entry "All requirements implemented, tested, and deployed"
```

### What `ticket close` Does
- ✅ Updates Status (default: `complete`, override with `--status`)
- ✅ Optionally updates Intent (via `--intent`)
- ✅ Appends a changelog entry (default: "Ticket closed")
- ✅ Updates LastUpdated timestamp
- ⚠️ Warns if tasks aren't all done (doesn't fail)
- 📊 Returns structured output with `--with-glaze-output`

### Human-Friendly Output
```
$ docmgr ticket close --ticket MEN-4242
Ticket MEN-4242 closed successfully.
Changes: Status: active → complete, Changelog updated, LastUpdated refreshed
Changelog: /path/to/changelog.md
```

### Structured Output (for automation)
```bash
$ docmgr ticket close --ticket MEN-4242 --with-glaze-output --output json
{
  "ticket": "MEN-4242",
  "all_tasks_done": true,
  "open_tasks": 0,
  "done_tasks": 5,
  "status": "complete",
  "intent": "long-term",
  "operations": {
    "status_updated": true,
    "intent_updated": false,
    "changelog_updated": true
  }
}
```

## Feature 2: Status Vocabulary System

### What Changed
Status is now **vocabulary-guided** (teams can customize), not free-form.

### Default Status Values
```bash
$ docmgr vocab list --category status

status: draft — Initial draft state
status: active — Active work in progress
status: review — Ready for review
status: complete — Work completed
status: archived — Archived/completed work
```

### Suggested Transitions (not enforced)
```
draft → active         (start work)
active → review        (ready for review)
review → active        (send back for changes)
review → complete      (approved)
complete → archived    (long-term storage)
complete → active      (reopen - unusual, warns)
```

### Adding Custom Status Values
```bash
# Add a custom status
docmgr vocab add --category status \
  --slug on-hold \
  --description "Work paused, waiting for dependencies"

# Add another
docmgr vocab add --category status \
  --slug blocked \
  --description "Blocked by external dependencies"
```

### Validation
```bash
# Doctor warns on unknown status (doesn't fail)
$ docmgr doctor --ticket MEN-4242

ticket       | issue          | severity | message
-------------|----------------|----------|---------------------------------------------------------------------------
MEN-4242     | unknown_status | warning  | unknown status: in-progress (valid values: draft, active, review, complete, archived; list via 'docmgr vocab list --category status')
```

## Feature 3: Task Completion Suggestions

### Automatic Suggestions
When you check off the last task, `task check` suggests closing the ticket:

```bash
$ docmgr task check --ticket MEN-4242 --id 5
Task checked: 5 (file=.../tasks.md)
💡 All tasks complete! Consider closing the ticket: docmgr ticket close --ticket MEN-4242
```

### Structured Output for Tasks
```bash
# Get task status programmatically
$ docmgr task check --ticket MEN-4242 --id 5 --with-glaze-output --output json
{
  "ticket": "MEN-4242",
  "tasks_file": "/path/to/tasks.md",
  "all_tasks_done": true,
  "open_tasks": 0,
  "done_tasks": 5,
  "total_tasks": 5,
  "checked_ids": [5]
}
```

Use this in scripts to conditionally trigger ticket close or status updates.

## Feature 4: Enhanced Doctor Validation

### Invalid Frontmatter Detection
Doctor now checks ALL markdown files for invalid YAML, not just index.md:

```bash
$ docmgr doctor --ticket MEN-4242

ticket   | issue               | severity | message
---------|---------------------|----------|------------------------------------------
MEN-4242 | invalid_frontmatter | error    | Failed to parse frontmatter: yaml: line 4: did not find expected ',' or ']'
```

This catches:
- Unclosed brackets: `Topics: [test`
- Unclosed braces: `Invalid: {unclosed`
- Missing commas in arrays
- Malformed YAML syntax

### Status Validation
Doctor validates status values against vocabulary:

```bash
$ docmgr doctor --all

ticket   | issue          | severity | message
---------|----------------|----------|---------------------------------------------------------------------------
MEN-4242 | unknown_status | warning  | unknown status: in-progress (valid values: draft, active, review, complete, archived; list via 'docmgr vocab list --category status')
```

## Complete Workflow Example

### Scenario: Implementing and Closing a Feature Ticket

```bash
# 1. Create ticket
docmgr ticket create-ticket \
  --ticket FEAT-100 \
  --title "Add user profile export API" \
  --topics backend,api

# 2. Add tasks
docmgr task add --ticket FEAT-100 --text "Design API endpoint structure"
docmgr task add --ticket FEAT-100 --text "Implement backend handler"
docmgr task add --ticket FEAT-100 --text "Add integration tests"
docmgr task add --ticket FEAT-100 --text "Update API documentation"

# 3. Add design doc
docmgr doc add --ticket FEAT-100 \
  --doc-type design-doc \
  --title "User Profile Export API Design"

# 4. Work through tasks
docmgr task check --ticket FEAT-100 --id 1
# ... implement ...
docmgr task check --ticket FEAT-100 --id 2
# ... implement ...

# 5. Relate implementation files
docmgr doc relate --ticket FEAT-100 \
  --file-note "backend/api/profile/export.go:Main export handler implementation" \
  --file-note "backend/api/profile/export_test.go:Integration tests for export endpoint"

# 6. Update changelog
docmgr changelog update --ticket FEAT-100 \
  --entry "Implemented user profile export API with CSV and JSON formats" \
  --file-note "backend/api/profile/export.go:Export handler" \
  --file-note "backend/api/routes.go:Route registration"

# 7. Check off remaining tasks
docmgr task check --ticket FEAT-100 --id 3,4
# Output: 💡 All tasks complete! Consider closing the ticket: docmgr ticket close --ticket FEAT-100

# 8. Validate before closing
docmgr doctor --ticket FEAT-100 --fail-on error

# 9. Close the ticket
docmgr ticket close --ticket FEAT-100 \
  --changelog-entry "Feature complete: user profile export API implemented, tested, and documented"

# 10. Verify closure
docmgr ticket list --ticket FEAT-100
# Shows: status=complete, tasks=4/4
```

## Automation Patterns

### CI/CD: Close on All Tests Passing
```bash
#!/bin/bash
# In CI pipeline after all tests pass

TICKET="$1"

# Check if all tasks are done (structured output)
TASK_STATUS=$(docmgr task list --ticket "$TICKET" --with-glaze-output --output json)
ALL_DONE=$(echo "$TASK_STATUS" | jq '[.[] | select(.checked == false)] | length == 0')

if [ "$ALL_DONE" = "true" ]; then
  # Close ticket with structured output
  docmgr ticket close --ticket "$TICKET" \
    --changelog-entry "CI: All tests passed, closing automatically" \
    --with-glaze-output --output json | jq
fi
```

### Script: Bulk Status Update
```bash
#!/bin/bash
# Mark all complete tickets as archived

# Get all complete tickets
docmgr ticket list --status complete --with-glaze-output --output json | \
  jq -r '.[] | .ticket' | \
  while read TICKET; do
    echo "Archiving $TICKET..."
    docmgr ticket close --ticket "$TICKET" \
      --status archived \
      --changelog-entry "Archived after 30 days in complete status"
  done
```

### LLM Integration: Check Before Suggesting Close
```python
import subprocess
import json

def should_close_ticket(ticket_id):
    """Check if ticket is ready to close."""
    # Get task status
    result = subprocess.run(
        ["docmgr", "task", "list", "--ticket", ticket_id, 
         "--with-glaze-output", "--output", "json"],
        capture_output=True, text=True
    )
    tasks = json.loads(result.stdout)
    
    # Check if all tasks are done
    all_done = all(task["checked"] for task in tasks)
    
    if all_done:
        return True, f"All {len(tasks)} tasks complete"
    else:
        open_count = sum(1 for t in tasks if not t["checked"])
        return False, f"{open_count} tasks still open"

# Use in LLM workflow
ready, reason = should_close_ticket("MEN-4242")
if ready:
    print(f"✅ Ready to close: {reason}")
    print("Run: docmgr ticket close --ticket MEN-4242")
else:
    print(f"⏸️ Not ready: {reason}")
```

## Quick Reference

### Status Values (Customizable)
| Status | Description | Common Transition From |
|--------|-------------|----------------------|
| draft | Initial draft state | (new ticket) |
| active | Active work in progress | draft, review (rework) |
| review | Ready for review | active |
| complete | Work completed | review |
| archived | Long-term storage | complete |

### Commands
| Command | Purpose | Example |
|---------|---------|---------|
| `ticket close` | Close ticket atomically | `docmgr ticket close --ticket T-1` |
| `vocab list` | Show vocabulary | `docmgr vocab list --category status` |
| `vocab add` | Add custom value | `docmgr vocab add --category status --slug blocked` |
| `task check` | Check task (with suggestion) | `docmgr task check --ticket T-1 --id 3` |
| `doctor` | Validate everything | `docmgr doctor --all --fail-on error` |

### Structured Output Flags
- `--with-glaze-output` — Enable structured mode
- `--output json|yaml|csv|table` — Choose format
- `--select field1,field2` — Pick specific fields
- `--fields field1,field2` — Reorder fields

### When to Use Structured Output
- ✅ CI/CD pipelines
- ✅ Bulk operations (scripts, migrations)
- ✅ LLM orchestration (reliable parsing)
- ✅ Reporting/dashboards
- ❌ Interactive terminal use (human output is better)

## Troubleshooting

### Issue: "ticket not found"
**Cause:** Invalid YAML frontmatter in index.md prevents ticket discovery.

**Solution:**
```bash
# Check for frontmatter errors
docmgr doctor --ticket YOUR-TICKET

# If you see "invalid_frontmatter", fix the YAML syntax
# Common issues:
# - Unquoted colons in Title: "IMPL: My ticket" → Title: 'IMPL: My ticket'
# - Unclosed brackets: Topics: [test → Topics: [test]
# - Duplicate frontmatter blocks (check for multiple --- pairs)
```

### Issue: "unknown_status" warning
**Cause:** Using a status value not in vocabulary.

**Solutions:**
```bash
# Option 1: Use a standard value
docmgr meta update --ticket YOUR-TICKET --field Status --value active

# Option 2: Add your value to vocabulary
docmgr vocab add --category status \
  --slug your-custom-status \
  --description "Description of when to use this"
```

### Issue: Task suggestion doesn't appear
**Cause:** Using `--with-glaze-output` mode (suggestions only in human mode).

**Solution:** Use structured output to check programmatically:
```bash
docmgr task check --ticket T-1 --id 5 --with-glaze-output --output json | jq '.all_tasks_done'
# Returns: true or false
```

## Links to Documentation

- **Ticket context:** `ttmp/2025/11/19/DOCMGR-CLOSE-implement-ticket-close-status-vocabulary-and-llm-friendly-outputs/`
- **Implementation diary:** See `log/01-implementation-diary...` for lessons learned
- **Intern guide:** See `reference/01-intern-guide...` for complete technical details
- **Main tutorial:** `docmgr help how-to-use` (Section 10: Closing Tickets)
- **CLI guide:** `docmgr help cli-guide` (Section 4.2: Ticket Management)

## Pro Tips

1. **Use ticket close for consistency:** Even if you only need to update status, using `ticket close` ensures changelog and LastUpdated are synchronized.

2. **Add custom status values proactively:** If your team uses "on-hold" or "blocked", add them to vocabulary before using them.

3. **Leverage task suggestions:** The automatic suggestion when all tasks are done reduces cognitive load.

4. **Use structured output in scripts:** For reliable parsing in automation, always use `--with-glaze-output --output json`.

5. **Check doctor before closing:** Run `docmgr doctor --ticket YOUR-TICKET` to catch issues before closing.

6. **Quote titles with special characters:** Titles with colons, quotes, or brackets should be quoted: `--title 'FIX: Issue description'`

## Examples by Use Case

### Use Case 1: Developer Closing After Code Review
```bash
# After PR is approved
docmgr ticket close --ticket FEAT-100 \
  --changelog-entry "Code review approved by @alice, merged to main in PR #142"
```

### Use Case 2: PM Archiving Old Tickets
```bash
# Archive tickets completed > 30 days ago
for TICKET in $(docmgr ticket list --status complete --with-glaze-output --output json | jq -r '.[] | .ticket'); do
  docmgr ticket close --ticket "$TICKET" --status archived
done
```

### Use Case 3: LLM Agent Checking Before Suggesting Actions
```bash
# LLM uses structured output to decide
RESULT=$(docmgr task list --ticket FEAT-100 --with-glaze-output --output json)
OPEN=$(echo "$RESULT" | jq '[.[] | select(.checked == false)] | length')

if [ "$OPEN" -eq 0 ]; then
  echo "Suggest: docmgr ticket close --ticket FEAT-100"
else
  echo "Suggest: Complete remaining $OPEN tasks first"
fi
```

### Use Case 4: Setting Custom Status During Close
```bash
# Close as "review" instead of "complete"
docmgr ticket close --ticket FEAT-100 \
  --status review \
  --changelog-entry "Implementation complete, ready for code review"
```

## Verification Checklist

After closing a ticket, verify:

```bash
# 1. Check status updated
docmgr ticket list --ticket YOUR-TICKET
# Should show: status=complete (or your custom status)

# 2. Check changelog entry added
cat $(find ttmp -path "*YOUR-TICKET*" -name "changelog.md")
# Should show dated entry with your message

# 3. Check no doctor errors
docmgr doctor --ticket YOUR-TICKET --fail-on error
# Should exit 0 with no errors

# 4. Verify structured output works
docmgr ticket close --ticket YOUR-TICKET --with-glaze-output --output json | jq
# Should return valid JSON with operations
```

## Related Tickets

- DOCMGR-CLOSE — Implementation ticket for these features
- DOCMGR-FRONTMATTER — Frontmatter parsing improvements
- DOCMGR-CODE-REVIEW — Debate rounds leading to these design decisions
