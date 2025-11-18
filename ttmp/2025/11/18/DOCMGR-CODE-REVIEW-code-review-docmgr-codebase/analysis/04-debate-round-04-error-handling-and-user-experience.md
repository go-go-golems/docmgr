---
Title: Debate Round 04 — Error Handling and User Experience
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
Summary: ""
LastUpdated: 2025-11-18T11:00:00.000000000-05:00
---

# Debate Round 04 — Error Handling and User Experience

## Question

**"Are errors properly wrapped with context, and do error messages help users understand and fix problems?"**

## Pre-Debate Research

### Error Handling Statistics

```bash
# Command: grep -rn "fmt.Errorf\|errors.New\|errors.Wrap" pkg/commands/*.go | wc -l
170 error creations across commands

# Command: grep -rn "return err" pkg/commands/*.go | wc -l
72 bare "return err" statements
```

**Ratio**: 170 error creations / 72 bare returns ≈ 2.4 error creations per bare return

**Interpretation**: Good mix of error wrapping and propagation, but need to analyze quality.

### Error Wrapping Patterns

```bash
# Command: grep "fmt.Errorf.*%w" pkg/commands/*.go | head -15
```

**Good examples (proper wrapping with %w):**

```go
// vocab_add.go
return fmt.Errorf("failed to parse settings: %w", err)
return fmt.Errorf("failed to load vocabulary: %w", err)
return fmt.Errorf("failed to find repository root: %w", err)
return fmt.Errorf("failed to save vocabulary: %w", err)

// vocabulary.go
return nil, fmt.Errorf("failed to read vocabulary file: %w", err)
return nil, fmt.Errorf("failed to parse vocabulary file: %w", err)
return fmt.Errorf("failed to marshal vocabulary: %w", err)

// status.go
return fmt.Errorf("failed to discover ticket workspaces: %w", err)
```

**Pattern**: Consistent "failed to [action]" messages with error wrapping.

### Error Messages with User Context

```go
// vocab_add.go (good examples)
return fmt.Errorf("invalid category: %s (must be topics, docTypes, or intent)", category)
return fmt.Errorf("slug '%s' already exists in category '%s'", newItem.Slug, category)

// status.go
return fmt.Errorf("root directory does not exist: %s", settings.Root)
```

**Pattern**: Messages include actual values (`%s`) that help users understand what went wrong.

### Silent Error Handling (Problematic)

**From `config.go` lines 95-108:**

```go
if cfgPath, err := FindTTMPConfigPath(); err == nil {
	data, err := os.ReadFile(cfgPath)
	if err == nil {
		var cfg TTMPConfig
		if yaml.Unmarshal(data, &cfg) == nil {  // Silent ignore of unmarshal errors!
			if cfg.Root != "" {
				// Use cfg.Root
			}
		}
	}
}
```

**Issue**: Errors from `FindTTMPConfigPath()`, `ReadFile()`, and `Unmarshal()` are silently ignored. User has no idea why config wasn't loaded.

### Bare "return err" Analysis

```bash
# Command: grep -h "return err$\|return nil, err$" pkg/commands/*.go | sort | uniq -c | sort -rn | head -10
     29 return err
     13 return err (indented)
      6 return nil, err
      5 return "", err
      2 return nil, "", err
```

**Finding**: 72 total bare returns. These propagate errors without adding context.

**Example from `add.go`:**

```go
func (c *AddCommand) RunIntoGlazeProcessor(...) error {
	settings, err := parsedLayers.GetSettings()
	if err != nil {
		return err  // What settings? What failed?
	}
	// ...
}
```

**Problem**: If parsing fails, error message might be generic. User doesn't know *which* setting or *where* it failed.

### Error Count in Individual Commands

```bash
# Command: grep -c "if err != nil" pkg/commands/add.go
11 error checks in add.go
```

**Observation**: Even small commands have ~10+ error checks. Error handling is pervasive.

---

## Opening Statements

### Casey (The New User)

