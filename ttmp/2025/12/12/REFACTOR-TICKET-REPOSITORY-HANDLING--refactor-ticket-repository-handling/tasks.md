# Tasks

## TODO

- [ ] Add tasks here

- [x] Implement workspace.Workspace core type + context + constructors (DiscoverWorkspace/NewWorkspaceFromContext) — Spec §4.1, §5.1, §11.1
- [x] Define SQLite schema + indexes for docs/topics/related_files (including multi-key normalization columns) — Spec §9.1–§9.2
- [x] Implement ingest-time skip rules + tagging (archive/scripts/sources/control-docs/.meta/_*/index.md) — Spec §6, §12.3
- [x] Implement ingestion walker: parse frontmatter, store parse_ok/parse_err, extract topics + related_files, store path-category tags — Spec §6, §7.1, §9.2
- [x] Implement path normalization pipeline using paths.Resolver; store canonical + fallback keys for RelatedFiles; document fallback matching strategy — Spec §5.1, §7.3, §9.2, §12.1
- [ ] Implement QueryDocs API + SQL compiler (scope + filters + options), including OR semantics for RelatedFile[]/RelatedDir[] — Spec §5.2, §10.1–§10.4
- [ ] Implement diagnostics contract in QueryDocs (emit core.Taxonomy for skips/parse errors/normalization issues) — Spec §7, §8, §10.6
- [ ] Port command: list docs -> Workspace.QueryDocs (match existing semantics; remove ad-hoc filepath.Walk) — Spec §11.2.1
- [ ] Port command: search -> Workspace.QueryDocs for metadata + reverse lookup; keep content search as post-filter (FTS deferred) — Spec §11.2.2, §13 (FTS deferred)
- [ ] Port command: doctor -> QueryDocs(IncludeErrors/IncludeDiagnostics) + unify RelatedFiles existence checks with Workspace normalization — Spec §11.2.3, §7.3
- [ ] Port command: relate -> use Workspace resolver + doc lookup (ScopeDoc/ScopeTicket) so normalization matches index — Spec §11.2.4, §5.1
- [ ] Cleanup: remove/retire duplicated walkers and helpers (findTicketDirectory, ad-hoc Walk loops); enforce single canonical traversal via Workspace — Spec §11.3
