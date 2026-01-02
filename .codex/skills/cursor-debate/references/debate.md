# Source

Copied from /home/manuel/.cursor/commands/debate.md

# Debate Framework — Technical Decision Exploration

## Overview

Run a presidential-style debate to explore complex technical decisions, surface ideas, trade-offs, and perspectives. This command uses multiple perspectives (including personified code entities) to argue with real codebase evidence.

**Important:** Debates don't make decisions—they surface ideas and arguments. The decision-maker reviews the debate rounds and then makes informed choices based on the evidence presented.

## When to Use

Use this framework when:
- **Decision has multiple valid approaches** (e.g., architecture choices, integration strategies)
- **Stakes are high** (affects many developers, touches many files)
- **Trade-offs are unclear** (pros/cons need careful analysis)
- **Data exists but interpretation varies** (codebase metrics, implementation details)

## Workflow

1. **Identify the question** - What technical decision needs exploration?
2. **Select candidates** - Which perspectives should participate? (See candidate types below)
3. **Research first** - Document actual commands, queries, and findings
4. **Write debate round** - Opening statements → Rebuttals → Moderator summary
5. **Store with docmgr** - Use docmgr to create and manage debate documents

## Candidate Types

### Human Developer Personas
- **The Pragmatist** - "Ship it and iterate"; focuses on cost, velocity, practical usability
- **The Architect** - "Structure enables scale"; focuses on boundaries, extensibility, long-term coherence
- **The Researcher** - "Understand deeply before building"; focuses on existing implementations, compatibility
- **The Integrator** - "Make tools work together"; focuses on APIs, developer experience, workflows
- **The Tool Builder** - "Build powerful tools"; focuses on query expressiveness, reusable primitives, use cases

### Code Entity Personas
- Personify actual code modules (e.g., `oak/pkg/patternmatcher/`, `internal/treesitter/parser.go`)
- Give them stats, perspective, personality, and tools
- Let them defend their design choices and capabilities

### Wildcards
- **The New Developer** - Fresh eyes, naive questions, learnability focus
- **Domain Experts** - Represent specific paradigms (e.g., "Prolog/Unification/Datalog" for declarative queries)
- **External Perspectives** - Historical context, dependency managers, etc.

## Debate Round Structure

```markdown
## Pre-Debate Research
[Document what each candidate discovered with actual commands/queries]

## Opening Statements (Round 1)
[Each candidate argues their position with data]

## Rebuttals (Round 2)
[Candidates respond to each other, adjust positions based on evidence]

## Moderator Summary
[Extract key arguments, tensions, interesting ideas, trade-offs, open questions]
```

## Research-First Principle

**Critical:** Research FIRST, then write the debate. Do NOT write arguments without evidence.

For each candidate:
1. Identify research needs - What questions do they need answered?
2. Run analysis tools - Execute queries, searches, file reads
3. Document research in "Pre-Debate Research" section - Show commands and results
4. Write opening statements - Candidates argue using the research data
5. Write rebuttals - Candidates respond to each other's evidence
6. Write moderator summary - Extract key arguments and tensions

## Example Usage

**Question:** "How should we integrate PAIP pattern matcher with Tree-sitter AST queries?"

**Candidates:**
- Taylor "The Tool Builder" - Needs query expressiveness for code review tools
- "Prolog/Unification/Datalog" - Evaluates declarative query approaches
- Jordan "The Researcher" - Analyzes existing patterns
- Morgan "The Integrator" - Considers workflow integration

**Research:**
- Analyze current implementation limitations
- Test Tree-sitter query capabilities
- Review PAIP pattern matcher API
- Identify real-world use cases

**Output:** Debate round document with research, arguments, and synthesis

## Integration with docmgr

After creating debate rounds:
1. Store in `reference/debate-round-N.md` using docmgr
2. Relate relevant files with `docmgr doc relate`
3. Update changelog with `docmgr changelog update`
4. Create synthesis document after all rounds complete

## Key Principles

- **Show, don't tell** - Use actual code examples, file paths, data
- **Let candidates use tools mid-debate** - Show queries being run
- **Allow position changes** - Best debates show learning from evidence
- **Keep it grounded** - Balance personality with rigorous analysis
- **Document everything** - Commands, queries, results must be reproducible

## Anti-Patterns

❌ **Predetermined outcome** - Don't run debate to justify a decision already made
❌ **Fake research** - Don't make up data; run actual queries
❌ **Ignoring evidence** - If data contradicts hypothesis, adjust position
❌ **Skipping synthesis** - Extract decisions into design doc and RFC after debates

## References

- Debate framework playbook: See project documentation
- docmgr documentation: `docmgr help how-to-use`
- Example debates: See `reference/debate-round-*.md` files

---

**Use this framework when you need rigorous, data-driven exploration of complex technical decisions.**
