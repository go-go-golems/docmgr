package commands

// GuidelineContent holds guideline content for different document types
var GuidelineContent = map[string]string{
	"index": `# Guidelines: Index Documents

## Purpose
The index document is the canonical entry point for a ticket workspace. It provides a single source of truth for understanding the ticket's scope, status, and key resources.

## Required Elements
- **Overview**: Brief description of the ticket and its goals
- **Status**: Current status of the ticket
- **Key Links**: References to related files and external sources
- **Topics**: The main topics this ticket addresses
- **Tasks**: Link to tasks.md
- **Changelog**: Link to changelog.md
- **Structure**: Brief overview of the directory structure

## Best Practices
- Keep the overview concise but informative
- Update status regularly as work progresses
- Maintain RelatedFiles to help LLMs discover relevant code
- Link to important decisions and documents
- Use consistent formatting for easy scanning
`,

	"design-doc": `# Guidelines: Design Documents

## Purpose
Design documents provide structured rationale and architecture notes. They document decisions, alternatives considered, and implementation plans.

## Required Elements
- **Executive Summary**: High-level overview
- **Problem Statement**: Clear description of the problem
- **Proposed Solution**: Detailed solution description
- **Design Decisions**: Key decisions with rationale
- **Alternatives Considered**: Other approaches evaluated
- **Implementation Plan**: Steps to implement

## Best Practices
- Start with executive summary for quick scanning
- Document the "why" behind decisions, not just the "what"
- Include diagrams or code examples where helpful
- Keep open questions visible until resolved
- Link to related RFCs, tickets, or references
- Use clear section hierarchy for easy navigation
`,

	"reference": `# Guidelines: Reference Documents

## Purpose
Reference documents provide copy/paste-ready context, API contracts, prompt packs, or quick-look tables. They're designed for reuse in LLM prompts or as quick reference during development.

## Required Elements
- **Goal**: What this reference accomplishes
- **Context**: Background needed to use the reference
- **Quick Reference**: The actual reference content (code, tables, prompts)
- **Usage Examples**: How to use this reference

## Best Practices
- Focus on copy/paste-ready content
- Keep context minimal but sufficient
- Format for easy scanning (tables, code blocks)
- Include practical examples
- Update when APIs or processes change
- Link to related references or design docs
`,

	"working-note": `# Guidelines: Working Notes

## Purpose
Working notes capture free-form logs, meeting summaries, research findings, or exploratory thoughts. They're ephemeral but help maintain context during active work.

## Required Elements
- **Summary**: Brief summary for LLM ingestion (at the top)
- **Notes**: The main content (free-form)
- **Decisions**: Any decisions made
- **Next Steps**: Follow-up actions

## Best Practices
- Always include a summary at the top
- Keep notes focused and organized
- Extract important decisions to changelog.md
- Mark resolved items clearly
- Use consistent date formatting
- Archive when ticket is complete
`,

	"tutorial": `# Guidelines: Tutorial Documents

## Purpose
Tutorials provide step-by-step workflows or reusable prompt playbooks. They guide readers through a process from start to finish.

## Required Elements
- **Overview**: Learning objectives
- **Prerequisites**: Required knowledge or setup
- **Step-by-Step Guide**: Detailed walkthrough
- **Verification**: How to verify completion
- **Troubleshooting**: Common issues and solutions

## Best Practices
- Start with clear learning objectives
- Break steps into logical chunks
- Include code examples and expected outputs
- Provide verification steps
- Anticipate common pitfalls
- Link to related resources
- Follow glazed documentation style guidelines
`,

	"playbook": `# Guidelines: Playbook Documents

## Purpose
Playbooks document command sequences, manual test procedures, or operational steps. They're designed for repeated execution.

## Required Elements
- **Purpose**: What the playbook accomplishes
- **Environment Assumptions**: Required setup
- **Commands**: The actual command sequence
- **Exit Criteria**: Success indicators

## Best Practices
- List all prerequisites upfront
- Use code blocks for commands
- Include expected outputs
- Document failure modes
- Update when commands change
- Keep environment assumptions explicit
- Include timing estimates if relevant
`,

	"task-list": `# Guidelines: Task List Documents

## Purpose
Task lists provide canonical TODO tracking with checkboxes and owners. They keep execution visible and machine-readable.

## Required Elements
- **Tasks**: List of uncompleted tasks with checkboxes
- **Completed**: List of completed tasks
- **Notes**: Additional context or blockers

## Best Practices
- Use Markdown checkboxes for machine readability
- Assign owners to tasks when possible
- Move completed items to completed section
- Keep tasks specific and actionable
- Update regularly as work progresses
- Link to related documents or code
`,

	"log": `# Guidelines: Log Documents

## Purpose
Logs maintain a running changelog, incident timeline, or decision log. They track what happened and when.

## Required Elements
- **Log Entries**: Chronological entries (newest first)

## Best Practices
- Use reverse chronological order (newest first)
- Include dates in entry headers
- Keep entries concise but informative
- Document decisions and their rationale
- Link to related commits or documents
- Maintain consistent formatting
- Update promptly after notable events
`,

	"script": `# Guidelines: Script Documents

## Purpose
Script documents describe temporary code, SQL snippets, or REPL transcripts. They live in the scripts/ directory alongside executable files.

## Required Elements
- **Purpose**: What the script does
- **Usage**: How to run it
- **Implementation**: Description or link to executable

## Best Practices
- Keep scripts focused on a single purpose
- Document all required parameters
- Include usage examples
- Note any side effects or dependencies
- Link to related design docs if applicable
- Archive when no longer needed
- Include a README.md in scripts/ summarizing all scripts
`,

	"code-review": `# Guidelines: Code Review Documents

## Purpose
Capture the outcome of code reviews with enough context that future readers can understand scope, findings, and decisions.

## Required Elements
- **Summary**: One-paragraph overview of what was reviewed and the conclusion
- **Context**: PRs, branches, features; links to tickets and references
- **Files Reviewed**: Key files and rationale (consider notes in RelatedFiles)
- **Findings**: Strengths and issues/risks
- **Decisions & Follow-ups**: Decisions made and action items

## Best Practices
- Link to PRs, commits, and related docs
- Keep findings concise; prefer bullet lists
- Capture actionable follow-ups with owners
- Use RelatedFiles with short notes to explain why files matter
- Update when follow-ups are completed
`,
}

// GetGuideline returns the guideline content for a given doc type
func GetGuideline(docType string) (string, bool) {
	guideline, ok := GuidelineContent[docType]
	return guideline, ok
}

// ListGuidelineTypes returns all available guideline types
func ListGuidelineTypes() []string {
	types := make([]string, 0, len(GuidelineContent))
	for k := range GuidelineContent {
		types = append(types, k)
	}
	return types
}
