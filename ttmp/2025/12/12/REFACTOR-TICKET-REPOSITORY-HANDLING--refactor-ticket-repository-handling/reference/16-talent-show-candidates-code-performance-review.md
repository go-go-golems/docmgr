---
Title: Talent Show Candidates - Code Performance Review
Ticket: REFACTOR-TICKET-REPOSITORY-HANDLING
Status: active
Topics:
    - refactor
    - tickets
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/paths/resolver.go
      Note: Norma's assistant - multi-anchor path resolution engine
    - Path: internal/workspace/index_builder.go
      Note: Ingrid the Indexer - The Librarian
    - Path: internal/workspace/normalization.go
      Note: Norma the Navigator - The Polyglot
    - Path: internal/workspace/query_docs.go
      Note: Q the Coordinator - The Conductor
    - Path: internal/workspace/query_docs_sql.go
      Note: SQL Sorcerer - The Wizard
    - Path: internal/workspace/skip_policy.go
      Note: DJ Skippy - The Bouncer
    - Path: internal/workspace/sqlite_export.go
      Note: Artie the Archivist - The Packager
    - Path: internal/workspace/sqlite_schema.go
      Note: Schema Shapeshifter - The Architect
    - Path: internal/workspace/workspace.go
      Note: Ward the Warden - The Sentinel
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-12T19:22:41.665274704-05:00
---



# 🎭 Workspace Refactor Talent Show: "Code's Got Talent!"

## Goal

Present the major components of the Workspace+SQLite refactor as personified contestants in a talent show, where each will demonstrate their unique abilities through actual executable performance tests. This serves as both a creative code review technique and a comprehensive integration test suite.

## Context

The REFACTOR-TICKET-REPOSITORY-HANDLING introduces a unified `workspace.Workspace` API backed by an in-memory SQLite index. Instead of a traditional code review, we're conducting a talent show where each subsystem is a contestant that will perform live demonstrations of their capabilities.

Think of this as "America's Got Talent" meets "The Coding Interview" — each candidate presents their background, talent, and then performs a live act that we can judge for correctness, elegance, and robustness.

## The Contestants

### 🎪 Contestant #1: "The Bouncer" (Skip Policy)
**Real Name:** `skip_policy.go`  
**Stage Name:** DJ Skippy  
**Hometown:** `internal/workspace`  
**Age:** Fresh (just born in this refactor)

**Background Story:**
DJ Skippy grew up in a chaotic neighborhood where files of all types tried to sneak into the club. Some were sketchy (`.meta/` directories), others were clearly underage (`_templates/`, `_guidelines/`), and a few were VIPs who deserved special treatment (`archive/`, `scripts/`). After years of apprenticeship under various inconsistent bouncers across the codebase, DJ Skippy finally got the solo gig as THE canonical gatekeeper for document ingestion.

**Talent:** Path-segment boundary detection and category tagging  
**Special Move:** Can tell the difference between `/archive/` (VIP area) and `/myarchive/` (just a weird naming choice)

**Performance Act:**
"The Directory Gauntlet" — DJ Skippy will be given a series of tricky paths and must correctly decide:
- Skip or index?
- What tags to apply?

**Test Cases:**
```go
// Act 1: The Classic Skip
Path: "ttmp/.meta/implementation-notes.md"
Expected: SKIP ENTIRELY ❌

// Act 2: The Underscore Challenge  
Path: "ttmp/_templates/ticket-template.md"
Expected: SKIP ENTIRELY ❌

// Act 3: The Archive VIP Pass
Path: "ttmp/2025/12/12/TICKET-123/archive/old-design.md"
Expected: INDEX with is_archived_path=true ✅🏷️

// Act 4: The Fake Archive
Path: "ttmp/2025/12/12/myarchive-project/doc.md"
Expected: INDEX normally (no special tags) ✅

// Act 5: The Control Doc Recognition
Path: "ttmp/2025/12/12/TICKET-123/tasks.md" (sibling to index.md)
Expected: INDEX with is_control_doc=true ✅🏷️

// Act 6: The Nested README Trick
Path: "ttmp/2025/12/12/TICKET-123/design/README.md" (no index.md sibling)
Expected: INDEX normally (NOT is_control_doc) ✅
```