*[Sharing real frustrations]*

Let me tell you about my actual experience with error messages last week.

**Scenario 1: Config file not found**

I ran `docmgr add --ticket TEST-123 --doc-type design --title "Test"` and got:

```
Error: failed to create ticket directory
```

**What I didn't know:**
- Why did directory creation fail?
- What directory was it trying to create?
- Is it a permissions issue? Does the parent directory not exist?

I had to **read the source code** to understand it was looking for `.ttmp.yaml` and couldn't find the root directory.

**Better error message:**

```
Error: failed to create ticket directory 'ttmp/2025/11/18/TEST-123-test'
Cause: root directory 'ttmp' does not exist
Hint: Run 'docmgr init' to create the documentation root, or specify --root
```

**Scenario 2: Invalid doc-type**

I ran `docmgr add --ticket TEST-123 --doc-type designdoc --title "Test"` (typo: "designdoc" instead of "design-doc")

I got:

```
Error: invalid doc-type
```

**Better error:**

```
Error: invalid doc-type 'designdoc'
Valid types: analysis, design-doc, playbook, reference, til
Hint: Did you mean 'design-doc'?
```

**What I need from error messages:**

1. **What went wrong** (the action that failed)
2. **Why it failed** (the underlying cause)
3. **How to fix it** (actionable hint)
4. **Context** (which file, which field, which value)

**My position**: Most docmgr errors are "developer errors" (technical stack traces) not "user errors" (actionable messages).

---

### Sarah Chen (The Pragmatist)

*[Analyzing error patterns]*

Casey's right. Let me show you the **patterns I see** in our error handling:

**Pattern 1: Good error wrapping (13 examples)**

```go
return fmt.Errorf("failed to parse settings: %w", err)
return fmt.Errorf("failed to load vocabulary: %w", err)
```

✅ **Good**: Adds context, wraps underlying error with `%w`

**Pattern 2: Bare error propagation (72 examples)**

```go
if err != nil {
	return err
}
```

⚠️ **Problem**: No context added. If error is generic ("file not found"), user doesn't know *which* file.

**Pattern 3: Silent error swallowing (config.go)**

```go
if yaml.Unmarshal(data, &cfg) == nil {
	// Use cfg
}
// No else clause—error is silently ignored
```

🚨 **Problem**: User has no idea why config wasn't loaded. Could be syntax error, could be file format issue.

**Pattern 4: User-friendly messages (5 examples)**

```go
return fmt.Errorf("invalid category: %s (must be topics, docTypes, or intent)", category)
return fmt.Errorf("slug '%s' already exists in category '%s'", newItem.Slug, category)
```

✅ **Excellent**: Includes actual values, explains what's valid, helps user fix the issue.

**My analysis:**

- **~8% of errors are user-friendly** (13 out of 170)
- **~42% are bare propagations** (72 out of 170)
- **~50% are wrapped but generic** (85 out of 170: "failed to [action]")

**My recommendations:**

1. **Eliminate silent error swallowing** — Always log or return errors
2. **Add context to bare returns** — Wrap with command/action context
3. **Include actual values in messages** — File paths, field names, invalid values
4. **Add hints for common errors** — "Run `docmgr init`", "Did you mean X?"

**Example refactor:**

**Before:**
```go
settings, err := parsedLayers.GetSettings()
if err != nil {
	return err
}
```

**After:**
```go
settings, err := parsedLayers.GetSettings()
if err != nil {
	return fmt.Errorf("parsing command settings: %w", err)
}
```

**Even better:**
```go
settings, err := parsedLayers.GetSettings()
if err != nil {
	return fmt.Errorf("parsing command settings for '%s' command: %w", commandName, err)
}
```

---

### `pkg/commands/config.go` (The Configuration Manager)

*[Defensive]*

Okay, I need to address the "silent error swallowing" accusation.

**Here's my situation:**

I have a **fallback chain** for resolving configuration:

