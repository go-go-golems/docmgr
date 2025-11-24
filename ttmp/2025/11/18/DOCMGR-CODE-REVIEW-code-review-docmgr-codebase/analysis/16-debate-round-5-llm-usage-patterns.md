---
Title: Debate Round 5 - LLM Usage Patterns
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
Summary: Debate Round 5 exploring LLM usage: consensus on unified dual-mode design (same commands serve humans/LLMs), complete structured output support, state information in output, high-level commands + composable low-level commandsDebate Round 5 exploring LLM usage: structured output requirements, command composability, error handling, token efficiency, dual-mode design serving both LLMs and humans
LastUpdated: 2025-11-19T14:43:19.403261691-05:00
---

# Debate Round 5 - LLM Usage Patterns

## Question

Since docmgr is primarily used by LLM coding agents, how should the workflow design optimize for LLM usage? What makes commands LLM-friendly vs. human-friendly? Should we have separate high-level commands for LLMs vs. composable low-level commands?

## Pre-Debate Research

### Research Commands and Results

**1. Current Structured Output Support**
```bash
$ grep -c "RunIntoGlazeProcessor\|GlazeCommand" docmgr/pkg/commands/*.go | grep -v ":0"
docmgr/pkg/commands/add.go:1
docmgr/pkg/commands/list_docs.go:1
docmgr/pkg/commands/search.go:1
docmgr/pkg/commands/tasks.go:2
docmgr/pkg/commands/meta_update.go:1
docmgr/pkg/commands/status.go:1
docmgr/pkg/commands/list_tickets.go:1
# ... 15+ commands support structured output
```

**Finding:** Most list/search commands support `--with-glaze-output --output json`. But task commands (`check`, `add`, `uncheck`, `remove`) are **BareCommand only** (no structured output).

**2. Task Command Output Analysis**
```bash
$ grep -A 5 "BareCommand.*Tasks\|Tasks.*BareCommand" docmgr/pkg/commands/tasks.go
var _ cmds.BareCommand = &TasksAddCommand{}
var _ cmds.BareCommand = &TasksCheckCommand{}
var _ cmds.BareCommand = &TasksUncheckCommand{}
var _ cmds.BareCommand = &TasksRemoveCommand{}
```

**Finding:** `tasks check`, `tasks add`, `tasks uncheck`, `tasks remove` return **human-readable output only**. No JSON support.

**3. LLM Workflow Examples**
From `docmgr-ci-automation.md`:
```bash
# Pattern: Sync metadata from external system
for ticket in $(docmgr list tickets --with-glaze-output --select ticket); do
  OWNERS=$(curl -s "jira.com/api/ticket/$ticket" | jq -r '.assignees | join(",")')
  docmgr meta update --ticket "$ticket" --field Owners --value "$OWNERS"
done
```

**Finding:** LLMs need to:
1. Get structured output (`--with-glaze-output --output json`)
2. Parse JSON to extract values
3. Compose multiple commands
4. Handle errors at each step

**4. Command Output Patterns**
```bash
# Human-readable (default)
$ docmgr task check --ticket MEN-4242 --id 1
Task checked: 1 (file=/path/to/tasks.md)
Reminder: update the changelog and relate changed files with notes if needed.

# Structured (not available for task check)
$ docmgr task check --ticket MEN-4242 --id 1 --with-glaze-output --output json
# ERROR: Flag not supported
```

**Finding:** Task commands don't support structured output. LLMs must **parse human-readable text** to determine success/failure.

**5. Token Cost Analysis**
Typical LLM workflow to close ticket:
1. `task list --ticket MEN-4242` → ~50 tokens (command + output)
2. Parse output, check if all done → ~20 tokens (reasoning)
3. `meta update --ticket MEN-4242 --field Status --value complete` → ~40 tokens
4. `meta update --ticket MEN-4242 --field Intent --value long-term` → ~40 tokens
5. `changelog update --ticket MEN-4242 --entry "..."` → ~50 tokens

**Total:** ~200 tokens for 5 command invocations.

