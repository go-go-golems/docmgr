package commands

// TemplateContent holds template content for different document types
var TemplateContent = map[string]string{
	"index": `---
Title: {{TITLE}}
Ticket: {{TICKET}}
Status: draft
Topics:
{{TOPICS}}
DocType: index
Intent: short-term
Owners:
{{OWNERS}}
RelatedFiles: []
ExternalSources: []
Summary: >
  {{SUMMARY}}
LastUpdated: {{DATE}}
---

# {{TITLE}}

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **{{STATUS}}**

## Topics

{{TOPICS_LIST}}

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
`,

	"design-doc": `---
Title: {{TITLE}}
Ticket: {{TICKET}}
Status: draft
Topics:
{{TOPICS}}
DocType: design-doc
Intent: long-term
Owners:
{{OWNERS}}
RelatedFiles: []
ExternalSources: []
Summary: >
  {{SUMMARY}}
LastUpdated: {{DATE}}
---

# {{TITLE}}

## Executive Summary

<!-- Provide a high-level overview of the design proposal -->

## Problem Statement

<!-- Describe the problem this design addresses -->

## Proposed Solution

<!-- Describe the proposed solution in detail -->

## Design Decisions

<!-- Document key design decisions and rationale -->

## Alternatives Considered

<!-- List alternative approaches that were considered and why they were rejected -->

## Implementation Plan

<!-- Outline the steps to implement this design -->

## Open Questions

<!-- List any unresolved questions or concerns -->

## References

<!-- Link to related documents, RFCs, or external resources -->
`,

	"reference": `---
Title: {{TITLE}}
Ticket: {{TICKET}}
Status: draft
Topics:
{{TOPICS}}
DocType: reference
Intent: long-term
Owners:
{{OWNERS}}
RelatedFiles: []
ExternalSources: []
Summary: >
  {{SUMMARY}}
LastUpdated: {{DATE}}
---

# {{TITLE}}

## Goal

<!-- What is the purpose of this reference document? -->

## Context

<!-- Provide background context needed to use this reference -->

## Quick Reference

<!-- Provide copy/paste-ready content, API contracts, or quick-look tables -->

## Usage Examples

<!-- Show how to use this reference in practice -->

## Related

<!-- Link to related documents or resources -->
`,

	"working-note": `---
Title: {{TITLE}}
Ticket: {{TICKET}}
Status: draft
Topics:
{{TOPICS}}
DocType: working-note
Intent: short-term
Owners:
{{OWNERS}}
RelatedFiles: []
ExternalSources: []
Summary: >
  {{SUMMARY}}
LastUpdated: {{DATE}}
---

# {{TITLE}}

## Summary

<!-- Brief summary for LLM ingestion -->

## Notes

<!-- Free-form notes, meeting summaries, or research findings -->

## Decisions

<!-- Any decisions made during this working session -->

## Next Steps

<!-- Actions or follow-ups -->
`,

	"tutorial": `---
Title: {{TITLE}}
Ticket: {{TICKET}}
Status: draft
Topics:
{{TOPICS}}
DocType: tutorial
Intent: long-term
Owners:
{{OWNERS}}
RelatedFiles: []
ExternalSources: []
Summary: >
  {{SUMMARY}}
LastUpdated: {{DATE}}
---

# {{TITLE}}

## Overview

<!-- What will readers learn in this tutorial? -->

## Prerequisites

<!-- What knowledge or setup is required? -->

## Step-by-Step Guide

### Step 1: Setup

<!-- First step -->

### Step 2: Implementation

<!-- Next steps -->

## Verification

<!-- How to verify the tutorial was completed successfully -->

## Troubleshooting

<!-- Common issues and solutions -->

## Related Resources

<!-- Links to related documentation or examples -->
`,

	"playbook": `---
Title: {{TITLE}}
Ticket: {{TICKET}}
Status: draft
Topics:
{{TOPICS}}
DocType: playbook
Intent: short-term
Owners:
{{OWNERS}}
RelatedFiles: []
ExternalSources: []
Summary: >
  {{SUMMARY}}
LastUpdated: {{DATE}}
---

# {{TITLE}}

## Purpose

<!-- What does this playbook accomplish? -->

## Environment Assumptions

<!-- What environment or setup is required? -->

## Commands

<!-- List of commands to execute -->

` + "```bash" + `
# Command sequence
` + "```" + `

## Exit Criteria

<!-- What indicates success or completion? -->

## Notes

<!-- Additional context or warnings -->
`,

	"task-list": `---
Title: {{TITLE}}
Ticket: {{TICKET}}
Status: draft
Topics:
{{TOPICS}}
DocType: task-list
Intent: short-term
Owners:
{{OWNERS}}
RelatedFiles: []
ExternalSources: []
Summary: >
  {{SUMMARY}}
LastUpdated: {{DATE}}
---

# {{TITLE}}

## Tasks

- [ ] Task 1
- [ ] Task 2
- [ ] Task 3

## Completed

- [x] Example completed task

## Notes

<!-- Additional context or blockers -->
`,

	"log": `---
Title: {{TITLE}}
Ticket: {{TICKET}}
Status: draft
Topics:
{{TOPICS}}
DocType: log
Intent: short-term
Owners:
{{OWNERS}}
RelatedFiles: []
ExternalSources: []
Summary: >
  {{SUMMARY}}
LastUpdated: {{DATE}}
---

# {{TITLE}}

<!-- Log entries in reverse chronological order (newest first) -->

## {{DATE}} - Entry Title

<!-- Log entry content -->

## {{DATE}} - Entry Title

<!-- Previous log entry -->
`,

	"script": `---
Title: {{TITLE}}
Ticket: {{TICKET}}
Status: draft
Topics:
{{TOPICS}}
DocType: script
Intent: throwaway
Owners:
{{OWNERS}}
RelatedFiles: []
ExternalSources: []
Summary: >
  {{SUMMARY}}
LastUpdated: {{DATE}}
---

# {{TITLE}}

## Purpose

<!-- What does this script do? -->

## Usage

` + "```bash" + `
# Usage example
` + "```" + `

## Implementation

<!-- Describe the script implementation or link to executable file -->

## Notes

<!-- Additional context or warnings -->
`,
}

// GetTemplate returns the template content for a given doc type
func GetTemplate(docType string) (string, bool) {
	template, ok := TemplateContent[docType]
	return template, ok
}