**How Judges Can Watch the Performance:**

```bash
# Method 1: Run the enhanced performance test suite
go test ./internal/workspace -run TestSkipPolicy -v

# Method 2: Analyze a real workspace
go run ./cmd/docmgr workspace analyze-skip --root ttmp

# Method 3: See detailed JSON report
go run ./cmd/docmgr workspace analyze-skip --root ttmp --show-all --output json

# Method 4: Only show interesting decisions (tagged or skipped)
go run ./cmd/docmgr workspace analyze-skip --root ttmp
```

**What Judges Will See:**

The performance tests produce richly formatted output showing:
- 🎪 Act headers with difficulty ratings (⭐ to ⭐⭐⭐)
- ✅ or ❌ for each decision with detailed reasoning
- 📊 Performance summaries with accuracy metrics
- 📋 JSON reports for automated analysis

Example output:
```
🎪 ACT 2: The Segment Boundary Challenge
═══════════════════════════════════════════
✅ ⭐⭐⭐ Advanced  myarchive/doc.md
   Expected tags: archived=false, scripts=false, sources=false
   Actual tags:   archived=false, scripts=false, sources=false
   Explanation: False positive avoidance: 'myarchive' is not '/archive/'
   ✓ JUDGMENT: CORRECT
```

**Judging Criteria:**
- Correctness of skip decisions
- Accuracy of tag assignment
- No false positives on segment-boundary checks
- Consistent behavior across path representations
- Observable decision-making process
- Clear error messages when rules are violated

---

### 🎪 Contestant #2: "The Librarian" (Index Builder)
**Real Name:** `index_builder.go`  
**Stage Name:** Ingrid the Indexer  
**Hometown:** `internal/workspace`  
**Age:** Seasoned professional (evolved from multiple walker implementations)

**Background Story:**
Ingrid spent years watching chaos unfold as different parts of the codebase reinvented document walking. Each command had its own ideas about what constituted "the set of docs." Some silently skipped broken files, others crashed, and nobody could agree on what to do with malformed YAML. After witnessing one too many "works on my machine" bugs, Ingrid decided to become THE canonical source of truth — one walk to rule them all.

**Talent:** Single-pass document ingestion with graceful error handling  
**Special Move:** Indexes broken docs with `parse_ok=0` instead of dropping them, making repair workflows possible

**Performance Act:**
"The Broken Document Ballet" — Ingrid will ingest a workspace containing:
- Perfect docs
- Syntactically broken YAML
- Missing frontmatter
- Files with unusual RelatedFiles paths

**Test Cases:**
```go
// Act 1: The Perfect Document
Input: Valid markdown with complete frontmatter
Expected: 
  - parse_ok=1 ✅
  - All fields hydrated correctly
  - Topics extracted
  - RelatedFiles normalized

// Act 2: The Syntax Error
Input: Markdown with malformed YAML (unclosed quote)
Expected:
  - parse_ok=0 ❌
  - parse_err populated with helpful message
  - ticket_id inferred from directory structure
  - Still queryable via ScopeTicket + IncludeErrors=true

// Act 3: The Transactionality Test
Input: 10 docs, where doc #5 causes filesystem error mid-walk
Expected:
  - Either all 10 indexed OR none indexed (no partial state)
  - Error message explains what failed

// Act 4: The Path Normalization Ensemble
Input: Doc with RelatedFiles containing:
  - Absolute path: /home/user/repo/pkg/file.go
  - Repo-relative: pkg/file.go
  - Doc-relative: ../../pkg/file.go
  - Cleaned but relative: ./pkg/../pkg/file.go
Expected:
  - All 7 norm_* columns populated
  - Canonical key matches repo-relative when possible
  - Anchor correctly identifies which resolution method worked
```

**Judging Criteria:**
- Transactional integrity (all-or-nothing)
- Graceful degradation for broken docs
- Path normalization correctness
- Performance (can handle 1000+ docs in reasonable time)

---

### 🎪 Contestant #3: "The Polyglot" (Path Normalization)
**Real Name:** `normalization.go` + `paths.Resolver`  
**Stage Name:** Norma the Navigator  
**Hometown:** `internal/workspace` (with guest appearances in `internal/paths`)  
**Age:** Wise elder (based on the battle-tested `paths.Resolver`)

