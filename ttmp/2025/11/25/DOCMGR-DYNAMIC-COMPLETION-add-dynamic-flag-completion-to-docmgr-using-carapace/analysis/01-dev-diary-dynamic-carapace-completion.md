---
Title: Dev Diary — Dynamic Carapace Completion
Ticket: DOCMGR-DYNAMIC-COMPLETION
Status: active
Topics:
    - backend
    - cli
    - completion
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-11-25T16:42:02.710377294-05:00
---

# Dev Diary — Dynamic Carapace Completion

## 2025-11-25 — tmux dynamic completion smoke test

- What I did:
  - Built `dist/docmgr` and put it on PATH in tmux panes (bash, zsh).
  - Sourced dynamic snippets: `source <(docmgr _carapace bash|zsh)`.
  - Programmatically invoked completion via `_carapace` for:
    - `doc add --doc-type`, `--status`, `--intent`, `--topics`, `--ticket`.

- What worked:
  - Bash returned vocabulary-driven completions:
    - doc-types: design-doc, reference, playbook, index
    - status: draft, active, review, complete, archived
    - intent: long-term, short-term, throwaway, ticket-specific, only-during-ticket
    - topics: backend, cli, docmgr, documentation, glaze, …
  - Tickets completion included `DOCMGR-DYNAMIC-COMPLETION` and other existing tickets.
  - Zsh produced the same sets (with descriptions rendered by zsh snippet).

- What didn’t work:
  - N/A in this pass. Didn’t cover fish/pwsh yet, and didn’t validate other verbs.

- What I learned:
  - Calling `_carapace bash|zsh` with the current line (`compline''`) reproduces interactive completion in tests.
  - The first line’s `false\\u0001…` prefix indicates the no-space directive followed by values (bash).

- What to do better:
  - Add a small helper script to exercise common completions to avoid manual `COMP_LINE` boilerplate.
  - Expand tests to fish and to the rest of the verbs in the design checklist.

### Captured samples (bash)

```
doc-type: design-doc, reference, playbook, index
status: draft, active, review, complete, archived
intent: long-term, short-term, throwaway, ticket-specific, only-during-ticket
topics: backend, cli, docmgr, documentation, glaze, …
tickets: DOCMGR-DYNAMIC-COMPLETION, DOCMGR-*, …
```

## 2025-11-25 — Verb wiring, zsh checks, and README completion docs

- What I did:
  - Wired carapace FlagCompletion across verbs:
    - doc: add, list, search, relate, renumber, guidelines
    - ticket: create, list/tickets, close, rename-ticket
    - tasks: list, add, check, uncheck, edit, remove (dynamic id completion)
    - meta: update (field enum + dynamic value for status/intent/topics/docType)
    - changelog: update (ticket, file-note multiparts, topics)
    - vocab: list/add (category enum)
    - workspace: doctor, init/status (root)
    - template: validate (root, path)
  - Built docmgr and verified zsh completion using tmux (two panes: bash/zsh). Confirmed root commands and flag names still autocomplete, and dynamic values work.
  - Added a Shell Completion section to README covering dynamic (carapace) and static (cobra) install for bash, zsh, fish, and PowerShell.

- What worked:
  - Dynamic flags return live values (tickets, vocabulary, files/dirs).
  - Traditional command/flag completions unaffected.
  - Zsh testing flows via `_carapace` snippet; tabbing shows menus (menu select enabled).

- What didn’t work:
  - Capturing zsh’s interactive menu isn’t very verbose in logs (expected behavior). The helper still sends TAB and we confirmed behavior manually.

- What I learned:
  - Centralizing completion actions (tickets/vocab/files/dirs/task IDs/meta fields/values) simplifies per-verb registration.
  - The postfix templates (e.g., tasks, list docs/tickets) intentionally print YAML at the end; it’s controlled by verb templates under `ttmp/templates/**`.

- What to do better next time:
  - Extend the tmux helper to run a verb-by-verb zsh matrix and save concise “expected” lists side-by-side for easier diffing.
  - Add fish coverage to the helper for parity.

- Next steps:
  - Fish and PowerShell snippet validation.
  - Optional: suppress postfix templates for certain human outputs or add a flag to hide them.

