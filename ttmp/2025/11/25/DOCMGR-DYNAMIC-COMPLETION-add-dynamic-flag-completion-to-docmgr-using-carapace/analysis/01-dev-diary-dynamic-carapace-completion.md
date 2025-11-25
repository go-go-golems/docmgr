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

