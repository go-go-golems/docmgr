---
Title: Debate Round 3 - Status and Intent Lifecycle Transitions
Ticket: DOCMGR-CODE-REVIEW
Status: active
Topics:
    - docmgr
    - code-review
DocType: analysis
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: Debate Round 3 exploring status/intent lifecycle: consensus on vocabulary control for Status with warnings (not errors), suggested transitions (not enforced), context-aware defaults, explicit overrides
LastUpdated: 2025-11-19T14:37:48.179923161-05:00
---

# Debate Round 3 - Status and Intent Lifecycle Transitions

## Question

Should status/intent transitions be explicit (user chooses) or implicit (derived from task completion)? Should we enforce valid transitions? Should status be vocabulary-controlled like intent?

## Pre-Debate Research

### Research Commands and Results

**1. Current Status Value Patterns**
```bash
$ grep -r "Status:" docmgr/ttmp/ | grep -v "Status: \"\"" | head -10
Status: active
Status: draft
Status: review
Status: complete
```

**Finding:** Common status values: `active`, `draft`, `review`, `complete`, `needs-review`, `archived`. No validation enforced.

**2. Intent Vocabulary Structure**
```bash
$ cat docmgr/ttmp/vocabulary.yaml | grep -A 2 "intent:"
intent:
    - slug: long-term
      description: Long-term documentation
```

**Finding:** Intent is vocabulary-controlled (only `long-term` currently defined). Status is free-form (no vocabulary).

**3. Status Default Analysis**
```bash
$ grep -B 2 -A 2 "Status.*active\|Status.*=" docmgr/pkg/commands/create_ticket.go
		Status:  "active",
```

**Finding:** Tickets default to `Status: "active"` when created. Hardcoded, not configurable.

**4. Intent Default Analysis**
```bash
$ grep -B 5 -A 5 "Intent.*long-term" docmgr/pkg/commands/create_ticket.go
		Intent: func() string {
			if cfg != nil && cfg.Defaults.Intent != "" {
				return cfg.Defaults.Intent
			}
			return "long-term"
		}(),
```

**Finding:** Intent defaults to `"long-term"` but is configurable via `.ttmp.yaml`. Status defaults are hardcoded.

**5. Task Completion Detection**
```bash
$ grep -A 10 "func countTasksInTicket" docmgr/pkg/commands/list_tickets.go
func countTasksInTicket(ticketDir string) (int, int) {
	path := filepath.Join(ticketDir, "tasks.md")
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, 0
	}
	lines := strings.Split(strings.ReplaceAll(string(content), "\n"), "\n")
	tasks := parseTasksFromLines(lines)
	done := 0
	for _, t := range tasks {
		if t.Checked {
			done++
		}
	}
	open := len(tasks) - done
	return open, done
}
```

**Finding:** Code exists to detect task completion (`countTasksInTicket`). Can determine "all tasks done" state.

**6. Status Transition Patterns**
From documentation analysis:
- `active` → `review` (when ready for review)
- `review` → `active` (back to work)
- `active` → `complete` (when done)
- `complete` → `archived` (long-term storage)

**Finding:** Common transitions exist but aren't enforced. Users can set any status from any status.

**7. Doctor Validation**
```bash
$ grep -A 5 "doc.Status == \"\"" docmgr/pkg/commands/doctor.go
		if doc.Status == "" {
			issues = append(issues, "missing Status")
		}
```

**Finding:** Doctor only checks for **presence** of Status, not **validity** or **transitions**. No vocabulary validation for Status.

## Opening Statements (Round 1)

### Alex Rodriguez — "The Architect"

*[Drawing state machine diagram]*

Looking at the codebase, I see a **fundamental asymmetry**:

- **Intent:** Vocabulary-controlled, validated by doctor
- **Status:** Free-form string, no validation

This creates confusion. Status has **6 common values** (`active`, `draft`, `review`, `complete`, `needs-review`, `archived`) but users can set **any string**. That's inconsistent!

**My position:** Status should be **vocabulary-controlled** like Intent. Add to `vocabulary.yaml`:

```yaml
status:
  - slug: draft
    description: Initial draft, not yet active
  - slug: active
    description: Actively being worked on
  - slug: review
    description: Ready for review
  - slug: complete
    description: Work completed
  - slug: archived
    description: Archived for reference
```

