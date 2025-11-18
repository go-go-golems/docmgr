---
Title: Debate Round 07 — Code Duplication (See Round 2)
Ticket: DOCMGR-CODE-REVIEW
Status: active
Topics:
    - docmgr
    - code-review
DocType: analysis
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: "Code duplication was comprehensively covered in Round 2. This round references those findings and decisions."
LastUpdated: 2025-11-18T11:15:00.000000000-05:00
---

# Debate Round 07 — Code Duplication (See Round 2)

## Note

**This debate round was already comprehensively covered in [Debate Round 02 — Command Implementation Patterns](./02-debate-round-02-command-implementation-patterns.md).**

## Summary from Round 2

### Findings

**Code duplication identified:**
- ✅ **4 frontmatter implementations** across commands
- ✅ **18 directory walk operations** across 5 commands  
- ✅ **43 file operations** across 18 command files
- ✅ **238 lines of command registration boilerplate** in `main.go`

### Decisions

**Consensus (all candidates agreed):**
1. ✅ Extract `internal/documents/` with utilities:
   - `ReadDocumentWithFrontmatter()`
   - `WriteDocumentWithFrontmatter()`
   - `WalkDocuments()`

2. ✅ Abstract command registration boilerplate (reduce `main.go` by 33%)

3. ✅ Write tests for extracted utilities

4. ✅ Document utilities to prevent future duplication

### Connection to Round 1 Decision

The Round 1 decision to create `internal/` package structure provides the perfect home for these extracted utilities:

```
internal/
├── documents/
│   ├── frontmatter.go    # Extracted from Round 2 findings
│   └── walk.go           # Extracted from Round 2 findings
├── workspace/
│   └── ...
└── templates/
    └── ...
```

### Implementation Status

**From Round 2 consensus:**
- [ ] Extract frontmatter parsing utilities
- [ ] Extract directory walking utilities
- [ ] Migrate commands incrementally
- [ ] Abstract command registration

**See Round 2 for:**
- Detailed duplication analysis
- Debate between candidates
- API design considerations (e.g., `WalkOptions` with error handling strategies)

## References

- **[Debate Round 02](./02-debate-round-02-command-implementation-patterns.md)** — Full duplication analysis and consensus
- **[Debate Round 01](./01-debate-round-01-architecture-and-code-organization.md)** — Decision to create `internal/` package
