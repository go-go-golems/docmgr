# Changelog

## 2026-07-05

- Initial workspace created


## 2026-07-05

Investigation complete: four file:line-anchored codebase reviews; go-minitrace pipeline over 240 sessions (14166 docmgr calls in 139 sessions) with reusable JS query commands; intern-level analysis/design/implementation guide written (paths-anchor design + agent CLI contract + doctor overhaul + UI parity + docmgr ai subsystem). Live-reproduced two bugs during bookkeeping (comma-split --file-note notes silently dropped; positional task-ID trap). (commit 427676e + follow-up)

### Related Files

- /home/manuel/workspaces/2026-07-05/improve-docmgr/docmgr/ttmp/2026/07/05/DOCMGR-200-improve-docmgr-for-coding-agents--improve-docmgr-as-a-documentation-tool-for-coding-agents-full-review-ux-frontmatter-paths-agent-usage-analysis/design-doc/01-improving-docmgr-for-coding-agents-analysis-design-and-implementation-guide.md — Primary deliverable
- /home/manuel/workspaces/2026-07-05/improve-docmgr/docmgr/ttmp/2026/07/05/DOCMGR-200-improve-docmgr-for-coding-agents--improve-docmgr-as-a-documentation-tool-for-coding-agents-full-review-ux-frontmatter-paths-agent-usage-analysis/reference/01-investigation-diary.md — Chronological diary (steps 1-5)


## 2026-07-05

Second deliverable: go-minitrace field report (analysis/01-...) - strengths + 9-item friction log + CLI ergonomics + JS API assessment + measured adapter-fidelity matrix with source-anchored root causes (toolUseResult dropped; emit timestamps overwritten; codex scrapes timing from output text; dual query stack with DuckDB toll) + prioritized P0-P3 backlog.

### Related Files

- /home/manuel/workspaces/2026-07-05/improve-docmgr/docmgr/ttmp/2026/07/05/DOCMGR-200-improve-docmgr-for-coding-agents--improve-docmgr-as-a-documentation-tool-for-coding-agents-full-review-ux-frontmatter-paths-agent-usage-analysis/analysis/01-go-minitrace-field-report-assessment-from-the-docmgr-usage-mining-project.md — The report


## 2026-07-06

Implementation complete (P0-P4 + docs/skills refresh, docmgr ai skipped per user). Commits db2cca4 (P0 silent failures/honest exists/doctor output), 66db822 (P1 agent contract), 2f4fef8 (P2 paths v2 anchors), 044dc37 (P3 doctor v2 + stable task IDs), 685d509 (P4 UI parity), plus docs refresh. go-minitrace fully implemented in its worktree (adapters P0, single-engine migration with DuckDB removed at -73.5MB binary, intake, UX, docs) - commits pending due to read-only parent gitdir.


## 2026-07-06

Step 13: PR 43 code/project review added; local Go/UI validation passes, but review found merge blockers around doc add canonical ticket metadata, stable task IDs in HTTP/UI, and a failing Advanced Security CodeQL check.

### Related Files

- /home/manuel/workspaces/2026-07-05/improve-docmgr/docmgr/ttmp/2026/07/05/DOCMGR-200-improve-docmgr-for-coding-agents--improve-docmgr-as-a-documentation-tool-for-coding-agents-full-review-ux-frontmatter-paths-agent-usage-analysis/analysis/02-pr-43-code-review-and-project-review.md — PR 43 intern-style review deliverable
- /home/manuel/workspaces/2026-07-05/improve-docmgr/docmgr/ttmp/2026/07/05/DOCMGR-200-improve-docmgr-for-coding-agents--improve-docmgr-as-a-documentation-tool-for-coding-agents-full-review-ux-frontmatter-paths-agent-usage-analysis/scripts/04-pr43-review-experiments.sh — Validation and reproduction harness
- /home/manuel/workspaces/2026-07-05/improve-docmgr/docmgr/ttmp/2026/07/05/DOCMGR-200-improve-docmgr-for-coding-agents--improve-docmgr-as-a-documentation-tool-for-coding-agents-full-review-ux-frontmatter-paths-agent-usage-analysis/sources/pr43-review-experiments.txt — Saved test and experiment output


## 2026-07-06

Step 13 follow-up: expanded PR 43 review with a third correctness blocker: doctor --ticket keeps the raw short ref and can report No tickets checked while ticket show resolves the same ref.

### Related Files

- /home/manuel/workspaces/2026-07-05/improve-docmgr/docmgr/internal/workspace/query_docs_sql.go — ScopeTicket uses exact ticket_id SQL
- /home/manuel/workspaces/2026-07-05/improve-docmgr/docmgr/pkg/commands/doctor.go — Doctor builds ScopeTicket from raw settings.Ticket
- /home/manuel/workspaces/2026-07-05/improve-docmgr/docmgr/ttmp/2026/07/05/DOCMGR-200-improve-docmgr-for-coding-agents--improve-docmgr-as-a-documentation-tool-for-coding-agents-full-review-ux-frontmatter-paths-agent-usage-analysis/analysis/02-pr-43-code-review-and-project-review.md — Updated PR 43 review with doctor short-ref finding