**Background Story:**
Norma grew up in a multilingual household where everyone referred to the same file in different ways. Grandma used absolute paths, Dad preferred repo-relative, Mom was all about doc-relative with `../`, and the kids just wanted tilde expansion to work. After years of family therapy (and fixing reverse-lookup bugs), Norma mastered the art of understanding that `/home/user/repo/pkg/commands/search.go` and `pkg/commands/search.go` and `../../pkg/commands/search.go` are all the same file — you just have to listen carefully to the context.

**Talent:** Multi-anchor path resolution with fallback strategies  
**Special Move:** Produces 7 different representations of the same path, guaranteeing at least one will match at query time

**Performance Act:**
"The Path Reconciliation" — Norma will normalize paths in various formats and prove they all refer to the same file.

**Test Cases:**
```go
// Setup:
// - Repo root: /home/user/docmgr
// - Docs root: /home/user/docmgr/ttmp
// - Doc being processed: /home/user/docmgr/ttmp/2025/12/12/TICKET/design/api.md
// - Target file: /home/user/docmgr/pkg/commands/search.go

// Act 1: Absolute Path
Input: "/home/user/docmgr/pkg/commands/search.go"
Expected norm_canonical: "pkg/commands/search.go"
Expected norm_repo_rel: "pkg/commands/search.go"
Expected norm_abs: "/home/user/docmgr/pkg/commands/search.go"
Expected anchor: "repo"

// Act 2: Repo-Relative
Input: "pkg/commands/search.go"
Expected: Same as Act 1 (proves consistency)

// Act 3: Doc-Relative  
Input: "../../../../../pkg/commands/search.go"
Expected norm_canonical: "pkg/commands/search.go"
Expected norm_doc_rel: "../../../../../pkg/commands/search.go"
Expected anchor: "doc"

// Act 4: The Tilde Expansion
Input: "~/docmgr/pkg/commands/search.go"
Expected norm_abs: "/home/user/docmgr/pkg/commands/search.go"

// Act 5: The Wonky Path
Input: "./pkg/../pkg/./commands/search.go"
Expected norm_clean: "pkg/commands/search.go"

// Grand Finale: The Reverse Lookup Proof
Query: User searches for "pkg/commands/search.go"
Challenge: Match against RelatedFiles entries from Acts 1-5
Expected: ALL of them match via different norm_* columns! ✅
```

**Judging Criteria:**
- Correctness of normalization across anchors
- Consistency (same input → same output)
- Fallback coverage (at least one key always populated)
- Reverse lookup success rate

---

### 🎪 Contestant #4: "The Wizard" (SQL Compiler)
**Real Name:** `query_docs_sql.go`  
**Stage Name:** SQL Sorcerer  
**Hometown:** `internal/workspace`  
**Age:** Mysterious (appears young but speaks ancient SQL dialects)

**Background Story:**
SQL Sorcerer was abandoned as a child in a datacenter and raised by stored procedures. They learned early that the difference between elegant SQL and disastrous SQL is understanding when to use `EXISTS` vs `JOIN`, and that `LIKE` patterns are powerful but dangerous without proper escaping. After witnessing countless SQL injection vulnerabilities and Cartesian product disasters in their youth, they took a vow: "I shall compile safe, correct SQL from structured requests, or I shall return an error."

**Talent:** Compiling structured queries into safe, performant SQL  
**Special Move:** The "EXISTS Cascade" — handles OR semantics across multiple filters without Cartesian explosions

**Performance Act:**
"The Query Compilation Gauntlet" — SQL Sorcerer will compile increasingly complex DocQuery requests into SQL and prove they're correct.

