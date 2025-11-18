---
Title: Debate Round 08 — Configuration and Path Resolution Design
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
Summary: "6-level fallback chain is clever but opaque. Consensus: Add --verbose flag to show resolution path, warn on malformed config (from Round 4), document fallback order"
LastUpdated: 2025-11-18T11:40:00.000000000-05:00
---

# Debate Round 08 — Configuration and Path Resolution Design

## Question

**"Is the configuration system (WorkspaceConfig, path resolution, root discovery) well-designed and understandable?"**

## Pre-Debate Research

### Configuration Sources

```go
// From pkg/commands/config.go analysis:
// 6-level fallback chain for resolving root directory:
1. --root flag (explicit command-line argument)
2. .ttmp.yaml in current directory
3. .ttmp.yaml in parent directories (walk up tree)
4. .ttmp.yaml in user's home directory
5. .ttmp.yaml in git repository root
6. Default value "ttmp"
```

### Configuration Structure

```go
type TTMPConfig struct {  // Note: Renamed to WorkspaceConfig in Round 6
	Root       string `yaml:"root"`
	Vocabulary string `yaml:"vocabulary"`
}
```

**Fields:**
- `Root`: Base directory for ticket workspaces
- `Vocabulary`: Path to vocabulary.yaml file

### Path Resolution Complexity

```bash
# From Round 4 research: config.go lines 80-109
func ResolveRoot(root string) string {
	// If explicit non-default, use it
	if root != "ttmp" && root != "" {
		if filepath.IsAbs(root) {
			return root
		}
		// Relative to cwd
		if cwd, err := os.Getwd(); err == nil {
			return filepath.Join(cwd, root)
		}
		return root
	}

	// Try config file (silent on errors - see Round 4 issue)
	if cfgPath, err := FindTTMPConfigPath(); err == nil {
		// ... load config, silently ignore errors
	}

	// ... more fallbacks
	return "ttmp"  // Ultimate default
}
```

**Issues identified in Round 4:**
- Silent error swallowing
- No visibility into which fallback succeeded
- Users don't know why config wasn't loaded

---

## Opening Statements

### `pkg/commands/config.go` (The Configuration Manager)

*[Defensive but open to feedback]*

Let me explain **why I have a 6-level fallback chain**:

**User scenarios I'm supporting:**

**Scenario 1: Project-specific workspace**

User creates `.ttmp.yaml` in project root:

```yaml
root: docs/workspace
vocabulary: docs/vocabulary.yaml
```

Commands run from any subdirectory find this config (walk up tree).

**Scenario 2: Global workspace**

User puts `.ttmp.yaml` in `~/.ttmp.yaml`:

```yaml
root: ~/Documents/documentation
vocabulary: ~/Documents/documentation/vocabulary.yaml
```

All projects use the same workspace unless overridden.

**Scenario 3: Git-aware**

For monorepos, find `.ttmp.yaml` at git root, not just current directory.

**Scenario 4: Explicit override**

`docmgr add --root /tmp/test-workspace ...` bypasses all config.

**Scenario 5: Default behavior**

No config? Use "ttmp" in current directory (just works™).

**The problem:** This flexibility creates **opacity**. Users don't know which fallback was used.

**My proposal (informed by Round 4):**

**1. Add visibility flag:**

```bash
$ docmgr add --verbose --ticket TEST-123 ...
[docmgr] Trying --root flag: not set
[docmgr] Trying .ttmp.yaml in /home/user/project: not found
[docmgr] Trying .ttmp.yaml in /home/user: not found
[docmgr] Trying .ttmp.yaml in git root: found at /home/user/project/.ttmp.yaml
[docmgr] Using root: /home/user/project/docs
```

**2. Warn on malformed config (from Round 4 consensus):**

```bash
[docmgr] Warning: Config file ~/.ttmp.yaml is malformed: yaml: line 3: did not find expected key
[docmgr] Falling back to default root: ttmp
```

**3. Add `config show` command:**

