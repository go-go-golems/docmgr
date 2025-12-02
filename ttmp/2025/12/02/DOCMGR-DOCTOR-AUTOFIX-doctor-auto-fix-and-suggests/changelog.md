# Changelog

## 2025-12-02
- Ticket created to track adding suggest/auto-fix to doctor.
- Context: `validate frontmatter` already supports suggest/auto-fix with backups; doctor currently only reports issues.
- Planned: share fix generator (`pkg/commands/validate_frontmatter.go`, `pkg/frontmatter/frontmatter.go`), add flags to doctor, update rules/docs/smokes.