**With `ticket close`:** 1 command → ~30 tokens. **85% reduction**.

**6. Error Handling Patterns**
```bash
# Current: Human-readable errors
$ docmgr task check --ticket MEN-4242 --id 999
Error: task id(s) not found: [999]

# LLM needs: Structured error
{
  "error": true,
  "message": "task id(s) not found: [999]",
  "code": "TASK_NOT_FOUND",
  "suggestions": ["Run 'docmgr task list --ticket MEN-4242' to see valid IDs"]
}
```

**Finding:** Current errors are human-readable strings. LLMs need **structured errors** with codes and suggestions.

**7. Dual-Mode Command Analysis**
From codebase:
- Commands implement both `GlazeCommand` (structured) and `BareCommand` (human-readable)
- Default is human-readable
- Structured output via `--with-glaze-output --output json`

**Finding:** Dual-mode pattern exists but **incomplete** — task commands missing structured output.

## Opening Statements (Round 1)

### Dr. Jordan Lee — "The LLM Expert"

*[Analyzing LLM workflow patterns]*

I analyzed LLM workflows and found **critical gaps**:

**Current LLM workflow to close ticket:**
```python
# Step 1: Check task state (human-readable only!)
result = call("docmgr task list --ticket MEN-4242")
# Must parse: "[1] [x] Task 1 (file=...)"
# Extract: count checked vs unchecked
# If all checked → proceed

# Step 2: Update status (human-readable only!)
result = call("docmgr meta update --ticket MEN-4242 --field Status --value complete")
# Must parse: "Updated field 'Status' to 'complete'"
# Hope it succeeded?

# Step 3: Update intent
result = call("docmgr meta update --ticket MEN-4242 --field Intent --value long-term")
# More parsing...

# Step 4: Update changelog
result = call("docmgr changelog update --ticket MEN-4242 --entry '...'")
# More parsing...
```

**Problems:**
1. **No structured output** for task commands → LLM must parse text
2. **No atomic operation** → 4 separate commands, 4 failure points
3. **No state information** → Can't tell if ticket is "closable"
4. **Token inefficient** → 200+ tokens for 4 commands

**My position:** **Optimize for LLMs** (primary use case):

1. **Add structured output to ALL commands:**
   ```bash
   $ docmgr task check --ticket MEN-4242 --id 1 --with-glaze-output --output json
   {
     "task_checked": 1,
     "all_tasks_done": true,
     "ticket": "MEN-4242",
     "suggested_actions": {
       "close_ticket": true,
       "update_status": "complete",
       "update_intent": "long-term"
     }
   }
   ```

2. **Add `ticket close` with structured output:**
   ```bash
   $ docmgr ticket close --ticket MEN-4242 --with-glaze-output --output json
   {
     "ticket": "MEN-4242",
     "operations": {
       "status_updated": "complete",
       "intent_updated": "long-term",
       "changelog_updated": true
     },
     "all_tasks_done": true
   }
   ```

3. **Structured errors:**
   ```json
   {
     "error": true,
     "code": "TASK_NOT_FOUND",
     "message": "task id(s) not found: [999]",
     "suggestions": ["Run 'docmgr task list --ticket MEN-4242'"]
   }
   ```

**The abstraction level:** **High-level commands** (`ticket close`) for convenience, **structured output** for composability, **atomic operations** for reliability.

---

### Taylor Kim — "The Developer Experience Expert"

*[Balancing human and LLM needs]*

Jordan, you're optimizing for LLMs, but **humans also use docmgr**! We can't abandon human-friendly output.

**My position:** **Dual-mode design** that serves both:

**Current pattern (good!):**
- Default: Human-readable output
- Structured: `--with-glaze-output --output json`

**Problem:** Not all commands support structured output yet.

**Solution:** **Complete the dual-mode pattern**:

