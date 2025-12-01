---
Title: Root Cause Analysis
Ticket: DOCMGR-TICKET-MOVE-001
Status: active
Topics:
    - bug
    - ticket-move
    - path-template
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-01T17:59:18.642289755-05:00
---

# Root Cause Analysis

## Problem Statement

When a ticket's `index.md` title field contains the ticket identifier, the `docmgr ticket move` command (and `ticket create-ticket`) creates directory names with duplicate ticket identifiers.

**Example:**
- Ticket: `TEST-9999`
- Title: `TEST-9999: Test ticket with ticket in title`
- Resulting directory: `TEST-9999-test-9999-test-ticket-with-ticket-in-title`

## Root Cause

The bug originates in the path rendering logic in `pkg/commands/ticket_move.go` (lines 126-131) and similar logic in `pkg/commands/create_ticket.go`.

### Code Flow

1. **Slug computation** (`ticket_move.go:126-131`):
   ```go
   slug := utils.Slugify(settings.Ticket)  // Initial: "men-5678" (from ticket)
   title := settings.Ticket
   if srcDoc != nil && strings.TrimSpace(srcDoc.Title) != "" {
       title = strings.TrimSpace(srcDoc.Title)  // Override with title from index.md
       slug = utils.Slugify(title)              // Slugify the full title
   }
   ```

2. **Path template rendering** (`ticket_move.go:135`):
   ```go
   destDir, err := renderTicketPath(settings.Root, pathTemplate, settings.Ticket, slug, title, now)
   ```

3. **Template replacement** (`create_ticket.go:317-343`):
   ```go
   replacements := map[string]string{
       "{{TICKET}}": ticket,  // "TEST-9999"
       "{{SLUG}}":   slug,    // "test-9999-test-ticket-with-ticket-in-title" (includes ticket!)
   }
   ```

4. **Result**: `{{TICKET}}-{{SLUG}}` becomes `TEST-9999-test-9999-test-ticket-with-ticket-in-title`

### Why This Happens

- The `utils.Slugify()` function lowercases and slugifies the entire title string, including any ticket identifier that appears in it.
- The path template `{{TICKET}}-{{SLUG}}` assumes the slug is derived from a title that doesn't include the ticket identifier.
- When the title contains the ticket identifier (common in some workflows), the slug includes it, causing duplication.

## Affected Code Locations

1. **`pkg/commands/ticket_move.go`** (lines 126-131)
   - `applyMove()` function computes slug from title without stripping ticket identifier

2. **`pkg/commands/create_ticket.go`** (similar logic)
   - `Run()` function has the same pattern when creating new tickets

3. **`pkg/commands/create_ticket.go`** (function `renderTicketPath`, lines 317-343)
   - Template replacement doesn't account for ticket identifier in slug

## Impact

- **Severity**: Medium
  - Functional impact: Low (directories still work, just have redundant names)
  - UX impact: Medium (confusing directory names, violates DRY principle)
  - Maintenance impact: Low (doesn't break functionality)

- **Affected Commands**:
  - `docmgr ticket create-ticket` (when title includes ticket)
  - `docmgr ticket move` (when moving tickets with titles containing ticket identifier)

- **Frequency**: Occurs whenever a ticket title includes the ticket identifier, which may be common in some workflows.

## Proposed Solutions

### Option 1: Strip Ticket Identifier from Title Before Slugifying (Recommended)

Modify the slug computation to remove the ticket identifier from the title before slugifying:

```go
slug := utils.Slugify(settings.Ticket)
title := settings.Ticket
if srcDoc != nil && strings.TrimSpace(srcDoc.Title) != "" {
    title = strings.TrimSpace(srcDoc.Title)
    // Strip ticket identifier from title before slugifying
    titleForSlug := strings.TrimSpace(strings.TrimPrefix(title, settings.Ticket+":"))
    titleForSlug = strings.TrimSpace(strings.TrimPrefix(titleForSlug, settings.Ticket+" -"))
    titleForSlug = strings.TrimSpace(strings.TrimPrefix(titleForSlug, settings.Ticket+" "))
    if titleForSlug == "" {
        slug = utils.Slugify(settings.Ticket)
    } else {
        slug = utils.Slugify(titleForSlug)
    }
}
```

**Pros:**
- Simple and targeted fix
- Preserves title as-is (only affects slug computation)
- Handles common patterns: "TICKET:", "TICKET -", "TICKET "

**Cons:**
- May not handle all edge cases (e.g., ticket in middle of title)
- Requires pattern matching logic

### Option 2: Use Title Without Ticket for Slug, Keep Full Title for Display

Always compute slug from title, but strip ticket identifier patterns:

```go
func computeSlugFromTitle(title, ticket string) string {
    // Remove common ticket identifier patterns
    patterns := []string{
        ticket + ":",
        ticket + " -",
        ticket + " ",
    }
    cleaned := title
    for _, pattern := range patterns {
        if strings.HasPrefix(cleaned, pattern) {
            cleaned = strings.TrimSpace(strings.TrimPrefix(cleaned, pattern))
            break
        }
    }
    if cleaned == "" || cleaned == title {
        return utils.Slugify(ticket)
    }
    return utils.Slugify(cleaned)
}
```

**Pros:**
- More robust pattern matching
- Reusable function
- Handles edge cases better

**Cons:**
- More complex implementation
- Still may miss some patterns

### Option 3: Change Path Template to Not Include Ticket When Slug Already Contains It

Detect if slug contains ticket identifier and adjust template accordingly. This is more complex and error-prone.

**Pros:**
- No changes to slug computation

**Cons:**
- Complex logic
- Hard to detect all cases
- May cause inconsistencies

## Recommendation

**Option 1** is recommended because:
1. It's a targeted fix that addresses the root cause
2. It's simple and maintainable
3. It handles the most common patterns (ticket at start of title with separator)
4. It preserves backward compatibility (titles without ticket identifiers work as before)

## Testing Strategy

1. **Unit tests** for slug computation with various title patterns:
   - `"TEST-9999: Description"` → slug should be `"description"`
   - `"TEST-9999 - Description"` → slug should be `"description"`
   - `"Description"` → slug should be `"description"` (no ticket in title)
   - `"TEST-9999 Description"` → slug should be `"description"`

2. **Integration tests**:
   - Create ticket with title containing ticket identifier
   - Move ticket with title containing ticket identifier
   - Verify directory names don't have duplication

3. **Edge cases**:
   - Title is exactly the ticket identifier
   - Title starts with ticket but no separator
   - Title contains ticket identifier in middle (should not strip)

## Related Files

- `pkg/commands/ticket_move.go` - Main bug location
- `pkg/commands/create_ticket.go` - Similar bug in ticket creation
- `pkg/utils/slug.go` - Slugify function (may need enhancement)
- `test-scenarios/testing-doc-manager/17-ticket-move.sh` - Test scenario (should be updated to catch this)