**Test Cases:**
```go
// Act 1: The Simple Ticket Scope
Input: DocQuery{
  Scope: ScopeTicket{TicketID: "MEN-3475"},
  Filters: DocFilters{Status: "active"},
}
Expected SQL (conceptual):
  WHERE d.ticket_id = ? AND d.status = ? AND d.parse_ok = 1

// Act 2: The Reverse Lookup Solo
Input: DocQuery{
  Scope: ScopeRepo,
  Filters: DocFilters{
    RelatedFile: []string{"pkg/commands/search.go"},
  },
}
Expected SQL:
  WHERE EXISTS (
    SELECT 1 FROM related_files rf 
    WHERE rf.doc_id = d.doc_id 
    AND (rf.norm_canonical IN (...) OR rf.norm_repo_rel IN (...) OR ...)
  )
Verification: No SQL injection, all values parameterized ✅

// Act 3: The Multi-File OR Challenge
Input: DocQuery{
  Filters: DocFilters{
    RelatedFile: []string{"file1.go", "file2.go", "file3.go"},
  },
}
Expected: OR semantics (docs match if they reference ANY of the three)
Verification: Count of matched docs is UNION, not intersection

// Act 4: The Directory Prefix Dance
Input: DocQuery{
  Filters: DocFilters{
    RelatedDir: []string{"pkg/commands/", "internal/workspace/"},
  },
}
Expected SQL: LIKE patterns with proper path-boundary handling
Verification: Matches "pkg/commands/search.go" but NOT "pkg/commandserver/main.go"

// Act 5: The Contradictory Query Trap
Input: DocQuery{
  Scope: ScopeTicket{TicketID: "MEN-3475"},
  Filters: DocFilters{Ticket: "MEN-9999"}, // CONTRADICTION!
}
Expected: Hard error before SQL generation ❌

// Grand Finale: The Kitchen Sink
Input: DocQuery{
  Scope: ScopeTicket{TicketID: "MEN-3475"},
  Filters: DocFilters{
    Status: "active",
    DocType: "design",
    TopicsAny: []string{"refactor", "sqlite"},
    RelatedFile: []string{"internal/workspace/query_docs.go"},
    RelatedDir: []string{"internal/workspace/"},
  },
  Options: DocQueryOptions{
    IncludeArchivedPath: false,
    IncludeControlDocs: false,
    OrderBy: OrderByLastUpdated,
    Reverse: true,
  },
}
Expected: Complex SQL with:
  - All filters as AND clauses
  - EXISTS for topics (OR semantics)
  - EXISTS for related file/dir (OR within each type)
  - Proper visibility defaults
  - Correct ORDER BY with DESC
Verification: Manual query against test DB returns expected docs
```

**Judging Criteria:**
- SQL correctness (syntactically valid)
- Safety (no SQL injection vectors)
- Semantic correctness (AND vs OR logic)
- Error handling (contradictions caught early)
- Performance characteristics (uses indexes, avoids Cartesian products)

---

### 🎪 Contestant #5: "The Conductor" (QueryDocs Orchestrator)
**Real Name:** `query_docs.go`  
**Stage Name:** Q the Coordinator  
**Hometown:** `internal/workspace`  
**Age:** Mid-career (recently promoted to avoid N+1 queries)

**Background Story:**
Q started their career as a simple result-iterator, happily scanning rows and calling hydration helpers. Life was simple until The Great Deadlock of 2025, when Q realized that querying for topics *while iterating the main cursor* with `MaxOpenConns=1` was a recipe for disaster. After a dramatic refactoring montage (complete with training music), Q emerged as a batch-hydration expert who now executes a fixed number of queries regardless of result size.

**Talent:** Coordinating query compilation, execution, and hydration without deadlocks  
**Special Move:** "The Batch Blitz" — collects all doc IDs first, then hydrates topics and RelatedFiles in just 2 additional queries

**Performance Act:**
"The Performance Comparison" — Q will execute the same query using both the old N+1 approach and the new batched approach, demonstrating speed and correctness.