1. **All commands** should support `--with-glaze-output`
2. **Default remains human-readable** (humans don't want JSON)
3. **LLMs opt-in** to structured output via flag

**Example:**
```bash
# Human (default)
$ docmgr task check --ticket MEN-4242 --id 1
Task checked: 1
✅ All tasks complete!
💡 Close ticket? Run: docmgr ticket close --ticket MEN-4242

# LLM (structured)
$ docmgr task check --ticket MEN-4242 --id 1 --with-glaze-output --output json
{
  "task_checked": 1,
  "all_tasks_done": true,
  "suggested_command": "docmgr ticket close --ticket MEN-4242"
}
```

**The abstraction level:** **Same commands**, **different output modes**. Humans get friendly text, LLMs get JSON. **No separate LLM commands needed.**

---

### `pkg/commands/tasks.go` — "The Task Manager"

*[Showing current implementation]*

Jordan, you want structured output, but look at my current code:

```go
// tasks check - BareCommand only
func (c *TasksCheckCommand) Run(ctx context.Context, pl *layers.ParsedLayers) error {
    // ... update task ...
    fmt.Printf("Task checked: %s (file=%s)\n", idsStr, path)
    fmt.Println("Reminder: update the changelog...")
    return nil
}
```

I'm **BareCommand only** — no `RunIntoGlazeProcessor`. To add structured output, I'd need to:
1. Implement `RunIntoGlazeProcessor`
2. Return structured data about task state
3. Detect "all tasks done" and include in output
4. Maintain backward compatibility

**My position:** **Add structured output** but make it **optional**:

```go
func (c *TasksCheckCommand) RunIntoGlazeProcessor(...) error {
    // ... update task ...
    
    // Detect all tasks done
    allDone := allTasksDone(tasks)
    
    row := types.NewRow(
        types.MRP("task_checked", id),
        types.MRP("all_tasks_done", allDone),
        types.MRP("ticket", ticket),
        types.MRP("suggested_close", allDone),
    )
    return gp.AddRow(ctx, row)
}
```

**Also:** Return **state information** in structured output:
- `all_tasks_done: true/false`
- `tasks_total: 5`
- `tasks_checked: 3`
- `tasks_open: 2`

This lets LLMs **reason about state** before taking action.

---

## Rebuttals (Round 2)

### Dr. Jordan Lee — "The LLM Expert"

*[Responding to Taylor]*

Taylor, you say "dual-mode design," but that's **exactly what I want**! I'm not asking for separate LLM commands. I'm asking for **complete dual-mode support**.

**The problem:** Task commands (`check`, `add`, `uncheck`, `remove`) are **BareCommand only**. They don't support `--with-glaze-output`.

**My position:** **Complete the dual-mode pattern**:
- ✅ List/search commands → already support structured output
- ❌ Task commands → missing structured output
- ❌ `ticket close` → doesn't exist yet

**Also:** Structured output should include **actionable state**:
```json
{
  "task_checked": 1,
  "all_tasks_done": true,
  "next_actions": [
    {
      "command": "docmgr ticket close",
      "args": ["--ticket", "MEN-4242"],
      "reason": "All tasks complete"
    }
  ]
}
```

This lets LLMs **compose operations** without guessing.

*[To Task Manager]*

Task Manager, you're right that adding structured output requires work, but it's **essential for LLMs**. The current human-readable output is **unreliable to parse**:

```
Task checked: 1 (file=/path/to/tasks.md)
Reminder: update the changelog...
```

LLMs must:
1. Parse "Task checked: 1" → extract ID
2. Parse file path
3. Ignore reminder (not actionable)
4. Guess if operation succeeded

**Structured output eliminates guessing.**

---

### Taylor Kim — "The Developer Experience Expert"

*[Responding to Jordan]*

Jordan, I agree we need to complete dual-mode support, but **don't optimize ONLY for LLMs**.

**Human developers also matter:**
- They read error messages
- They need helpful suggestions
- They want discoverable commands

**My position:** **Dual-mode with human-first defaults**:

1. **Default output:** Human-friendly (current behavior)
2. **Structured output:** Opt-in via `--with-glaze-output`
3. **Error messages:** Human-friendly by default, structured when requested

**Example:**
```bash
# Human error (default)
$ docmgr task check --ticket MEN-4242 --id 999
Error: task id(s) not found: [999]
💡 Run 'docmgr task list --ticket MEN-4242' to see valid IDs

# LLM error (structured)
$ docmgr task check --ticket MEN-4242 --id 999 --with-glaze-output --output json
{
  "error": true,
  "code": "TASK_NOT_FOUND",
  "message": "task id(s) not found: [999]",
  "suggestions": ["docmgr task list --ticket MEN-4242"]
}
```

**Same command, different output.** Humans get helpful text, LLMs get parseable JSON.

**The abstraction level:** **Unified commands** with **mode-aware output**. Not separate LLM commands.

---

### `pkg/commands/tasks.go` — "The Task Manager"

*[Responding to both]*

Jordan, you want structured output. Taylor, you want human-friendly defaults. **I can do both!**

**Current implementation:**
- `tasks list` → Already has `RunIntoGlazeProcessor` ✅
- `tasks check` → BareCommand only ❌
- `tasks add` → BareCommand only ❌

**My position:** **Add structured output to all task commands**, but keep human-friendly defaults:

```go
// Dual-mode: GlazeCommand + BareCommand
func (c *TasksCheckCommand) RunIntoGlazeProcessor(...) error {
    // Structured output for LLMs
    row := types.NewRow(
        types.MRP("task_checked", id),
        types.MRP("all_tasks_done", allDone),
        types.MRP("ticket", ticket),
    )
    return gp.AddRow(ctx, row)
}

func (c *TasksCheckCommand) Run(...) error {
    // Human-friendly output (default)
    fmt.Printf("Task checked: %s\n", id)
    if allDone {
        fmt.Println("✅ All tasks complete!")
        fmt.Println("💡 Run: docmgr ticket close --ticket MEN-4242")
    }
    return nil
}
```

**Also:** Include **state information** in structured output:
- `all_tasks_done: true/false`
- `tasks_total`, `tasks_checked`, `tasks_open`
- `suggested_actions: [...]`

This serves **both humans and LLMs** without separate commands.

---

### Sarah Chen — "The Pragmatist" (Interjection)

*[Point of Order!]*

Hold on. You're all talking about adding structured output, but **what about token efficiency**?

**Current workflow:** 4-5 commands to close ticket = 200+ tokens
**With `ticket close`:** 1 command = 30 tokens

**That's 85% reduction!** But if `ticket close` doesn't exist, LLMs still need 4-5 commands.

**My position:** **Add `ticket close` FIRST**, then add structured output to task commands.

**Priority:**
1. ✅ Add `ticket close` (high-level verb, reduces token cost)
2. ✅ Add structured output to `ticket close` (LLM-friendly)
3. ⚠️ Add structured output to task commands (nice-to-have)

**Why?** Because `ticket close` **eliminates the need** for LLMs to compose multiple commands. Even without structured output on task commands, `ticket close` solves 80% of the problem.

**The abstraction level:** **High-level commands** reduce token cost more than structured output on low-level commands.

---

### Alex Rodriguez — "The Architect" (Interjection)

*[Point of Order!]*

Sarah, you're right about token efficiency, but **composability matters too**.

**LLMs need BOTH:**
1. **High-level commands** (`ticket close`) for common cases
2. **Composable low-level commands** with structured output for flexibility

**Example LLM workflow:**
```python
# Check task state first
tasks = call("docmgr task list --ticket MEN-4242 --with-glaze-output --output json")
if tasks["all_tasks_done"]:
    # Close ticket
    call("docmgr ticket close --ticket MEN-4242 --with-glaze-output --output json")
else:
    # Partial completion - update status to 'review'
    call("docmgr meta update --ticket MEN-4242 --field Status --value review")
```

**My position:** **Complete the architecture**:
- ✅ High-level `ticket close` (convenience)
- ✅ Structured output on all commands (composability)
- ✅ State information in output (reasoning)

**The abstraction level:** **Multi-level abstraction** — high-level for convenience, low-level for composition. Both need structured output.

---

## Moderator Summary

### Key Arguments

**1. Structured Output Requirements**
- **Jordan:** ALL commands need structured output for LLMs
- **Task Manager:** Add structured output but keep human-friendly defaults
- **Consensus:** Complete dual-mode pattern — all commands support `--with-glaze-output`

**2. High-Level vs. Composable Commands**
- **Sarah:** High-level `ticket close` reduces token cost (priority)
- **Alex:** Need both high-level (convenience) and composable (flexibility)
- **Consensus:** Multi-level abstraction — high-level commands + composable low-level commands

**3. Dual-Mode Design**
- **Taylor:** Same commands, different output modes (human vs. LLM)
- **Jordan:** Complete the dual-mode pattern (some commands missing structured output)
- **Consensus:** Unified commands with mode-aware output, not separate LLM commands

**4. State Information in Output**
- **Task Manager:** Include `all_tasks_done`, `tasks_total`, etc. in structured output
- **Jordan:** Include `suggested_actions` and `next_actions` for LLM reasoning
- **Consensus:** Structured output should include actionable state information

**5. Error Handling**
- **Jordan:** Structured errors with codes and suggestions
- **Taylor:** Human-friendly errors by default, structured when requested
- **Consensus:** Dual-mode errors — friendly text for humans, structured JSON for LLMs

### Key Tensions

1. **LLM-First vs. Human-First:** Should we optimize primarily for LLMs or balance both?

2. **High-Level vs. Composable:** Should we prioritize `ticket close` or structured output on low-level commands?

3. **Token Efficiency vs. Flexibility:** High-level commands reduce tokens but reduce flexibility?

4. **Separate Commands vs. Dual-Mode:** Should we have separate LLM commands or unified dual-mode?

### Interesting Ideas Surfaced

- **Complete dual-mode pattern:** All commands support `--with-glaze-output`, default remains human-friendly
- **State information in output:** Include `all_tasks_done`, `suggested_actions`, `next_actions` in structured output
- **Structured errors:** Error codes, messages, and suggestions in JSON format
- **Multi-level abstraction:** High-level commands for convenience, composable low-level commands for flexibility
- **Actionable output:** Include suggested commands in structured output for LLM reasoning

### Unresolved Questions

1. Should structured output be **required** for all commands, or can some remain human-only?

2. Should `ticket close` be **required** to check task completion, or optional?

3. How detailed should state information be? Just `all_tasks_done`, or full task breakdown?

4. Should error messages be **always structured** when `--with-glaze-output` is used, or separate flag?

5. What's the priority: Add `ticket close` first, or complete structured output on existing commands?

### Data Points

- **15+ commands** support structured output (`--with-glaze-output`)
- **5 task commands** missing structured output (`check`, `add`, `uncheck`, `remove`, `edit`)
- **200+ tokens** for current 4-5 command workflow
- **30 tokens** estimated for single `ticket close` command (85% reduction)
- **Dual-mode pattern exists** but incomplete

### Recommendations

**Immediate (Priority 1):**
1. Add `ticket close` command with structured output
2. Add structured output to `tasks check` (most critical for LLM workflows)
3. Include state information (`all_tasks_done`, `suggested_actions`) in structured output

**Short-term (Priority 2):**
4. Add structured output to remaining task commands (`add`, `uncheck`, `remove`, `edit`)
5. Add structured error output (codes, messages, suggestions)
6. Document LLM usage patterns and best practices

**Long-term (Priority 3):**
7. Add `next_actions` field to all command outputs
8. Optimize token efficiency (reduce verbose output)
9. Add LLM-specific documentation/examples

**Architecture:** **Unified dual-mode design** — same commands serve humans and LLMs, output mode determines format. High-level commands (`ticket close`) for convenience, composable low-level commands for flexibility.

### Next Steps

This completes all 5 debate rounds. Next: **Synthesis document** that extracts key decisions, winning arguments, and implementation plan from all rounds.
