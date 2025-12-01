package templates

import (
	"bytes"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/go-go-golems/docmgr/pkg/models"
)

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

	"code-review": `---
Title: {{TITLE}}
Ticket: {{TICKET}}
Status: draft
Topics:
{{TOPICS}}
DocType: code-review
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

## Summary

<!-- One-paragraph summary of the review scope and outcome -->

## Context

<!-- PRs, branches, or features reviewed; link to tickets and references -->

## Files Reviewed

<!-- Bullet list of key files; add rationale in RelatedFiles notes when possible -->

## Findings

- Strengths:
- Issues / Risks:

## Decisions & Follow-ups

- Decisions:
- Action Items:

## References

<!-- Links to PRs, commits, docs, or external resources -->
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

// LoadTemplate is now implemented in embedded.go with priority:
// filesystem (for customization) > embedded FS > legacy string map
// This function is kept for backwards compatibility but delegates to embedded.go

// extractFrontmatterAndBody splits a template into (frontmatter, body) using adrg/frontmatter library.
// If no frontmatter is found, returns ("", template).
// For templates with placeholders ({{TITLE}}, etc.), falls back to manual parsing since
// the library can't parse invalid YAML.
func ExtractFrontmatterAndBody(tpl string) (string, string) {
	// Use adrg/frontmatter library for robust parsing (works for valid YAML)
	reader := bytes.NewReader([]byte(tpl))
	var meta map[string]interface{}
	bodyBytes, err := frontmatter.Parse(reader, &meta)
	if err == nil {
		// Successfully parsed - return the body
		return "", string(bodyBytes)
	}

	// Parsing failed - could be:
	// 1. No frontmatter (template starts with content)
	// 2. Invalid YAML (template has placeholders like {{TITLE}})
	// For templates with placeholders, manually extract the body by finding the --- delimiter

	// Check if this looks like a template with frontmatter (has --- at start)
	if !strings.HasPrefix(strings.TrimSpace(tpl), "---") {
		// No frontmatter marker, return full template as body
		return "", tpl
	}

	// Try to manually extract body by finding the frontmatter delimiter
	// Look for the pattern: ---\n...content...\n---
	lines := strings.Split(tpl, "\n")
	if len(lines) < 2 {
		return "", tpl
	}

	// Find the first --- (should be line 0)
	if strings.TrimSpace(lines[0]) != "---" {
		return "", tpl
	}

	// Find the closing ---
	bodyStart := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			bodyStart = i + 1
			break
		}
	}

	if bodyStart == -1 {
		// No closing --- found, return full template
		return "", tpl
	}

	// Extract body (everything after the closing ---)
	body := strings.Join(lines[bodyStart:], "\n")
	// Remove leading newline if present
	body = strings.TrimPrefix(body, "\n")

	return "", body
}

// renderTemplateBody replaces placeholders in the template body based on the document values
func RenderTemplateBody(body string, doc *models.Document) string {
	now := time.Now().Format("2006-01-02")

	// Build lists
	topicsList := ""
	if len(doc.Topics) > 0 {
		var lines []string
		for _, t := range doc.Topics {
			lines = append(lines, "- "+t)
		}
		topicsList = strings.Join(lines, "\n")
	}

	// Indented YAML-ish lists for optional use in bodies
	topicsYaml := "[]"
	if len(doc.Topics) > 0 {
		var lines []string
		for _, t := range doc.Topics {
			lines = append(lines, "  - "+t)
		}
		topicsYaml = strings.Join(lines, "\n")
	}

	ownersYaml := "[]"
	if len(doc.Owners) > 0 {
		var lines []string
		for _, o := range doc.Owners {
			lines = append(lines, "  - "+o)
		}
		ownersYaml = strings.Join(lines, "\n")
	}

	r := strings.NewReplacer(
		"{{TITLE}}", doc.Title,
		"{{TICKET}}", doc.Ticket,
		"{{STATUS}}", doc.Status,
		"{{DATE}}", now,
		"{{SUMMARY}}", doc.Summary,
		"{{TOPICS_LIST}}", topicsList,
		"{{TOPICS}}", topicsYaml,
		"{{OWNERS}}", ownersYaml,
	)

	return r.Replace(body)
}
