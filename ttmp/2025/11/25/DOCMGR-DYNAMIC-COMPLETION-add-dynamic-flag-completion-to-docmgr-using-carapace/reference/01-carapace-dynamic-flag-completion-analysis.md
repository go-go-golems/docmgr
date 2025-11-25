---
Title: Carapace Dynamic Flag Completion Analysis
Ticket: DOCMGR-DYNAMIC-COMPLETION
Status: active
Topics:
    - backend
    - cli
    - completion
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: carapace/action.go
      Note: Action type definition and composition methods (Filter
    - Path: carapace/carapace.go
      Note: Main API file containing Gen() and FlagCompletion() methods - entry point for carapace integration
    - Path: carapace/compat.go
      Note: Bridge between carapace Actions and cobra's RegisterFlagCompletionFunc
    - Path: carapace/complete.go
      Note: Completion entry point that handles shell-specific argument patching and invokes traversal
    - Path: carapace/context.go
      Note: Context type providing runtime information during completion (Value
    - Path: carapace/defaultActions.go
      Note: Default action implementations including ActionCallback
    - Path: carapace/example/cmd/flag.go
      Note: Example demonstrating FlagCompletion usage with various flag types (Bool
    - Path: carapace/example/cmd/root.go
      Note: Example showing FlagCompletion with file system actions and PreInvoke hooks
    - Path: carapace/invokedAction.go
      Note: InvokedAction type representing an Action after callback invocation
    - Path: carapace/storage.go
      Note: Storage system that maintains per-command registry of completion actions
    - Path: carapace/traverse.go
      Note: Command-line traversal logic that determines what to complete based on parsing state
ExternalSources: []
Summary: Comprehensive analysis of how carapace implements dynamic flag completion for cobra commands, including architecture, key components, and integration patterns.
LastUpdated: 2025-11-25T15:16:58.17454331-05:00
---


# Carapace Dynamic Flag Completion Analysis

## Goal

This reference document provides a comprehensive analysis of how [carapace](https://github.com/carapace-sh/carapace) implements dynamic flag completion for cobra commands. It covers the architecture, key components, integration patterns, and code references to enable implementing similar functionality in docmgr.

## Context

Carapace is a command argument completion generator for [spf13/cobra](https://github.com/spf13/cobra) that provides dynamic, context-aware completion for flags and positional arguments. Unlike static completion scripts, carapace allows completion values to be computed at runtime based on:

- Current command context (which command/subcommand is active)
- Previously parsed flags and arguments
- File system state
- External command execution
- Custom callback functions

This analysis focuses specifically on **flag completion** - how carapace enables dynamic completion for command-line flags (e.g., `--ticket`, `--doc-type`, `--file`).

## Architecture Overview

Carapace uses a **storage-based architecture** where completion actions are registered per command and invoked dynamically during shell completion requests. The key components are:

1. **Storage System** (`storage.go`) - Per-command registry of completion actions
2. **Flag Completion Registration** (`compat.go`) - Bridges carapace actions to cobra's completion system
3. **Action System** (`action.go`) - Defines completion behaviors (static values, callbacks, file system, etc.)
4. **Traversal System** (`traverse.go`) - Determines what to complete based on command-line parsing state
5. **Context System** (`context.go`) - Provides runtime context (args, flags, working directory, etc.)
6. **Shell Script Generation** (`internal/shell/`) - Generates shell-specific completion scripts that call back into the binary

## How Completion Works: Carapace vs Cobra

### Cobra's Static Completion

Cobra generates **static shell scripts** that contain all completion logic:

```bash
# Cobra generates a script like this (simplified):
_complete_mycmd() {
    COMPREPLY=("--help" "--version" "--config")
}
complete -F _complete_mycmd mycmd
```

- Completion values are **baked into the script** at generation time
- Script is sourced once and stored in shell memory
- **Cannot** provide dynamic values based on runtime state
- Uses `cobra.RegisterFlagCompletionFunc` to register functions, but these are only called during script generation

### Carapace's Dynamic Completion

Carapace generates **shell scripts that call back into the binary**:

```bash
# Carapace generates a script like this (simplified):
_mycmd_completion() {
    data=$(echo "${COMP_LINE}" | xargs mycmd _carapace bash)
    COMPREPLY=($data)
}
complete -F _mycmd_completion mycmd
```

- Shell script **executes the binary** on every completion request
- Binary's `_carapace` subcommand handles the request dynamically
- **Can** provide dynamic values based on:
  - Current working directory
  - Previously parsed flags
  - File system state
  - External command execution
  - Custom callback functions

**Key Difference:** Cobra's completion functions (`RegisterFlagCompletionFunc`) are called during script generation to produce static values. Carapace's Actions are invoked **at runtime** when the user presses TAB, allowing truly dynamic completion.

## Key Components

### 1. Storage System

The storage system maintains a per-command registry of completion actions:

```74:86:carapace/carapace.go
// FlagCompletion defines completion for flags using a map consisting of name and Action.
func (c Carapace) FlagCompletion(actions ActionMap) {
	e := storage.get(c.cmd)
	e.flagMutex.Lock()
	defer e.flagMutex.Unlock()

	if e.flag == nil {
		e.flag = actions
	} else {
		for name, action := range actions {
			e.flag[name] = action
		}
	}
}
```

**Storage Entry Structure** (`storage.go`):

```17:28:carapace/storage.go
type entry struct {
	flag          ActionMap
	flagMutex     sync.RWMutex
	positional    []Action
	positionalAny *Action
	dash          []Action
	dashAny       *Action
	preinvoke     func(cmd *cobra.Command, flag *pflag.Flag, action Action) Action
	prerun        func(cmd *cobra.Command, args []string)
	bridged       bool
	initialized   bool
}
```

**Key Methods:**

- `hasFlag(cmd, name)` - Checks if a flag has completion registered
- `getFlag(cmd, name)` - Retrieves the Action for a flag, with fallback to cobra's native completion
- `preinvoke()` - Allows modifying actions before invocation (e.g., changing working directory)

### 2. Flag Completion Registration

Carapace bridges its Action system to cobra's `RegisterFlagCompletionFunc`:

```23:41:carapace/compat.go
func registerFlagCompletion(cmd *cobra.Command) {
	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if !storage.hasFlag(cmd, f.Name) {
			return // skip if not defined in carapace
		}
		if _, ok := cmd.GetFlagCompletionFunc(f.Name); ok {
			return // skip if already defined in cobra
		}

		err := cmd.RegisterFlagCompletionFunc(f.Name, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			a := storage.getFlag(cmd, f.Name)
			action := a.Invoke(Context{Args: args, Value: toComplete}) // TODO cmd might differ for persistentflags and either way args or cmd will be wrong
			return cobraValuesFor(action), cobraDirectiveFor(action)
		})
		if err != nil {
			LOG.Printf("failed to register flag completion func: %v", err.Error())
		}
	})
}
```

**Key Points:**

- Only registers completion for flags that have carapace actions defined
- Respects existing cobra completion functions (doesn't override)
- Converts carapace Actions to cobra's expected format (`[]string`, `ShellCompDirective`)
- Handles persistent flags by traversing up the command hierarchy

### 3. Action System

Actions define what values to complete. Common action types:

**Static Values:**

```101:137:carapace/example/cmd/flag.go
	carapace.Gen(flagCmd).FlagCompletion(carapace.ActionMap{
		"Bool":           carapace.ActionValues("true", "false"),
		"BoolSlice":      carapace.ActionValues("true", "false"),
		"BytesBase64":    carapace.ActionValues("MQo=", "Mgo=", "Mwo="),
		"BytesHex":       carapace.ActionValues("01", "02", "03"),
		"Count":          carapace.ActionValues(),
		"Duration":       carapace.ActionValues("1h", "2m", "3s"),
		"DurationSlice":  carapace.ActionValues("1h", "2m", "3s"),
		"Float32P":       carapace.ActionValues("1", "2", "3"),
		"Float32Slice":   carapace.ActionValues("1", "2", "3"),
		"Float64P":       carapace.ActionValues("1", "2", "3"),
		"Float64Slice":   carapace.ActionValues("1", "2", "3"),
		"Int16":          carapace.ActionValues("1", "2", "3"),
		"Int32":          carapace.ActionValues("1", "2", "3"),
		"Int32Slice":     carapace.ActionValues("1", "2", "3"),
		"Int64":          carapace.ActionValues("1", "2", "3"),
		"Int64Slice":     carapace.ActionValues("1", "2", "3"),
		"Int8":           carapace.ActionValues("1", "2", "3"),
		"Int":            carapace.ActionValues("1", "2", "3"),
		"IntSlice":       carapace.ActionValues("1", "2", "3"),
		"IPMask":         carapace.ActionValues("0.0.0.1", "0.0.0.2", "0.0.0.3"),
		"IP":             carapace.ActionValues("0.0.0.1", "0.0.0.2", "0.0.0.3"),
		"IPNet":          carapace.ActionValues("0.0.0.1/0", "0.0.0.2/0", "0.0.0.3/0"),
		"IPSlice":        carapace.ActionValues("0.0.0.1", "0.0.0.2", "0.0.0.3"),
		"StringArray":    carapace.ActionValues("1", "2", "3"),
		"String":         carapace.ActionValues("1", "2", "3"),
		"StringSlice":    carapace.ActionValues("1", "2", "3"),
		"StringToInt64":  carapace.ActionValues("a=1", "b=2", "c=3"),
		"StringToInt":    carapace.ActionValues("a=1", "b=2", "c=3"),
		"StringToString": carapace.ActionValues("a=1", "b=2", "c=3"),
		"Uint16":         carapace.ActionValues("1", "2", "3"),
		"Uint32":         carapace.ActionValues("1", "2", "3"),
		"Uint64":         carapace.ActionValues("1", "2", "3"),
		"Uint8":          carapace.ActionValues("1", "2", "3"),
		"Uint":           carapace.ActionValues("1", "2", "3"),
		"UintSlice":      carapace.ActionValues("1", "2", "3"),
	})
```

**Dynamic Callbacks:**

```26:29:carapace/defaultActions.go
// ActionCallback invokes a go function during completion.
func ActionCallback(callback CompletionCallback) Action {
	return Action{callback: callback}
}
```

**File System Actions:**

```29:33:carapace/example/cmd/root.go
	carapace.Gen(rootCmd).FlagCompletion(carapace.ActionMap{
		"chdir":           carapace.ActionDirectories(),
		"persistentFlag":  carapace.ActionValues("p1", "2", "p3"),
		"persistentFlag2": carapace.ActionValues("p4", "p5", "p6"),
	})
```

**External Command Execution:**

```31:51:carapace/defaultActions.go
// ActionExecCommand executes an external command.
//
//	carapace.ActionExecCommand("git", "remote")(func(output []byte) carapace.Action {
//	  lines := strings.Split(string(output), "\n")
//	  return carapace.ActionValues(lines[:len(lines)-1]...)
//	})
func ActionExecCommand(name string, arg ...string) func(f func(output []byte) Action) Action {
	return func(f func(output []byte) Action) Action {
		return ActionExecCommandE(name, arg...)(func(output []byte, err error) Action {
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					if firstLine := strings.SplitN(string(exitErr.Stderr), "\n", 2)[0]; strings.TrimSpace(firstLine) != "" {
						err = errors.New(firstLine)
					}
				}
				return ActionMessage(err.Error())
			}
			return f(output)
		})
	}
}
```

### 4. Context System

The Context provides runtime information during completion:

```18:33:carapace/context.go
// Context provides information during completion.
type Context struct {
	// Value contains the value currently being completed (or part of it during an ActionMultiParts).
	Value string
	// Args contains the positional arguments of current (sub)command (exclusive the one currently being completed).
	Args []string
	// Parts contains the splitted Value during an ActionMultiParts (exclusive the part currently being completed).
	Parts []string
	// Env contains environment variables for current context.
	Env []string
	// Dir contains the working directory for current context.
	Dir string

	mockedReplies map[string]string
	cmd           *cobra.Command // needed for ActionCobra
}
```

**Context Creation:**

```35:56:carapace/context.go
// NewContext creates a new context for given arguments.
func NewContext(args ...string) Context {
	if len(args) == 0 {
		args = append(args, "")
	}

	context := Context{
		Value: args[len(args)-1],
		Args:  args[:len(args)-1],
		Env:   os.Environ(),
	}

	if wd, err := os.Getwd(); err == nil {
		context.Dir = wd
	}

	if m, err := env.Sandbox(); err == nil {
		context.Dir = m.WorkDir()
		context.mockedReplies = m.Replies
	}
	return context
}
```

### 5. Traversal and Flag Detection

The traversal system determines what to complete based on command-line parsing:

```132:136:carapace/traverse.go
	// flag argument
	case inFlag != nil && inFlag.Consumes(context.Value):
		LOG.Printf("completing flag argument of %#v for arg %#v\n", inFlag.Name, context.Value)
		context.Parts = inFlag.Args
		return storage.getFlag(cmd, inFlag.Name), context
```

**Flag Lookup with Parent Traversal:**

```88:113:carapace/storage.go
func (s _storage) getFlag(cmd *cobra.Command, name string) Action {
	if flag := cmd.LocalFlags().Lookup(name); flag == nil && cmd.HasParent() {
		return s.getFlag(cmd.Parent(), name)
	} else {
		entry := s.get(cmd)
		entry.flagMutex.RLock()
		defer entry.flagMutex.RUnlock()

		flagAction, ok := entry.flag[name]
		if !ok {
			if f, ok := cmd.GetFlagCompletionFunc(name); ok {
				flagAction = ActionCobra(f)
			}
		}

		a := s.preinvoke(cmd, flag, flagAction)

		return ActionCallback(func(c Context) Action { // TODO verify order of execution is correct
			invoked := a.Invoke(c)
			if invoked.action.meta.Usage == "" {
				invoked.action.meta.Usage = flag.Usage
			}
			return invoked.ToA()
		})
	}
}
```

**Key Features:**

- Traverses up command hierarchy for persistent flags
- Falls back to cobra's native completion if no carapace action defined
- Applies `preinvoke` hooks before action invocation
- Automatically sets usage text from flag definition

### 6. Cobra Integration Bridge

The bridge system registers completion functions during cobra initialization:

```52:74:carapace/storage.go
func (s _storage) bridge(cmd *cobra.Command) {
	if entry := storage.get(cmd); !entry.bridged {
		bridgeMutex.Lock()
		defer bridgeMutex.Unlock()

		if entry := storage.get(cmd); !entry.bridged {
			cobra.OnInitialize(func() {
				if !entry.initialized {
					bridgeMutex.Lock()
					defer bridgeMutex.Unlock()

					if !entry.initialized {
						registerValidArgsFunction(cmd)
						registerFlagCompletion(cmd)
						entry.initialized = true
					}

				}
			})
			entry.bridged = true
		}
	}
}
```

**Important Note:** While carapace uses `cobra.RegisterFlagCompletionFunc` internally, **it does NOT rely on cobra's static completion generation**. Instead, carapace provides its own dynamic completion system that requires shell scripts to be sourced.

### 7. Shell Completion Script Generation

Carapace generates shell scripts that must be sourced to enable completion. These scripts call back into the binary dynamically:

**Shell Script Generation:**

```114:117:carapace/carapace.go
// Snippet creates completion script for given shell.
func (c Carapace) Snippet(name string) (string, error) {
	return shell.Snippet(c.cmd, name)
}
```

**Example Bash Script:**

```12:40:carapace/internal/shell/bash/snippet.go
// Snippet creates the bash completion script.
func Snippet(cmd *cobra.Command) string {
	result := fmt.Sprintf(`#!/bin/bash
_%[1]v_completion() {
  export COMP_LINE
  export COMP_POINT
  export COMP_TYPE
  export COMP_WORDBREAKS

  local nospace data compline="${COMP_LINE:0:${COMP_POINT}}"

  data=$(echo "${compline}''" | xargs %[2]v _carapace bash 2>/dev/null)
  if [ $? -eq 1 ]; then
    data=$(echo "${compline}'" | xargs %[2]v _carapace bash 2>/dev/null)
    if [ $? -eq 1 ]; then
    	data=$(echo "${compline}\"" | xargs %[2]v _carapace bash 2>/dev/null)
    fi
  fi

  IFS=$'\001' read -r -d '' nospace data <<<"${data}"
  mapfile -t COMPREPLY < <(echo "${data}")
  unset COMPREPLY[-1]

  [ "${nospace}" = true ] && compopt -o nospace
  local IFS=$'\n'
  [[ "${COMPREPLY[*]}" == "" ]] && COMPREPLY=() # fix for mapfile creating a non-empty array from empty command output
}

complete -o noquote -F _%[1]v_completion %[1]v
`, cmd.Name(), uid.Executable())

	return result
}
```

**Key Points:**

1. **Shell scripts must be sourced** - Users run `source <(command _carapace bash)` to enable completion
2. **Scripts call back into binary** - When completion is triggered, the shell script executes `command _carapace bash` with the command line
3. **Hidden `_carapace` subcommand** - Carapace adds a hidden subcommand that handles completion requests
4. **Dynamic execution** - Unlike cobra's static completion, carapace executes the binary on every completion request, allowing dynamic values

**Completion Flow:**

1. User types `command --flag <TAB>` in shell
2. Shell's completion system calls the sourced function (e.g., `_command_completion`)
3. Shell script executes: `command _carapace bash` with the command line
4. Binary's `_carapace` subcommand runs → calls `complete()` function
5. `traverse()` determines what to complete (flag, positional, etc.)
6. `storage.getFlag()` retrieves the Action for the flag
7. Action is invoked with Context → returns completion values
8. Values formatted for shell → returned to shell script
9. Shell script populates `COMPREPLY` array → shell displays completions

**Initialization Flow:**

1. `carapace.Gen(cmd)` is called → creates Carapace wrapper
2. `addCompletionCommand(cmd)` → adds hidden `_carapace` subcommand
3. `storage.bridge(cmd)` → registers cobra initialization hook
4. When cobra initializes → `registerFlagCompletion(cmd)` runs
5. Flag completion functions are registered with cobra (for fallback compatibility)
6. **Shell script must be sourced** → `source <(command _carapace bash)`
7. On completion request → Shell script calls binary → `_carapace` subcommand → Carapace Actions → Values returned

### 7. Value Format Conversion

Carapace converts its Action results to cobra's expected format:

```43:53:carapace/compat.go
func cobraValuesFor(action InvokedAction) []string {
	result := make([]string, len(action.action.rawValues))
	for index, r := range action.action.rawValues {
		if r.Description != "" {
			result[index] = fmt.Sprintf("%v\t%v", r.Value, r.Description)
		} else {
			result[index] = r.Value
		}
	}
	return result
}
```

**Shell Directive Conversion:**

```55:64:carapace/compat.go
func cobraDirectiveFor(action InvokedAction) cobra.ShellCompDirective {
	directive := cobra.ShellCompDirectiveNoFileComp
	for _, val := range action.action.rawValues {
		if action.action.meta.Nospace.Matches(val.Value) {
			directive = directive | cobra.ShellCompDirectiveNoSpace
			break
		}
	}
	return directive
}
```

## Usage Pattern

### Basic Example

```go
import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

var myCmd = &cobra.Command{
	Use:   "mycmd",
	Short: "Example command",
}

func init() {
	myCmd.Flags().String("ticket", "", "Ticket identifier")
	myCmd.Flags().String("doc-type", "", "Document type")
	
	// Register flag completions
	carapace.Gen(myCmd).FlagCompletion(carapace.ActionMap{
		"ticket": carapace.ActionCallback(func(c carapace.Context) carapace.Action {
			// Dynamic completion: list tickets from workspace
			tickets := listTicketsFromWorkspace()
			return carapace.ActionValues(tickets...)
		}),
		"doc-type": carapace.ActionValues("design-doc", "reference", "playbook"),
	})
}
```

### Advanced Example with Context

```go
carapace.Gen(cmd).FlagCompletion(carapace.ActionMap{
	"file": carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		// Use context to filter based on other flags
		if ticket := c.cmd.Flag("ticket").Value.String(); ticket != "" {
			return carapace.ActionFiles().Chdir("ttmp/" + ticket)
		}
		return carapace.ActionFiles()
	}),
})
```

### PreInvoke Hook Example

```56:58:carapace/example/cmd/root.go
	carapace.Gen(rootCmd).PreInvoke(func(cmd *cobra.Command, flag *pflag.Flag, action carapace.Action) carapace.Action {
		return action.Chdir(rootCmd.Flag("chdir").Value.String())
	})
```

This allows modifying actions based on other flag values (e.g., changing working directory based on `--chdir` flag).

## Key Design Patterns

### 1. Lazy Registration

Completion functions are registered only when `carapace.Gen()` is called, not during flag definition. This allows:
- Conditional completion registration
- Dynamic command structure (commands added in PreRun hooks)
- Clean separation of concerns

### 2. Action Composition

Actions can be composed and transformed:

```go
carapace.ActionValues("a", "b", "c").
	Filter("b").                    // Remove "b"
	StyleF(style.ForKeyword).       // Apply styling
	Usage("Select option").         // Add usage text
	NoSpace()                       // Don't add space after completion
```

### 3. Fallback Chain

1. Check carapace storage for flag action
2. If not found, check parent command (for persistent flags)
3. If still not found, use cobra's native completion function
4. If none exist, return empty completion

### 4. Thread Safety

Storage uses `sync.RWMutex` for concurrent access:

```76:86:carapace/storage.go
func (s _storage) hasFlag(cmd *cobra.Command, name string) bool {
	if flag := cmd.LocalFlags().Lookup(name); flag == nil && cmd.HasParent() {
		return s.hasFlag(cmd.Parent(), name)
	} else {
		entry := s.get(cmd)
		entry.flagMutex.RLock()
		defer entry.flagMutex.RUnlock()
		_, ok := entry.flag[name]
		return ok
	}
}
```

## Integration Points for docmgr

### Potential Flag Completions

1. **`--ticket`** - Complete with ticket IDs from workspace
   ```go
   "ticket": carapace.ActionCallback(func(c carapace.Context) carapace.Action {
       tickets := listTicketsFromWorkspace()
       return carapace.ActionValues(tickets...)
   }),
   ```

2. **`--doc-type`** - Complete with vocabulary doc types
   ```go
   "doc-type": carapace.ActionCallback(func(c carapace.Context) carapace.Action {
       types := loadDocTypesFromVocabulary()
       return carapace.ActionValues(types...)
   }),
   ```

3. **`--file` / `--file-note`** - Complete with file paths, filtered by ticket
   ```go
   "file": carapace.ActionCallback(func(c carapace.Context) carapace.Action {
       if ticket := c.cmd.Flag("ticket").Value.String(); ticket != "" {
           return carapace.ActionFiles().Chdir("ttmp/" + ticket)
       }
       return carapace.ActionFiles()
   }),
   ```

4. **`--topics`** - Complete with vocabulary topics
   ```go
   "topics": carapace.ActionCallback(func(c carapace.Context) carapace.Action {
       topics := loadTopicsFromVocabulary()
       return carapace.ActionValues(topics...)
   }),
   ```

5. **`--status`** - Complete with vocabulary status values
   ```go
   "status": carapace.ActionCallback(func(c carapace.Context) carapace.Action {
       statuses := loadStatusesFromVocabulary()
       return carapace.ActionValues(statuses...)
   }),
   ```

## Related Files and Symbols

### Core Files

- `carapace/carapace.go` - Main API (`Gen()`, `FlagCompletion()`)
- `carapace/storage.go` - Storage system (`_storage`, `entry`, `getFlag()`, `hasFlag()`)
- `carapace/compat.go` - Cobra integration (`registerFlagCompletion()`, `cobraValuesFor()`)
- `carapace/action.go` - Action type and methods
- `carapace/context.go` - Context type and creation
- `carapace/traverse.go` - Command-line traversal logic
- `carapace/complete.go` - Completion entry point

### Key Types

- `Carapace` - Wrapper around `*cobra.Command`
- `Action` - Completion behavior definition
- `ActionMap` - Map of flag names to Actions (`map[string]Action`)
- `Context` - Runtime completion context
- `InvokedAction` - Action after invocation (contains results)
- `CompletionCallback` - Function type for dynamic actions

### Key Functions

- `carapace.Gen(*cobra.Command) *Carapace` - Initialize carapace for a command
- `FlagCompletion(ActionMap)` - Register flag completions
- `ActionValues(...string) Action` - Create static value action
- `ActionCallback(CompletionCallback) Action` - Create callback action
- `ActionFiles(...string) Action` - Create file system action
- `ActionDirectories() Action` - Create directory action
- `ActionExecCommand(string, ...string)` - Create external command action

## Usage Examples

### Example 1: Static Values

```go
carapace.Gen(cmd).FlagCompletion(carapace.ActionMap{
	"output": carapace.ActionValues("json", "yaml", "csv"),
})
```

### Example 2: Dynamic from Workspace

```go
carapace.Gen(cmd).FlagCompletion(carapace.ActionMap{
	"ticket": carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		workspace := loadWorkspace()
		tickets := make([]string, 0, len(workspace.Tickets))
		for id := range workspace.Tickets {
			tickets = append(tickets, id)
		}
		return carapace.ActionValues(tickets...)
	}),
})
```

### Example 3: File System with Context

```go
carapace.Gen(cmd).FlagCompletion(carapace.ActionMap{
	"file": carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		// Filter files based on ticket flag
		if ticket := c.cmd.Flag("ticket").Value.String(); ticket != "" {
			ticketPath := filepath.Join("ttmp", ticket)
			return carapace.ActionFiles().Chdir(ticketPath)
		}
		return carapace.ActionFiles()
	}),
})
```

### Example 4: Multi-part Values

```go
carapace.Gen(cmd).FlagCompletion(carapace.ActionMap{
	"file-note": carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		// Complete file paths, then allow note entry
		return carapace.ActionFiles().MultiParts(":")
	}),
})
```

## Implementation Checklist for docmgr

- [ ] Add carapace as dependency to `go.mod`
- [ ] Call `carapace.Gen(rootCmd)` in root command initialization
- [ ] Create completion actions for key flags:
  - [ ] `--ticket` (list tickets from workspace)
  - [ ] `--doc-type` (load from vocabulary)
  - [ ] `--topics` (load from vocabulary)
  - [ ] `--status` (load from vocabulary)
  - [ ] `--file` / `--file-note` (file system with ticket filtering)
  - [ ] `--owners` (could complete with common team members)
- [ ] Test completion in bash/zsh/fish
- [ ] Document completion setup in user guide
- [ ] Consider PreInvoke hooks for context-aware completions

## Related

- [Carapace Documentation](https://carapace-sh.github.io/carapace/)
- [Carapace GitHub Repository](https://github.com/carapace-sh/carapace)
- [Cobra Completion Guide](https://github.com/spf13/cobra/blob/main/shell_completions.md)
- Ticket: DOCMGR-DYNAMIC-COMPLETION
