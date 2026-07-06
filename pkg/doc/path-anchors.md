---
Title: Path Anchors for Related Files
Slug: path-anchors
Short: How docmgr stores RelatedFiles paths with explicit anchors (repo://, ws://, docs://, abs://), how they resolve, and how to migrate legacy paths.
Topics:
- docmgr
- documentation
- paths
- frontmatter
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

## Why anchors exist

A `RelatedFiles` entry like `pkg/foo.go` is ambiguous: relative to what? The
repository root? The document's directory? The docs root? Historically docmgr
guessed, and the guesses diverged between `relate`, `doctor`, search, and the
web UI. Anchored paths remove the guessing by storing an explicit scheme with
every path, so an entry means exactly the same thing no matter which tool reads
it or which directory you run docmgr from.

## The anchor schemes

| Scheme | Example | Resolves relative to |
|---|---|---|
| `repo://` | `repo://pkg/foo.go` | The repository root (where `.git` lives) |
| `ws://` | `ws://glazed/pkg/fields.go` | A go.work workspace member: `ws://<member>/<rel>` resolves against the sibling directory `<member>` under the directory containing `go.work` |
| `docs://` | `docs://2026/07/05/MEN-1--x/design/01.md` | The docs root (`ttmp/`) |
| `doc://` | `doc://../reference/01-diary.md` | The directory of the document whose frontmatter contains the entry (read-side only; never written by docmgr) |
| `abs://` | `abs:///home/user/x.go` | Nothing — it is an absolute path (escape hatch) |

Entries without a scheme are **legacy** bare paths. They still resolve through
the historical multi-anchor guessing logic, so old documents keep working, but
new writes always use anchors.

## What docmgr writes (the tightest-containing-anchor rule)

`docmgr doc relate` and `docmgr changelog update --file-note` accept absolute
or relative input paths, resolve them, and persist the tightest containing
anchor:

1. Inside the repository → `repo://<rel>`
2. Inside the go.work workspace (a sibling repo listed in `go.work`) → `ws://<member>/<rel>`
3. Inside the docs root → `docs://<rel>`
4. Anywhere else → `abs://<abs>`

docmgr never writes `doc://` and never writes repo-escaping `../` chains.

```bash
# Input path can be absolute or relative; the stored value is anchored:
docmgr doc relate --ticket MEN-4242 \
  --file-note "/home/you/project/backend/api/register.go:Registers API routes"

# Resulting frontmatter:
# RelatedFiles:
#     - Path: repo://backend/api/register.go
#       Note: Registers API routes
```

You may also pass an already-anchored path in `--file-note`
(`repo://backend/api/register.go:note`); the `:` that ends the scheme is not
mistaken for the path/note separator.

## Resolution (read side)

All consumers — `doctor`, `doc search --file/--dir` reverse lookup, the
workspace index, and the web UI — resolve anchors through one shared resolver:

- `repo://` joins against the repository root discovered for the document.
- `ws://member/rel` joins against `<workspace-root>/<member>/<rel>`, where the
  workspace root is the directory containing `go.work`.
- `docs://` joins against the resolved docs root.
- `doc://` joins against the directory of the referencing document (it may
  escape the repository; it is tolerated on read for hand-written entries).
- `abs://` is used as-is.
- Legacy bare strings fall back to the historical guessing order (repo root,
  document directory, docs root, ...), preserved for backward compatibility.

Existence is checked honestly on disk; `doctor` reports related files that do
not resolve to an existing file.

## Migrating legacy paths

Migrate a ticket (or the whole workspace) with:

```bash
# Anchor migration only
docmgr doctor --ticket MEN-4242 --fix-anchors

# Or as part of the full safe-fix pass (frontmatter auto-repair + anchors)
docmgr doctor --ticket MEN-4242 --fix
```

The migration rewrites each legacy `RelatedFiles` path to its anchored
equivalent using the same tightest-containing-anchor rule. Entries that do not
resolve to an existing file are left untouched and reported as warnings, so
migration never invents paths.

## Practical guidance

- In scripts and agent workflows, pass absolute paths to `--file-note`; docmgr
  anchors them for you. This avoids any dependence on the current directory.
- Do not hand-edit anchors unless you know the base directory semantics above;
  prefer `docmgr doc relate`.
- `ws://` anchors only appear when your repository is part of a `go.work`
  workspace; single-repo setups will only see `repo://`, `docs://`, and
  `abs://`.

## See also

- `docmgr help how-to-use` — daily workflow, including relating files
- `docmgr help cli-guide` — command reference
- `docmgr help doctor-validation-workflow` — how doctor validates related files