**Test Cases:**
```go
// Setup: Index with 100 docs, each having 5 topics and 3 RelatedFiles

// Act 1: The Baseline Query
Input: Simple ticket scope query
Old Approach: 1 base query + (100 * 2 nested queries) = 201 queries
New Approach: 1 base query + 2 batch queries = 3 queries
Expected: Same results, 60x fewer queries ✅

// Act 2: The Error Handling Ballet
Input: DocQuery with IncludeErrors=true under ScopeTicket
Expected:
  - Parse-OK docs return with Doc populated
  - Parse-error docs return with Doc=nil, ReadErr populated
  - All docs properly sorted by OrderBy
  - No hydration attempts for broken docs

// Act 3: The Visibility Toggle
Input: Same query, but IncludeArchivedPath toggled
Round 1: IncludeArchivedPath=false
Round 2: IncludeArchivedPath=true
Expected: Round 2 returns superset of Round 1

// Act 4: The Ordering Proof
Input: DocQuery with OrderBy=OrderByLastUpdated, Reverse=true
Expected: Results sorted newest-first
Verification: Each result's LastUpdated >= next result's LastUpdated

// Grand Finale: The Stress Test
Input: Query matching 1000+ docs with complex filters
Expected:
  - Query completes in < 1 second
  - No connection pool exhaustion
  - No deadlocks
  - Memory usage remains reasonable
```

**Judging Criteria:**
- Query count efficiency (minimize round-trips)
- Correctness of hydrated data
- Error handling robustness
- Performance under load
- No connection pool issues

---

### 🎪 Contestant #6: "The Architect" (SQLite Schema Designer)
**Real Name:** `sqlite_schema.go`  
**Stage Name:** Schema Shapeshifter  
**Hometown:** `internal/workspace`  
**Age:** Timeless (schemas are forever... until migrations)

**Background Story:**
Schema Shapeshifter was once a naive designer who thought "just add a column" was always the answer. After living through several painful migration disasters and witnessing the horrors of unindexed foreign keys, they became obsessed with "minimal viable schema" — storing exactly what's needed for queries, with indexes on exactly the right columns, and no more, no less. Their catchphrase: "Every column tells a story, and every index proves you care about performance."

**Talent:** Schema design that supports query patterns without over-indexing  
**Special Move:** "The Seven Keys" — storing 7 different normalized path representations so reverse lookup always finds a match

**Performance Act:**
"The Schema Interrogation" — Schema Shapeshifter will defend their design choices under questioning.

**Test Cases:**
```sql
-- Act 1: The Foreign Key Integrity Test
-- Delete a doc and verify CASCADE works
INSERT INTO docs (...) VALUES (...);  -- doc_id = 1
INSERT INTO doc_topics (doc_id, ...) VALUES (1, ...);
INSERT INTO related_files (doc_id, ...) VALUES (1, ...);
DELETE FROM docs WHERE doc_id = 1;
-- Expected: doc_topics and related_files rows auto-deleted ✅

-- Act 2: The Index Coverage Analysis
-- Query 1: SELECT * FROM docs WHERE ticket_id = ?
EXPLAIN QUERY PLAN ...;
-- Expected: Uses idx_docs_ticket_id ✅

-- Query 2: SELECT * FROM docs WHERE is_archived_path = 0
EXPLAIN QUERY PLAN ...;
-- Expected: Uses idx_docs_path_tags ✅

-- Query 3: SELECT * FROM related_files WHERE norm_repo_rel = ?
EXPLAIN QUERY PLAN ...;
-- Expected: Uses idx_related_files_norm_repo_rel ✅

-- Act 3: The Pragma Justification
-- Why journal_mode = OFF?
-- Answer: In-memory DB, durability not needed, performance gain is free

-- Why foreign_keys = ON?
-- Answer: Referential integrity prevents orphaned topics/related_files

-- Why synchronous = OFF?
-- Answer: In-memory DB, no disk I/O to synchronize

-- Act 4: The Unique Constraint Proof
-- Try to insert duplicate doc path
INSERT INTO docs (path, ...) VALUES ('/same/path.md', ...);
INSERT INTO docs (path, ...) VALUES ('/same/path.md', ...);
-- Expected: Second insert fails due to UNIQUE constraint ✅

-- Act 5: The Topic Case-Insensitivity Verification
INSERT INTO doc_topics VALUES (1, 'refactor', 'Refactor');
INSERT INTO doc_topics VALUES (1, 'refactor', 'REFACTOR');
-- Expected: Second insert is ignored (PRIMARY KEY on topic_lower) ✅
```

**Judging Criteria:**
- Index coverage for expected queries
- Referential integrity enforcement
- Constraint correctness
- Pragma appropriateness for use case
- Schema minimalism (no unused columns)

---

