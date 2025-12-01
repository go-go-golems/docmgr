# Changelog

## 2025-12-01

- Initial workspace created


## 2025-12-01

Added diagnostics core scaffold, vocabulary/related-file contexts+rules, and diagnostics smoke test script.


## 2025-12-01

Added diagnostics smoke script under test-scenarios/testing-doc-manager/15-diagnostics-smoke.sh to exercise docmgr binary with vocabulary/related file warnings.


## 2025-12-01

Moved diagnostics rendering helpers into pkg/diagnostics/docmgr adapter and refactored doctor to use it; added diagnostics diary working-note.


## 2025-12-01

Added frontmatter/template taxonomies and rules; wrapped frontmatter parsing and template validate errors into taxonomy; checked tasks 3,6,10,14,20,22.


## 2025-12-01

Added listing/workspace taxonomies and rules; wired list_docs parse skips to taxonomy; cleaned imports and reran tests/smoke.


## 2025-12-01

Wired doctor staleness warnings to workspace taxonomy rendering; added listing/workspace taxonomies and rules earlier.


## 2025-12-01

Wired listing and workspace taxonomies (missing_index, stale) in doctor and list_docs; ensured meta_update/relate use taxonomy-wrapped errors via frontmatter parsing; added playbook guidance.


## 2025-12-01

Expanded diagnostics smoke to cover vocab/related/listing/workspace/frontmatter/template cases; fixed paths and reran successfully (expected template parse error).


## 2025-12-01

Added constructors helper for taxonomies; wired missing_index/stale to taxonomy; reran expanded diagnostics smoke; updated related files.


## 2025-12-01

Updated diagnostics playbook to reference constructors helper usage.


## 2025-12-01

Expanded diagnostics handover doc with context, code map, open work, and pseudocode per style guide.


## 2025-12-01

Expanded diagnostics handover doc with detailed context, steps, and file pointers per style guide.


## 2025-12-01

Wired doctor to emit frontmatter schema taxonomies (required fields, missing status/topics) and render parse errors; go test ./pkg/commands ./pkg/diagnostics/...


## 2025-12-01

Added doctor --diagnostics-json flag + renderer collector, updated diagnostics smoke to assert JSON output, and added adapter unit test.

