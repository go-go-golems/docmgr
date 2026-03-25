---
Title: Implementation plan for Glazed facade migration
Ticket: 008-GLAZED-FACADE-MIGRATION
Status: active
Topics:
    - docmgr
    - glaze
    - cli
    - tooling
    - testing
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../glazed/pkg/doc/tutorials/migrating-to-facade-packages.md
      Note: Primary migration reference
    - Path: cmd/docmgr/cmds/common/common.go
      Note: Shared command wiring that must move from layers to sections
    - Path: pkg/commands/add.go
      Note: Representative Glaze command still using removed packages
    - Path: pkg/commands/doctor.go
      Note: Complex command with old parsing and processor APIs
ExternalSources: []
Summary: ""
LastUpdated: 2026-03-24T21:55:58.945538476-04:00
WhatFor: ""
WhenToUse: ""
---


# Implementation plan for Glazed facade migration

## Executive Summary

`docmgr` no longer compiles against the current `glazed` release because the legacy `layers`, `parameters`, and `middlewares` packages were removed. The migration needs to move command definitions, values decoding, and Cobra wiring onto the new facade packages: `schema`, `fields`, `values`, and `sources`.

The work should proceed in a narrow dependency order. First fix the shared command builder and the Glaze command definitions in `pkg/commands`, then update any remaining wrappers and tests, and finally run the full test suite to catch the API edges that are not obvious from the first compiler errors.

## Problem Statement

`go test ./...` currently fails during package loading because many `docmgr` packages still import removed Glazed packages. The first compile errors come from `pkg/commands/add.go`, but the breakage is broad:

- Shared command wiring in `cmd/docmgr/cmds/common/common.go` still configures output defaults via legacy layers.
- Most Glaze-backed commands in `pkg/commands` still build flags with `parameters.NewParameterDefinition`, decode settings through `layers.ParsedLayers`, and emit rows through the old middleware processor interface.
- `scenariolog` helper binaries still import removed packages and will need the same migration treatment.

Until those imports and call sites are updated, `docmgr` cannot build, tests cannot run, and follow-on feature work is blocked.

## Proposed Solution

Adopt the facade-package migration guide as the source of truth and migrate `docmgr` in four slices:

1. Update command-definition construction.
   Replace legacy parameter/layer helpers with `fields.New(...)`, `schema.NewSection(...)`, and `cmds.WithSections(...)` or equivalent schema setters.

2. Update runtime parsing and output plumbing.
   Replace `RunIntoGlazeProcessor(... parsedLayers *layers.ParsedLayers, gp middlewares.Processor)` patterns with the current values/sources interfaces expected by the new Glazed version, then decode settings via `values.Values.DecodeSectionInto(...)`.

3. Update shared Cobra wiring.
   Rework `cmds/common/common.go` to initialize the standard Glazed section through the new `settings.NewGlazedSection(...)` API and configure short-help sections/parser options with `schema.DefaultSlug` rather than the removed layer constants.

4. Sweep remaining helpers, tests, and documentation.
   Update `scenariolog`, any command tests that assert Glazed integration behavior, and local docs/examples that still instruct contributors to use removed packages.

## Design Decisions

- Migrate to the new APIs directly instead of adding compatibility wrappers.
  The migration guide is explicit that the old aliases are gone. A local shim would add dead weight and hide future drift.

- Fix the shared builder first.
  `common.BuildCommand` affects most CLI entry points. Updating it early reduces duplicate fixes and makes later compiler errors more specific.

- Keep the migration incremental and commit in focused slices.
  This repo already uses ticket workspaces and changelogs; small commits make it easier to review whether the breakage is in schema definition, Cobra integration, or result emission.

- Use the migration guide and compiler failures as the primary acceptance criteria.
  The repo is already pinned to `github.com/go-go-golems/glazed v0.7.3`, so local compile/test results are more trustworthy than trying to infer behavior from stale docs.

## Alternatives Considered

- Downgrade `glazed` to a pre-facade version.
  Rejected because the user explicitly wants `docmgr` to work with the new Glazed version.

- Vendor temporary compatibility packages inside `docmgr`.
  Rejected because that would preserve obsolete concepts (`layers`, `parameters`) and make future upgrades harder.

- Update only the files reported by the first compile error.
  Rejected because the breakage is systemic across `pkg/commands`, `cmds/common`, and `scenariolog`; a shallow fix would leave the tree half-migrated.

## Implementation Plan

1. Capture the current failure state in the diary and ticket docs.
2. Migrate `cmd/docmgr/cmds/common/common.go` from legacy layers/settings helpers to sections/schema-based helpers.
3. Update `pkg/commands` Glaze commands:
   convert command definitions to fields/sections, update struct tags if needed, and replace parsed-layer decoding plus middleware processor usage with the new values/sources API.
4. Update `scenariolog` helpers and any tests that compile Glazed-backed commands.
5. Run `gofmt -w` on changed files and `go test ./...` until the workspace is green.
6. Update tasks, changelog, and diary after each completed slice and commit each slice separately.

## Open Questions

- Whether `docmgr` can keep the current `RunIntoGlazeProcessor` shape with only type substitutions, or whether the new Glazed release expects a different command interface entirely.
- Whether `glazed.parameter` struct tags remain accepted or need to be reduced to the newer `glazed:"..."` form everywhere in the repo.
- Whether any documentation under `pkg/doc` should be updated in this ticket or left for a follow-up once the code compiles again.

## References

- `glazed/pkg/doc/tutorials/migrating-to-facade-packages.md`
- `cmd/docmgr/cmds/common/common.go`
- `pkg/commands/add.go`
- `pkg/commands/doctor.go`