### 🎪 Contestant #7: "The Packager" (SQLite Exporter)
**Real Name:** `sqlite_export.go`  
**Stage Name:** Artie the Archivist  
**Hometown:** `internal/workspace`  
**Age:** Brand new (born specifically for debugging workflows)

**Background Story:**
Artie came into existence when developers got tired of screenshots and copy-pasted query results when debugging. "Just send me your DB!" became a common refrain, but the in-memory database disappeared as soon as the command exited. Artie solved this by creating shareable SQLite snapshots using the magical `VACUUM INTO` spell. But Artie went further — they added a `README` table with embedded documentation, making each exported DB self-describing. Now anyone can open an exported DB and immediately understand what they're looking at, even without the original repo.

**Talent:** Creating self-documenting shareable database snapshots  
**Special Move:** "The README Embed" — populates a `README` table with all of docmgr's help docs

**Performance Act:**
"The Export-Import Verification" — Artie will export a workspace index, then prove the exported file is complete and self-contained.

**Test Cases:**
```go
// Act 1: The Basic Export
Setup: Workspace with 50 docs indexed
Action: ExportIndexToSQLiteFile(ctx, {OutPath: "test.db"})
Expected:
  - File created at OutPath ✅
  - File is valid SQLite3 format
  - Contains all 50 docs

// Act 2: The Force Overwrite
Setup: test.db already exists
Action: Export with Force=false
Expected: Error (file exists)
Action: Export with Force=true  
Expected: Success (old file replaced) ✅

// Act 3: The README Table Verification
Setup: Export created
Action: Open test.db and query README table
Expected:
  - Row for "__about__.md" with explanation of export
  - Rows for "docmgr-how-to-use.md", "docmgr-how-to-setup.md", etc.
  - All content readable and formatted correctly

// Act 4: The Self-Contained Query Test
Setup: Export created
Action: Open test.db in fresh sqlite3 session
Queries:
  -- List tables
  SELECT name FROM sqlite_master WHERE type='table';
  -- Expected: docs, doc_topics, related_files, README ✅
  
  -- Read embedded doc
  SELECT content FROM README WHERE name='docmgr-how-to-use.md';
  -- Expected: Full markdown content ✅
  
  -- Query docs
  SELECT path, ticket_id, doc_type FROM docs LIMIT 10;
  -- Expected: Same results as in-memory DB ✅

// Act 5: The No-Mkdir Policy Enforcement
Setup: OutPath parent directory doesn't exist
Action: Export to /nonexistent/dir/test.db
Expected: Error (parent dir must exist) ❌

// Grand Finale: The Determinism Test
Action: Export same workspace twice
Expected: Identical row counts, identical README content
Verification: diff <(sqlite3 export1.db ".dump") <(sqlite3 export2.db ".dump")
```

**Judging Criteria:**
- File creation correctness
- VACUUM INTO integrity
- README table completeness
- Self-contained usability
- Error handling (force flag, missing dirs)

---

### 🎪 Contestant #8: "The Sentinel" (Discovery & Context Manager)
**Real Name:** `workspace.go` (DiscoverWorkspace + WorkspaceContext)  
**Stage Name:** Ward the Warden  
**Hometown:** `internal/workspace`  
**Age:** Wise guardian (consolidates years of scattered discovery logic)

**Background Story:**
Ward grew up watching chaos as each command invented its own way to find the docs root, config directory, and repo root. Some commands looked for `.ttmp.yaml`, others searched for `.git`, and a few just gave up and used `pwd`. After one too many "works in my terminal but not in CI" bugs, Ward stepped up to become the canonical "entry point" — the first function every command should call. Ward's motto: "Discovery should happen once, correctly, and consistently."

**Talent:** Best-effort workspace discovery with graceful degradation  
**Special Move:** "The Anchor Trinity" — simultaneously resolves Root, ConfigDir, and RepoRoot using fallback heuristics

**Performance Act:**
"The Discovery Gauntlet" — Ward will be placed in various directory contexts and must correctly discover workspace anchors.

