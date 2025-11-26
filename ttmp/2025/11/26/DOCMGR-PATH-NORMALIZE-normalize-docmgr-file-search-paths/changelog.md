# Changelog

## 2025-11-26

- Initial workspace created
- Captured current relate/search behavior in `reference/01-path-handling-analysis.md`
- Drafted canonical normalization + fuzzy-search plan in `design/01-path-normalization-canonicalization.md`
- Implemented `internal/paths` resolver + canonical storage, wired `relate`/`search` to normalization + fuzzy matching, and added regression coverage (Go tests + `14-path-normalization.sh` playbook)