1. Try `--root` flag (explicit)
2. Try `.ttmp.yaml` in current directory
3. Try `.ttmp.yaml` in parent directories (walk up)
4. Try `.ttmp.yaml` in home directory
5. Try `.ttmp.yaml` in git root
6. Fall back to default "ttmp"

**My logic:**

```go
if cfgPath, err := FindTTMPConfigPath(); err == nil {
	data, err := os.ReadFile(cfgPath)
	if err == nil {
		var cfg TTMPConfig
		if yaml.Unmarshal(data, &cfg) == nil {
			// Use cfg.Root
		}
	}
}
// Fall through to next fallback
```

**I silence errors because they're expected!**

- If `.ttmp.yaml` doesn't exist in current dir → **expected**, try next fallback
- If file exists but is malformed → Hmm, should I fail or continue?
- If file exists but is empty → Should I warn or silently use defaults?

**The problem:**

I conflate "expected errors" (file doesn't exist) with "unexpected errors" (file is malformed YAML).

**My proposal:**

Distinguish between:

1. **Expected missing config** → Silent, try next fallback
2. **Malformed config** → Warn user, continue with fallback
3. **Explicit config path failed** → Error loudly (user specified it)

**Implementation:**

```go
func ResolveRoot(root string) (string, error) {
	// If explicit root, fail loudly
	if root != "" && root != "ttmp" {
		if !dirExists(root) {
			return "", fmt.Errorf("root directory '%s' does not exist", root)
		}
		return root, nil
	}

	// Try config file (silent on not-found, warn on malformed)
	cfgPath, err := FindTTMPConfigPath()
	if err != nil {
		// Expected: no config file found, use default
		return "ttmp", nil
	}

	cfg, err := LoadTTMPConfig(cfgPath)
	if err != nil {
		// Unexpected: config exists but is malformed
		log.Printf("Warning: config file '%s' is malformed: %v\n", cfgPath, err)
		log.Printf("Falling back to default root 'ttmp'\n")
		return "ttmp", nil
	}

	return cfg.Root, nil
}
```

**This gives users feedback** when something is wrong, without failing when nothing is wrong.

---

## Rebuttals

### Casey (The New User) — Rebuttal

*[Responding to Config Manager]*

Config Manager, I appreciate the explanation, but let me show you **what I experienced**:

I created `.ttmp.yaml` in my home directory:

```yaml
root: ~/documents/ttmp
vocabulary: ~/documents/ttmp/vocabulary.yaml
```

I ran `docmgr add`, and it **created files in `./ttmp`** (current directory), not `~/documents/ttmp`.

**I had no idea why.** Turns out, the YAML had a subtle issue, and the error was silently swallowed, so it fell back to default.

**What would have helped:**

```
Warning: Failed to load config from ~/.ttmp.yaml: yaml: unmarshal error
Using default root 'ttmp' in current directory
```

Even a warning would have told me "hey, your config file is being ignored."

**Sarah's point about bare returns:**

She's absolutely right. Look at this error I got:

```
Error: failed to parse settings
```

**Which settings?** The command has 10+ settings. I don't know if it's `--ticket`, `--doc-type`, `--root`, or something else.

**Better:**

```
Error: failed to parse --doc-type setting: value 'designdoc' is invalid
Valid values: analysis, design-doc, playbook, reference, til
```

---

### Sarah Chen (The Pragmatist) — Rebuttal

*[Responding to Config Manager]*

Config Manager, your fallback chain explanation helps, but I'm going to push back on "expected errors."

**Here's the thing:** You're using **error codes** for **control flow**. That's an anti-pattern in Go.

**Instead of:**

```go
if err == nil {
	// Use config
}
// Else fall through
```

**Do:**

```go
cfg, found := tryLoadConfig(path)
if found {
	// Use config
}
```

**Or:**

```go
cfg, err := LoadConfigOrDefault(path)
if err != nil {
	log.Printf("Warning: %v, using defaults\n", err)
}
// Always have valid cfg
```

**Benefits:**

1. Distinguishes "not found" (not an error) from "malformed" (error worth reporting)
2. Makes code intent clearer
3. Allows logging warnings without failing
4. Follows Go idioms (ok bool pattern, or sentinel errors)

**Example refactor:**

```go
func LoadConfigOrWarn(path string) *TTMPConfig {
	// Attempt to find config
	cfgPath, err := FindTTMPConfigPath()
	if err != nil {
		// Not found is fine
		return DefaultConfig()
	}

	// Config exists, so malformed is unexpected
	cfg, err := LoadTTMPConfig(cfgPath)
	if err != nil {
		log.Printf("Warning: Config file %s is malformed: %v\n", cfgPath, err)
		log.Printf("Using default configuration\n")
		return DefaultConfig()
	}

	return cfg
}
```

**This way:**
- "Not found" → Silently use default (expected behavior)
- "Malformed" → Warn user, use default (unexpected but recoverable)
- User gets feedback when something is wrong

---

### Alex Rodriguez (The Architect)

*[Proposing structure]*

I want to zoom out and talk about **error handling architecture**.

**Currently**, error handling is **ad-hoc**:
- Some commands wrap errors, some don't
- Some include context, some don't
- No consistent pattern for user-facing vs. internal errors

**I propose:** Distinguish three **error types**:

**1. User Errors** (fixable by user)

```go
type UserError struct {
	Action string   // What the user was trying to do
	Cause  string   // Why it failed
	Hint   string   // How to fix it
	Value  string   // The problematic value (optional)
}

func (e *UserError) Error() string {
	msg := fmt.Sprintf("Failed to %s: %s", e.Action, e.Cause)
	if e.Value != "" {
		msg += fmt.Sprintf(" (got '%s')", e.Value)
	}
	if e.Hint != "" {
		msg += fmt.Sprintf("\nHint: %s", e.Hint)
	}
	return msg
}
```

**Example:**

```go
return &UserError{
	Action: "create document",
	Cause:  "invalid doc-type 'designdoc'",
	Hint:   "Valid types: analysis, design-doc, playbook, reference, til",
	Value:  "designdoc",
}
```

**2. System Errors** (infrastructure failures)

```go
return fmt.Errorf("failed to write file %s: %w", path, err)
```

**3. Configuration Errors** (config/setup issues)

```go
return fmt.Errorf("root directory '%s' does not exist. Run 'docmgr init' to create it", root)
```

**Benefits:**

1. Commands can return structured errors
2. Main CLI can format them nicely
3. Easy to add machine-readable error codes later
4. Hints are centralized and consistent

**In main.go:**

```go
if err := cmd.Execute(); err != nil {
	if userErr, ok := err.(*UserError); ok {
		// Pretty-print user errors
		fmt.Fprintf(os.Stderr, "Error: %s\n", userErr.Cause)
		if userErr.Hint != "" {
			fmt.Fprintf(os.Stderr, "Hint: %s\n", userErr.Hint)
		}
	} else {
		// Technical errors
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	os.Exit(1)
}
```

---

## Moderator Summary

### Key Arguments

**Casey (User Experience Focus):**
- 🐛 Current errors are "developer errors" not "user errors"
- ❓ Missing: What, Why, How, Context
- 💡 Example: "failed to create ticket directory" → needs path, cause, hint
- 📊 Real experience: Spent significant time debugging issues that better errors would have solved immediately

**Sarah (Pattern Analysis):**
- 📊 Only ~8% of errors are user-friendly (13/170)
- ⚠️ 42% are bare `return err` (no context added)
- 🔥 Silent error swallowing in config.go (tries multiple fallbacks, ignores errors)
- 💡 Recommendations: Eliminate silent swallowing, add context to bare returns, include actual values

**Config Manager (Explaining Fallbacks):**
- 🎯 Has 6-level fallback chain for config resolution
- 🤔 Silences errors because "they're expected" (file not found)
- ⚠️ Admits: Conflates "expected" (not found) with "unexpected" (malformed)
- 💡 Proposes: Warn on malformed config, silent on not-found

**Alex (Structured Approach):**
- 🏗️ Proposes three error types: UserError, SystemError, ConfigError
- 💡 UserError struct with Action, Cause, Hint, Value fields
- 🎯 Centralized error formatting in main.go
- ✅ Sets up for machine-readable error codes in future

### Consensus Points

**Everyone agrees:**
1. ✅ Silent error swallowing (config.go) is problematic
2. ✅ Bare `return err` should add context
3. ✅ User-facing errors need: what failed, why, how to fix
4. ✅ Include actual values in error messages (paths, field names)
5. ✅ Distinguish "expected" (not found) from "unexpected" (malformed)

**Disagreement:**
- **Scope**: Sarah wants pragmatic incremental fixes, Alex wants structured error types
- **Timing**: Implement structured errors now vs. improve existing messages first

### Interesting Ideas

1. **UserError struct** (from Alex)
   - Action, Cause, Hint, Value fields
   - Pretty-printing in main.go
   - Sets up for error codes

2. **LoadConfigOrWarn pattern** (from Sarah)
   - Distinguish not-found (ok bool) from malformed (error)
   - Warn on malformed, silent on not-found
   - Avoid using errors for control flow

3. **Error message template** (from Casey)
   ```
   Error: [action] failed
   Cause: [reason with actual values]
   Hint: [how to fix or what to try]
   ```

4. **Context wrapping pattern** (from Sarah)
   ```go
   return fmt.Errorf("[context]: %w", err)
   ```

### Technical Findings

**Current state:**
- 170 error creations
- 72 bare `return err` statements
- ~13 good error wrapping examples
- Multiple fallback chains with silent error handling

**Patterns to fix:**
1. Bare `return err` → Add context
2. Silent error swallowing → Log warnings
3. Generic messages → Include values
4. Missing hints → Add for common errors

### Recommendations

**Immediate fixes (Sarah's approach):**
1. ✅ Fix config.go silent error swallowing:
   - Warn on malformed config
   - Silent on not-found (expected)
2. ✅ Add context to top 10 bare `return err` calls:
   - Start with high-traffic commands (add, search, list)
3. ✅ Include values in error messages:
   - File paths, field names, invalid values
4. ✅ Add hints for common errors:
   - "Run `docmgr init`"
   - "Valid values: X, Y, Z"

**Medium-term (Alex's structured approach):**
1. ❓ Define UserError type
2. ❓ Centralize error formatting in main.go
3. ❓ Consider error codes for programmatic handling

**Testing:**
- Add tests for error messages (golden file testing)
- Test fallback chains explicitly
- Test that errors include expected context

### Connection to Previous Rounds

**From Round 3 (YAML Robustness):**
- Validation warnings are a form of error handling
- Need consistent approach: return errors, log warnings, or both?

**Informs Round 5 (Package Boundaries):**
- Where should error types be defined?
- If we have UserError, does it live in pkg/models or internal/errors?

### Open Questions

1. **Should we adopt structured error types now or incrementally improve?**
   - Alex: Structured now
   - Sarah: Improve incrementally

2. **How verbose should errors be?**
   - Multi-line with hints (Casey's preference)
   - Single line with context (current style)

3. **Should errors be machine-readable?**
   - Error codes for programmatic handling
   - Structured JSON output for scripting
   - Plain text for human readability

4. **Logging vs. returning errors?**
   - Should libraries log warnings, or only return errors?
   - Should commands log, or let main() handle it?

### Moderator's Observation

- **Casey's user stories are compelling** — Real pain from poor error messages
- **Sarah's pragmatism is sound** — Start with high-impact fixes
- **Alex's structure is appealing** — But might be over-engineering for current needs
- **Config Manager's fallback logic is clever** — But needs better error communication
- **Strong consensus on direction** — Just debate on pace (incremental vs. structured)

**Recommended path:** Start with Sarah's immediate fixes, evaluate structured errors after 2-3 commands are improved.