**Test Cases:**
```go
// Act 1: The Ideal Case
Context: CWD is repo root, .ttmp.yaml exists, .git exists
Expected:
  Root: <repo>/ttmp ✅
  ConfigDir: <repo> ✅
  RepoRoot: <repo> ✅
  Config: Successfully loaded

// Act 2: The Subdirectory Challenge
Context: CWD is <repo>/pkg/commands/
Expected:
  Root: <repo>/ttmp (walks up to find it)
  ConfigDir: <repo>
  RepoRoot: <repo>

// Act 3: The Missing Config Graceful Degradation
Context: .ttmp.yaml is malformed or missing
Expected:
  Root: Resolved via fallback heuristics ✅
  ConfigDir: Best-effort guess
  RepoRoot: <repo>
  Config: nil (but no crash)

// Act 4: The Explicit Override
Input: DiscoverOptions{RootOverride: "/explicit/path/to/docs"}
Expected:
  Root: /explicit/path/to/docs ✅
  ConfigDir: <parent of RootOverride>
  RepoRoot: <found via git>

// Act 5: The Invariant Enforcement
Setup: Mock missing repoRoot
Action: NewWorkspaceFromContext({Root: "...", ConfigDir: "...", RepoRoot: ""})
Expected: Error ("missing RepoRoot") ❌

// Grand Finale: The InitIndex Integration
Setup: DiscoverWorkspace succeeds
Action: ws.InitIndex(ctx, BuildIndexOptions{})
Expected:
  - DB initialized ✅
  - Docs ingested using discovered Root
  - Path normalization uses discovered anchors
  - QueryDocs immediately usable
```