```bash
$ docmgr config show
Configuration sources (in precedence order):
  1. --root flag: <not set>
  2. .ttmp.yaml (current dir): <not found>
  3. .ttmp.yaml (parent dirs): <not found>
  4. .ttmp.yaml (home dir): found at ~/.ttmp.yaml
  5. .ttmp.yaml (git root): <not applicable, not in git repo>
  6. Default: ttmp

Active configuration:
  root: ~/Documents/documentation
  vocabulary: ~/Documents/documentation/vocabulary.yaml
  source: ~/.ttmp.yaml
```

This gives users **full transparency**.

---

### Casey (The New User)

*[Sharing confusion]*

I experienced **exactly the opacity** Config Manager describes.

**What happened:**

I created `.docmgr.yaml` (note: I used the name from Round 6 renaming discussion):

```yaml
root: ~/docs
vocabulary: ~/docs/vocab.yaml
```

Commands ignored it. I didn't know why.

Eventually discovered it needs to be `.ttmp.yaml`, not `.docmgr.yaml`.

**But:** No error message told me "config file not found" or "wrong file name". It just silently used default.

**Config Manager's proposals would have helped:**

- `--verbose` would show "trying .docmgr.yaml: wrong name"
- `docmgr config show` would show "no config found"
- Warning on unexpected files

**Additional request: `docmgr init` should create config file**

```bash
$ docmgr init
Created workspace root: ./ttmp
Created config file: .ttmp.yaml
  root: ttmp
  vocabulary: ttmp/vocabulary.yaml

Hint: You can customize these paths by editing .ttmp.yaml
```

Right now, `docmgr init` just creates the directory. It should **also create the config file** so users know it exists.

---

### Alex Rodriguez (The Architect)

*[Analyzing the design]*

Let me evaluate the **fallback chain design** from an architecture perspective.

**Good aspects:**

1. ✅ **Explicit beats implicit** — `--root` flag takes precedence
2. ✅ **Local beats global** — Current dir beats home dir
3. ✅ **Git-aware** — Smart for monorepos
4. ✅ **Sensible default** — "ttmp" just works

**Design smell:**

The function has **two responsibilities**:
1. Find config file
2. Resolve root path

This creates **mixed concerns**. What if I want to find config but not resolve root?

**Proposed refactor:**

```go
// Separate concerns: finding vs. using config

// FindConfig returns the config and its source path
func FindConfig() (*WorkspaceConfig, string, error) {
	// Try each source, return first found
	// Return error only for malformed files
	// Return nil, "", nil if no config found
}

// ResolveRoot resolves the workspace root path
func ResolveRoot(explicitRoot string) (string, error) {
	if explicitRoot != "" && explicitRoot != "ttmp" {
		// Use explicit
		return resolveExplicitRoot(explicitRoot)
	}

	cfg, source, err := FindConfig()
	if err != nil {
		// Malformed config
		log.Printf("Warning: Config at %s is malformed: %v\n", source, err)
		log.Printf("Using default root: ttmp\n")
		return "ttmp", nil
	}

	if cfg != nil {
		// Config found
		return resolveConfigRoot(cfg.Root, source), nil
	}

	// No config, use default
	return "ttmp", nil
}
```

**Benefits:**
- Single Responsibility Principle
- Easier to test
- Clearer error handling
- Can call `FindConfig()` independently (for `config show` command)

---

## Rebuttals

### `pkg/commands/config.go` (The Configuration Manager) — Rebuttal

*[Accepting the refactor proposal]*

Alex, your refactor is spot-on. I **am** mixing concerns.

Let me show the benefit with a concrete example:

**Current code (mixed concerns):**

```go
func ResolveRoot(root string) string {
	// ... 109 lines of mixed finding + resolving
}
```

**After refactor (separated concerns):**

```go
// config.go
func FindConfig() (*WorkspaceConfig, string, error) {
	sources := []configSource{
		{name: "current dir", path: "./.ttmp.yaml"},
		{name: "parent dirs", finder: findInParents},
		{name: "home dir", path: filepath.Join(os.UserHomeDir(), ".ttmp.yaml")},
		{name: "git root", finder: findInGitRoot},
	}

	for _, src := range sources {
		cfg, err := tryLoadConfig(src)
		if err != nil {
			return nil, src.name, fmt.Errorf("malformed: %w", err)
		}
		if cfg != nil {
			return cfg, src.name, nil
		}
	}

	return nil, "", nil  // Not found (not an error)
}

func ResolveRoot(explicitRoot string, cfg *WorkspaceConfig, cfgSource string) (string, error) {
	// Much simpler: just resolve paths, don't find configs
}
```

