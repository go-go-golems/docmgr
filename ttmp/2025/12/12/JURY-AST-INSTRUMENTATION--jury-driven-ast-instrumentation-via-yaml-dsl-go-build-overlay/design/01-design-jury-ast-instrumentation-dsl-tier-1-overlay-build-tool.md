---
Title: 'Design: Jury AST Instrumentation DSL (Tier 1) + Overlay Build Tool'
Ticket: JURY-AST-INSTRUMENTATION
Status: active
Topics:
    - refactor
    - tooling
    - ast
    - observability
DocType: design
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-12T19:53:57.119029799-05:00
---

# Design: Jury AST Instrumentation DSL (Tier 1) + Overlay Build Tool

## 1. Goal (what we’re building)

We want a way for “the jury” (reviewers, judges, test harnesses) to **observe the internal behavior** of specific code paths (e.g. skip-policy decisions) **without**:

- adding logging hooks to production code,
- adding build tags or “jury awareness” to the repository,
- or requiring developers to think about jury mode while coding.

The deliverable is:

1) A small **YAML DSL** that describes what to observe (Tier‑1 events: `on_entry`, `on_return`, `on_call`, `on_error`) and what to output.  
2) A **tool** that:
   - reads the YAML,
   - generates instrumented *replacement files* outside the repo,
   - produces a Go **overlay JSON** mapping original source files → instrumented copies,
   - compiles a jury binary using `go build -overlay ...`.

The entire approach is “opt‑in” and leaves the main codebase untouched.

## 2. Intern-friendly background: why are we doing this?

If you’re new to the project, here’s the problem we keep running into:

- The codebase contains pieces of “critical behavior” that are correct but subtle.
- When reviewing a refactor, we often want evidence that the behavior is correct for:
  - edge cases,
  - defaults,
  - and tricky path normalization / filtering semantics.
- The best evidence is to **run real code** on real inputs and show “why” decisions were made.

Normally we might add logging or extra tests. But:

- adding logging is invasive (and can accidentally become “forever logging”),
- adding tests sometimes requires exposing internals or writing a lot of harness code,
- and both approaches bias the implementation to be “testable for this review” rather than “clean for long-term use”.

So we want a third option:

> The jury can temporarily instrument the code *at build time* using an overlay, and then run the instrumented binary to observe behavior — while the production code remains clean.

This is effectively “aspect‑oriented observation”, but implemented as:

- a declarative YAML file (the “aspect specification”),
- a code transformation tool (AST-based rewrite),
- and a compilation overlay (so we never modify the repo’s files).

## 3. Scope: Tier‑1 DSL features (and what we explicitly do NOT do)

### 3.1 Tier‑1 events (MVP)

We support observing these events for a target function:

- **`on_entry`**: just after entering the function body
- **`on_return`**: just before each `return` statement
- **`on_call`**: around calls *from within* the function to another function
- **`on_error`**: when the function assigns/receives an error and checks `err != nil` (Go-idiomatic error path)

### 3.2 Out of scope for MVP (avoid overcomplexity)

We explicitly do not support (yet):

- arbitrary boolean expression languages for `when`
- modifying runtime behavior (no rewriting return values)
- injecting new package dependencies into the main codebase
- adding brand-new `.go` source files to the repo (overlay replaces existing files only)
- concurrency probes (`go` statements), allocation probes, lock probes, etc.
- full “trace everything” mode across the entire repo by default

## 4. Tier‑1 YAML DSL (MVP spec)

### 4.1 File shape

YAML file name is arbitrary, but we’ll assume:

- `judging.yaml` lives outside the repo or in a local-only folder (e.g. `.jury/judging.yaml`).
- It is never committed.

At the top level:

- **`version`**: string (e.g. `"v1"`)
- **`name`**: human friendly name (for output headers)
- **`targets`**: list of instrumentation targets
- **`output`**: optional output settings

### 4.2 Core concepts (terms)

- **Target**: a function we are instrumenting (e.g. `internal/workspace.DefaultIngestSkipDir`).
- **Event**: a hook point in that function (`on_entry`, `on_return`, `on_call`, `on_error`).
- **Template**: a string that can refer to captured values.

### 4.3 Naming functions and symbols (important)

In the DSL we refer to functions using **import-path-like identifiers**:

- `pkg`: Go package import path relative to module root (e.g. `github.com/go-go-golems/docmgr/internal/workspace`)
- `func`: function name (e.g. `DefaultIngestSkipDir`)