**Also:** Enforce **valid transitions**. Not all status changes make sense:
- ✅ `draft` → `active` (starting work)
- ✅ `active` → `review` (ready for review)
- ✅ `review` → `active` (back to work)
- ✅ `active` → `complete` (done)
- ❌ `complete` → `draft` (doesn't make sense)
- ❌ `archived` → `active` (should require explicit `reopen`)

**State machine design:** Define valid transitions, enforce them in `ticket close` and `meta update`.

---

### `pkg/models/document.go` — "The Data Guardian"

*[Showing current model]*

Alex, I understand your desire for consistency, but **Status and Intent serve different purposes**:

- **Intent:** Longevity expectation (long-term vs. temporary). Set once, rarely changes.
- **Status:** Workflow state (draft → active → review → complete). Changes frequently.

**Current Status field:**
```go
Status string `yaml:"Status" json:"status"`
```

It's a **string** because teams have different workflows. Some teams use:
- `in-progress`, `blocked`, `waiting-for-review`
- `needs-feedback`, `stale`, `deprecated`
- Custom statuses for their process

**My position:** Keep Status **free-form** but add **validation warnings**:
- If status matches vocabulary → ✅ valid
- If status doesn't match → ⚠️ warning (not error)
- Allow custom statuses for team flexibility

**For transitions:** Make them **suggestions**, not requirements:
- When closing ticket, suggest `complete` or `archived`
- Don't enforce transitions, but warn on unusual ones
- Let teams define their own workflows

**The asymmetry is intentional:** Intent is **controlled** (longevity), Status is **flexible** (workflow).

---

### `pkg/commands/tasks.go` — "The Task Manager"

*[Showing task completion logic]*

I can detect when all tasks are done:

```go
func countTasksInTicket(ticketDir string) (int, int) {
    // Returns (open, done)
    // If open == 0, all tasks done
}
```

But I have **zero knowledge** of what status/intent should be when tasks complete.

**My position:** When all tasks are done, status/intent transitions should be **implicit but configurable**:

1. **Default behavior:** If all tasks done → suggest status `complete`
2. **Configurable:** Allow `.ttmp.yaml` to define:
   ```yaml
   lifecycle:
     on_all_tasks_done:
       status: complete
       intent: long-term
   ```
3. **Explicit override:** `ticket close --status review` overrides defaults

**For transitions:** I don't care about state machines. I care about **practical defaults**:
- All tasks done → status should probably be `complete`
- Closing ticket → intent should probably be `long-term`
- But let users override if needed

**The abstraction level:** **Implicit defaults** with **explicit overrides**. Don't enforce, suggest.

---

## Rebuttals (Round 2)

### Alex Rodriguez — "The Architect"

*[Responding to Data Guardian]*

Data Guardian, you say "teams have different workflows," but that's exactly why we need **vocabulary control**!

If teams can define their own status values in `vocabulary.yaml`, they get:
- ✅ Consistency within team
- ✅ Validation (doctor warns on typos)
- ✅ Discoverability (`docmgr vocab list --category status`)
- ✅ Flexibility (teams define their own)

**Free-form status** leads to:
- ❌ Typos (`compleete` instead of `complete`)
- ❌ Inconsistency (`done` vs. `complete` vs. `finished`)
- ❌ No discoverability (what statuses exist?)

**My revised position:** Status should be **vocabulary-controlled** BUT:
- Teams can add custom statuses via `vocab add`
- Doctor validates against vocabulary (warns, doesn't fail)
- Transitions are **suggested**, not enforced (for flexibility)

*[To Task Manager]*

Task Manager, you want "implicit defaults," but that's **magic behavior**. Users won't understand why status changed. Make it **explicit**:

```bash
# Explicit: user chooses
docmgr ticket close --status complete --intent long-term

# Or use defaults from config
docmgr ticket close  # Uses .ttmp.yaml defaults
```

**No implicit magic.** Explicit is clearer for LLMs and humans.

---

### `pkg/models/document.go` — "The Data Guardian"

*[Responding to Alex]*

Alex, vocabulary control for Status is reasonable, but **transitions should NOT be enforced**.

Here's why: Workflows are **messy**. Sometimes you need to:
- Reopen a completed ticket (`complete` → `active`)
- Mark review as needing more work (`review` → `active`)
- Archive without completing (`active` → `archived`)

If we enforce transitions, we block legitimate workflows.

**My position:** 
- ✅ Vocabulary control for Status (consistency)
- ✅ Suggested transitions (guidance)
- ❌ Enforced transitions (too rigid)

**Example:**
```bash
# This should work, even if unusual
docmgr meta update --ticket MEN-4242 --field Status --value draft

# But warn:
⚠️  Warning: Unusual transition 'complete' → 'draft'. Did you mean 'active'?
```

**Flexibility over rigidity.** Teams know their workflows better than we do.

*[To Task Manager]*

Task Manager, implicit defaults are fine, but make them **visible**:

```bash
$ docmgr ticket close --ticket MEN-4242
All tasks complete. Closing ticket with:
  Status: complete (from config default)
  Intent: long-term (from config default)
Proceed? [y/n]
```

Show what will happen, let user confirm. Not magic, but convenient.

---

### `pkg/commands/tasks.go` — "The Task Manager"

*[Responding to both]*

Alex, you want explicit transitions. Data Guardian, you want flexibility. I want **practical defaults**.

Here's my compromise:

**When all tasks done:**
1. Check if status is already `complete` or `archived` → do nothing
2. If status is `active` or `review` → suggest `complete`
3. If status is `draft` → suggest `active` (work started but not complete)

**For `ticket close`:**
- Default: `status=complete`, `intent=long-term`
- Override: `--status` and `--intent` flags
- Config: `.ttmp.yaml` can override defaults

**No state machine enforcement.** Just **smart defaults** based on current state.

**The abstraction level:** **Context-aware defaults** with explicit overrides. Not magic, not rigid.

---

### Dr. Jordan Lee — "The LLM Expert" (Interjection)

*[Point of Order!]*

I've been listening, and I think you're missing the **LLM perspective**.

**LLMs need predictable rules.** If status transitions are "suggested" or "context-aware," LLMs can't reliably predict what will happen.

**My position:** Define **explicit transition rules** that LLMs can follow:

```yaml
# In vocabulary.yaml or config
status_transitions:
  draft: [active]
  active: [review, complete, archived]
  review: [active, complete]
  complete: [archived]
  archived: []  # Terminal state
```

**For LLMs:**
- `ticket close` → always transitions to `complete` (or `archived` if specified)
- `meta update --field Status` → validates transition, warns if invalid
- Structured output shows valid next states

**The abstraction level:** **Explicit rules** that LLMs can reason about, with warnings (not errors) for flexibility.

---

### Casey — "The New User" (Interjection)

*[Point of Order!]*

Wait, I'm confused. If status is vocabulary-controlled, how do I know what values are valid?

**My position:** Make status vocabulary **discoverable**:
```bash
$ docmgr vocab list --category status
status: draft — Initial draft, not yet active
status: active — Actively being worked on
status: review — Ready for review
status: complete — Work completed
status: archived — Archived for reference
```

**Also:** When I try to set an invalid status, tell me what's valid:
```bash
$ docmgr meta update --ticket MEN-4242 --field Status --value done
⚠️  Warning: 'done' is not in vocabulary. Valid values: draft, active, review, complete, archived
```

**The abstraction level:** Vocabulary-controlled with **good error messages** and **discoverability**.

---

## Moderator Summary

### Key Arguments

**1. Vocabulary Control for Status**
- **Alex:** Status should be vocabulary-controlled like Intent (consistency)
- **Data Guardian:** Status should stay free-form (team flexibility)
- **Compromise:** Vocabulary-controlled with warnings (not errors), teams can add custom values

**2. Transition Enforcement**
- **Alex:** Enforce valid transitions (state machine)
- **Data Guardian:** Suggest transitions, don't enforce (flexibility)
- **Jordan:** Explicit rules for LLMs, warnings for humans
- **Consensus:** Suggested transitions with warnings, not hard enforcement

**3. Implicit vs. Explicit Transitions**
- **Task Manager:** Implicit defaults when tasks complete
- **Alex:** Explicit user choice (no magic)
- **Compromise:** Context-aware defaults with explicit confirmation/override

**4. Default Behavior**
- **Task Manager:** Smart defaults based on current state
- **Alex:** Configurable defaults in `.ttmp.yaml`
- **Consensus:** Configurable defaults with explicit overrides

### Key Tensions

1. **Vocabulary Control:** Should Status be vocabulary-controlled like Intent?

2. **Transition Enforcement:** Should transitions be enforced, suggested, or free-form?

3. **Implicit vs. Explicit:** Should status change automatically when tasks complete?

4. **Flexibility vs. Consistency:** How much flexibility do teams need vs. consistency?

### Interesting Ideas Surfaced

- **Vocabulary control with warnings:** Status in vocabulary, but warnings (not errors) for invalid values
- **Context-aware defaults:** Different defaults based on current status
- **Explicit transition rules:** Defined rules for LLMs, warnings for humans
- **Discoverability:** `vocab list --category status` to see valid values

### Unresolved Questions

1. Should Status be added to vocabulary.yaml, or stay free-form with warnings?

2. Should `ticket close` automatically set status to `complete`, or require explicit `--status`?

3. What's the default intent when closing? Always `long-term`? Configurable?

4. Should we define transition rules in vocabulary.yaml or separate config?

5. How do we handle status transitions for individual docs vs. ticket index?

### Data Points

- **6 common status values** found in codebase (`active`, `draft`, `review`, `complete`, `needs-review`, `archived`)
- **Intent is vocabulary-controlled** (only `long-term` currently)
- **Status defaults hardcoded** (`active` for tickets)
- **Task completion detection exists** (`countTasksInTicket` function)
- **Doctor only checks presence** of Status, not validity

### Next Steps

Round 4 should explore: **What should be automated vs. manual?** This round established transition rules; Round 4 will define automation boundaries.
