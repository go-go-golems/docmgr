# Changelog

## 2025-11-19

- Initial workspace created


## 2025-11-19

Fixed duplicate frontmatter issue in templates. When templates contain placeholders like {{TITLE}}, the YAML parser fails because placeholders aren't valid YAML. Added fallback logic to manually extract body by finding the closing --- delimiter when library parsing fails.

### Related Files

- /home/manuel/workspaces/2025-11-18/code-review-docmgr/docmgr/internal/templates/templates.go — Added manual frontmatter stripping fallback


## 2025-11-19

Implementation complete: Fixed template frontmatter parsing with hybrid approach