For MVP we require targets to be:

- **top-level functions** or **methods** declared in project source (not generated from cgo, not in stdlib)
- resolvable by parsing the repo using `go list`/`go/packages` (tool detail in §6)

### 4.4 Captures: `arg.*`, `ret.*`, `call.*`, `err.*`

We support a simple “access model” for templates:

- **Function args**: `arg.<paramName>`
  - Example: `arg.docPath`
  - Example: `arg.d.Name()` (method call) is allowed only in “capture expressions”, not in templates (MVP simplification).
- **Return values**: `ret` or `ret0` for first return value
  - If function returns multiple values: `ret0`, `ret1`, ...
- **Call context**: `call.<callee>` has:
  - `call.callee`: callee name
  - `call.args`: captured call args (by position: `call.args[0]`, `call.args[1]`)
  - `call.ret0` (if the call expression is assigned)
- **Error context**: `err.value` (the error variable name/value in scope), and optionally `err.site` (which call produced it) when statically detectable.

MVP simplification:

- Templates can reference only values that the tool explicitly captures at that hook point.
- Captures can be declared; templates refer to capture names.

### 4.5 Conditions: `when` (MVP)

We allow a minimal conditional filter:

- `when: always` (default)
- `when: <captureName> == <literal>`
- `when: <captureName> != <literal>`

Literal types: strings, booleans.

This is intentionally tiny to keep implementation small and predictable.

## 5. YAML examples (Tier‑1)

### 5.1 Example A: observe skip decisions (`on_return`)

```yaml
version: v1
name: "DJ Skippy: skip decisions"

targets:
  - target:
      pkg: github.com/go-go-golems/docmgr/internal/workspace
      func: DefaultIngestSkipDir

    on_return:
      capture:
        - name: dir
          from: arg.d_name        # tool-provided derived capture (see note below)
        - name: skip
          from: ret0

      log: "{skip? '❌ SKIP' : '✅ INDEX'}  {dir}"
      when: always
```

**Note:** for MVP we allow a small set of “derived captures” for common Go patterns to avoid parsing expressions inside YAML:

- `arg.d_name` means: `d.Name()` if an argument named `d` exists and has a `Name() string` method.

This keeps YAML simple and reduces an entire class of “expression parsing” complexity.

### 5.2 Example B: observe tags (`on_entry` + `on_return`)

```yaml
version: v1
name: "DJ Skippy: tag computation"

targets:
  - target:
      pkg: github.com/go-go-golems/docmgr/internal/workspace
      func: ComputePathTags

    on_entry:
      capture:
        - name: path
          from: arg.docPath
      log: "→ ComputePathTags({path})"

    on_return:
      capture:
        - name: path
          from: arg.docPath
        - name: tags
          from: ret0
      log: "← tags for {path}: index={tags.IsIndex} control={tags.IsControlDoc} archived={tags.IsArchivedPath} scripts={tags.IsScriptsPath} sources={tags.IsSourcesPath}"
      when: always
```

### 5.3 Example C: observe internal helper calls (`on_call`)

```yaml
version: v1
name: "DJ Skippy: segment boundary checks"

targets:
  - target:
      pkg: github.com/go-go-golems/docmgr/internal/workspace
      func: ComputePathTags

    on_call:
      callee:
        pkg: github.com/go-go-golems/docmgr/internal/workspace
        func: containsPathSegment

      capture:
        - name: seg
          from: call.arg0
        - name: path
          from: call.arg1
        - name: matched
          from: call.ret0

      log: "containsPathSegment(seg={seg}) => {matched}"
      when: matched == true
```

**MVP note:** We allow `call.arg0`, `call.arg1`, … and `call.ret0` because calls don’t have named parameters.

### 5.4 Example D: observe error paths (`on_error`)

```yaml
version: v1
name: "Index build: error paths"

targets:
  - target:
      pkg: github.com/go-go-golems/docmgr/internal/workspace
      func: ingestWorkspaceDocs

    on_error:
      capture:
        - name: err
          from: err.value
        - name: site
          from: err.site
      log: "ERROR: {err} (site={site})"
      when: always
```

**MVP note:** `err.site` is “best effort”:

- If the tool can determine the call expression feeding the error (e.g. `x, err := Foo()`),
  it records `site=Foo`.
- If not, it sets `site=?`.

## 6. Tool design: AST transform + Go overlay build

### 6.1 Key requirement: *no jury artifacts in the repo*

