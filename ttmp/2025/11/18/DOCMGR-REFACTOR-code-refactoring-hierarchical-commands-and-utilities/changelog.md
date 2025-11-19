# Changelog

## 2025-11-18

- Initial workspace created


## 2025-11-18

Added cmd/docmgr/cmds hierarchy for Cobra commands and moved config/workspace/template helpers under internal/.

### Related Files

- cmd/docmgr/cmds/root.go — Cobra hierarchy
- cmd/docmgr/main.go — Root now delegates to cmds
- internal/templates/templates.go — template helpers moved
- internal/workspace/config.go — config resolution moved