**Judging Criteria:**
- Discovery correctness across environments
- Fallback robustness (doesn't crash on missing config)
- Override behavior (explicit beats heuristic)
- Invariant enforcement (required fields must be present)
- Integration with downstream components

---

## The Grand Finale Performance

**Group Act: "The Integration Spectacular"**

All contestants perform together in a full end-to-end scenario:

```go
// Setup
testRoot := createTestWorkspace(t, WorkspaceStructure{
  Docs: []TestDoc{
    {Path: "2025/12/12/TICKET-A/index.md", Frontmatter: {...}, Body: "..."},
    {Path: "2025/12/12/TICKET-A/design/api.md", Frontmatter: {...}},
    {Path: "2025/12/12/TICKET-A/archive/old.md", Frontmatter: {...}},
    {Path: "2025/12/12/TICKET-A/.meta/notes.md", Frontmatter: {...}},  // should be skipped
    {Path: "2025/12/12/TICKET-B/broken.md", Frontmatter: MALFORMED},
  },
})

// Act 1: Discovery (Ward the Warden)
ws, err := workspace.DiscoverWorkspace(ctx, workspace.DiscoverOptions{
  RootOverride: testRoot,
})
assert.NoError(t, err)

// Act 2: Index Building (Ingrid the Indexer + DJ Skippy + Norma)
err = ws.InitIndex(ctx, workspace.BuildIndexOptions{IncludeBody: true})
assert.NoError(t, err)

// Act 3: Query Compilation & Execution (SQL Sorcerer + Q the Coordinator)
result, err := ws.QueryDocs(ctx, workspace.DocQuery{
  Scope: workspace.Scope{Kind: workspace.ScopeTicket, TicketID: "TICKET-A"},
  Filters: workspace.DocFilters{
    DocType: "design",
  },
  Options: workspace.DocQueryOptions{
    IncludeArchivedPath: false,  // Should exclude archive/old.md
    IncludeErrors: false,         // Should exclude broken.md
  },
})
assert.NoError(t, err)
assert.Len(t, result.Docs, 1)  // Only design/api.md
assert.Equal(t, "design/api.md", filepath.Base(result.Docs[0].Path))

// Act 4: Reverse Lookup (Norma + SQL Sorcerer)
result, err = ws.QueryDocs(ctx, workspace.DocQuery{
  Scope: workspace.Scope{Kind: workspace.ScopeRepo},
  Filters: workspace.DocFilters{
    RelatedFile: []string{"pkg/commands/search.go"},
  },
})
assert.NoError(t, err)
// Verify matches docs that reference search.go in various formats

// Act 5: Error Discovery (Ingrid + Q)
result, err = ws.QueryDocs(ctx, workspace.DocQuery{
  Scope: workspace.Scope{Kind: workspace.ScopeTicket, TicketID: "TICKET-B"},
  Options: workspace.DocQueryOptions{IncludeErrors: true},
})
assert.NoError(t, err)
assert.Len(t, result.Docs, 1)
assert.NotNil(t, result.Docs[0].ReadErr)  // broken.md included with error

// Act 6: Export (Artie the Archivist + Schema Shapeshifter)
exportPath := filepath.Join(testRoot, "export.db")
err = ws.ExportIndexToSQLiteFile(ctx, workspace.ExportSQLiteOptions{
  OutPath: exportPath,
  Force: true,
})
assert.NoError(t, err)

// Verify exported DB is self-contained
exportDB, err := sql.Open("sqlite3", exportPath)
assert.NoError(t, err)
var readmeCount int
err = exportDB.QueryRow("SELECT COUNT(*) FROM README").Scan(&readmeCount)
assert.NoError(t, err)
assert.Greater(t, readmeCount, 5)  // Has embedded docs

// Final verification: Query exported DB
var docCount int
err = exportDB.QueryRow("SELECT COUNT(*) FROM docs WHERE parse_ok = 1").Scan(&docCount)
assert.NoError(t, err)
assert.Equal(t, 3, docCount)  // index.md, design/api.md, archive/old.md (broken.md and .meta skipped)
```

**Standing Ovation Criteria:**
- All contestants perform their specialized roles correctly
- No contestant blocks or deadlocks another
- End-to-end result matches specification
- Exported artifact is shareable and self-contained
- Performance completes in reasonable time (<1 second for small test workspace)

---

## Usage Examples

### Running Individual Performances

Each contestant can be tested in isolation:

```bash
# Test DJ Skippy (Skip Policy)
go test ./internal/workspace -run TestSkipPolicy -v

# Test Ingrid (Index Builder)
go test ./internal/workspace -run TestIndexBuilder -v

# Test Norma (Path Normalization)
go test ./internal/workspace -run TestNormalization -v

# Test SQL Sorcerer (Query Compiler)
go test ./internal/workspace -run TestCompileDocQuery -v

# Test Q (QueryDocs)
go test ./internal/workspace -run TestQueryDocs -v

# Test Schema Shapeshifter
go test ./internal/workspace -run TestSQLiteSchema -v

# Test Artie (Export)
go test ./internal/workspace -run TestExportSQLite -v

# Test Ward (Discovery)
go test ./internal/workspace -run TestDiscoverWorkspace -v
```

### Running the Grand Finale

```bash
# Full integration test
go test ./internal/workspace -run TestWorkspaceIntegration -v

# Real-world scenario test using test-scenarios
DOCMGR_PATH=$(go build -o /tmp/docmgr-talent ./cmd/docmgr && echo /tmp/docmgr-talent) \
bash test-scenarios/testing-doc-manager/run-all.sh /tmp/talent-show-results
```

### Creating Custom Performance Tests

```go
func TestMyCustomPerformance(t *testing.T) {
    // Setup
    ws := createTestWorkspace(t)
    
    // Challenge the contestants
    result, err := ws.QueryDocs(ctx, workspace.DocQuery{
        // Your challenge here
    })
    
    // Judge the performance
    assert.NoError(t, err)
    assert.Equal(t, expectedResult, result)
}
```

---

## Judging Rubric

For each contestant, we evaluate:

### Technical Correctness (40%)
- Does it do what it claims to do?
- Are edge cases handled correctly?
- Are errors meaningful and actionable?

### Performance (20%)
- Time complexity reasonable?
- Memory usage appropriate?
- No unnecessary round-trips or allocations?

### Robustness (20%)
- Graceful degradation on errors?
- No crashes or panics?
- Transaction safety where needed?

### Integration (10%)
- Plays well with other contestants?
- Clear interface boundaries?
- No hidden dependencies?

### Style & Elegance (10%)
- Code readability
- Naming clarity
- Comments where needed
- Follows spec requirements

---

## Related

- **Design Spec:** `design/01-workspace-sqlite-repository-api-design-spec.md`
- **Implementation Diary:** `reference/15-diary.md`
- **Code Review Guide:** `analysis/03-code-review-guide-senior.md`
- **Testing Strategy:** `analysis/02-testing-strategy-integration-first.md`