**Benefits for `docmgr config show`:**

```go
func (c *ConfigShowCommand) Run() error {
	cfg, source, err := FindConfig()
	
	fmt.Println("Configuration sources:")
	fmt.Printf("  Found: %s\n", source)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	}
	if cfg != nil {
		fmt.Printf("  root: %s\n", cfg.Root)
		fmt.Printf("  vocabulary: %s\n", cfg.Vocabulary)
	}
	return nil
}
```

**Clean separation** of finding vs. using.

---

### Sarah Chen (The Pragmatist) — Rebuttal

*[Prioritizing fixes]*

Let me separate **quick wins** from **larger refactors**:

**Quick wins (do these now):**

1. ✅ **Add `--verbose` flag** (from Config Manager)
   - Shows which fallback was used
   - Low effort, high user value

2. ✅ **Warn on malformed config** (from Round 4)
   - Already agreed, just implement it
   - Use `log.Printf` for warnings

3. ✅ **`docmgr init` creates config file** (from Casey)
   - Template config with defaults
   - One-time setup, users know config exists

4. ✅ **Document fallback order in help text**
   - `docmgr help config` explains precedence
   - Users understand behavior

**Larger refactor (defer):**

- ❓ Separate `FindConfig()` / `ResolveRoot()` (from Alex)
  - Good design, but requires rewriting callers
  - Do this when adding `config show` command

**Don't let perfect be the enemy of good.** Ship the quick wins, iterate.

---

## Moderator Summary

### Key Findings

**Configuration System:**
- ✅ 6-level fallback chain is flexible
- ⚠️ But opaque — users don't know which fallback succeeded
- 🔥 Silent error swallowing (Round 4 issue)
- ⚠️ Mixed concerns (finding vs. resolving)

### Consensus

**Everyone agrees:**
1. ✅ Add `--verbose` flag to show resolution path
2. ✅ Warn on malformed config (Round 4 follow-up)
3. ✅ `docmgr init` should create config file
4. ✅ Document fallback order in help text
5. ✅ Add `docmgr config show` command

**Disagreement on timing:**
- Alex: Refactor now (separate concerns)
- Sarah: Quick wins first, refactor later

### Action Items

**High priority (quick wins):**
1. ✅ Add `--verbose` flag (or `DOCMGR_DEBUG` env var)
2. ✅ Warn on malformed config files
3. ✅ `docmgr init` creates `.ttmp.yaml` template
4. ✅ Document fallback order in help/README

**Medium priority:**
5. ✅ Implement `docmgr config show` command
6. ✅ Refactor: Separate `FindConfig()` / `ResolveRoot()`

### Design Principles Confirmed

**Good fallback chain design:**
1. ✅ Explicit > Implicit (--root flag first)
2. ✅ Local > Global (project config > home config)
3. ✅ Sensible defaults (ttmp just works)
4. ✅ Git-aware (monorepo support)

**Add: Transparency principle:**
5. ✅ Users should be able to see which fallback was used
6. ✅ Errors should be visible (warnings, not silent)

### Connection to Other Rounds

- **Round 4**: Silent errors → Warn on malformed config
- **Round 6**: Rename `TTMPConfig` → `WorkspaceConfig`
- **Round 9**: Document fallback order in help text

### Moderator's Observation

- **Config Manager's fallback chain is clever** — Supports many use cases
- **Opacity is the main issue** — Not a design flaw, but a visibility problem
- **Casey's experience validates** — Users need feedback when config fails
- **Quick wins available** — `--verbose`, warnings, `init` improvements
- **Alex's refactor is sound** — But can be deferred (Sarah's pragmatism wins)

**Recommendation:** Implement quick wins in next PR, refactor when adding `config show` command.