This design assumes:

- we do not commit generated files,
- we do not add build tags to production files,
- and we do not add “jury stubs” to the codebase.

Therefore, the tool must:

- generate instrumented copies **outside** the repo tree (e.g. `/tmp/jury-gen/...`),
- and build using **`go build -overlay`** so the compiler sees replacements.

### 6.2 What is an overlay (for interns)

The Go toolchain supports a flag:

```bash
go build -overlay overlay.json ./cmd/docmgr
```

Where `overlay.json` is a mapping:

- “when you compile file A, actually read file B instead”.

This is how IDEs compile your code with unsaved buffers, and it’s also useful for “compile a transformed version of the code without modifying the repo.”

### 6.3 Non-judge mode build pipeline

Non-judge mode means: **no YAML file**.

Pipeline:

1. Developer runs normal build:
   ```bash
   go build -o /tmp/docmgr ./cmd/docmgr
   ```
2. No transformation tool is invoked.
3. No overlay is used.
4. Resulting binary is a normal binary.

This is the default developer workflow. Nobody thinks about the jury.

### 6.4 Judge mode build pipeline (overlay-based)

Judge mode means: **YAML file exists and is explicitly provided**.

Pipeline:

1. Judge writes a YAML file (e.g. `.jury/judging.yaml`).
2. Run the jury tool:
   - parse YAML
   - locate target functions in source
   - generate instrumented replacement files into `/tmp/jury-gen/<hash>/...`
   - generate `/tmp/jury-gen/<hash>/overlay.json`
3. Build using overlay:
   ```bash
   go build -overlay /tmp/jury-gen/<hash>/overlay.json -o /tmp/docmgr-jury ./cmd/docmgr
   ```
4. Run the jury binary to observe output.

### 6.5 Output behavior

The instrumented code should output to:

- `stderr` by default (so it doesn’t interfere with command output pipelines),
- optionally to a file (configured by YAML `output` section in a later iteration).

MVP output is simple line-based logs.

## 7. AST rewriting rules (Tier‑1)

This section describes what the tool would do to implement Tier‑1 events.

### 7.1 `on_entry`

Rewrite: at the start of the function body, insert:

- capture statements (copy args into local temps if needed)
- log statement

### 7.2 `on_return`

Rewrite: for each `return ...` statement:

- evaluate return values into temps (so we can log without double-evaluating expressions),
- log,
- return temps.

This is the main reason AST rewriting is useful: it can safely capture returns without changing semantics.

### 7.3 `on_call`

Rewrite: for calls inside the target function:

- wrap call sites to capture duration and values if needed:
  - before: `x := Foo(a, b)`
  - after:
    - log “before call”
    - `tmp := Foo(a, b)`
    - log “after call” (including tmp)
    - `x := tmp`

MVP can start with:

- only instrumenting call expressions that appear in assignment contexts (`:=` or `=`),
- and only capturing `call.ret0` for single-value calls.

### 7.4 `on_error`

We focus on idiomatic patterns:

- `if err != nil { ... return err }`
- `if err := Foo(); err != nil { ... }`
- `x, err := Foo(); if err != nil { ... }`

Rewrite:

- immediately inside the `if err != nil {` block, inject a log statement capturing:
  - `err` value
  - optional site (best effort)

This gives high-signal observation of the “unhappy path” without tracing everything.

## 8. Safety and developer-experience constraints

### 8.1 “Do not change behavior”

All transformations must be observational only:

- no reordering that changes semantics,
- no swallowing panics,
- no changing return values.

### 8.2 “Do not pollute the repo”

We keep the repo clean by:

- generating replacement files outside the tree
- using overlay for compilation
- producing a distinct output binary (`docmgr-jury`)

### 8.3 Reproducibility

To make jury runs repeatable:

- The tool should print:
  - the config file path,
  - the overlay path,
  - and the git revision / module version (if available).

## 9. Open questions / follow-ups

- Should the DSL allow matching methods (`type: Receiver.Method`) in v1, or defer?
- How far do we go in resolving “callee identity” for `on_call` (imports, aliases, selectors)?
- How strict should target resolution be (must be unique match)?
- Do we want `output.format=json` early for machine scoring, or keep MVP human-first?

## 10. Appendix: minimal output section (optional)

MVP can omit this; if we include it:

```yaml
output:
  stream: stderr        # stderr|stdout
  prefix: "🎭 JURY "     # prefix for each line
```

