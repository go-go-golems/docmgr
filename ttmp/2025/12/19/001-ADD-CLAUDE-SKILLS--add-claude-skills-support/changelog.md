# Changelog

## 2025-12-19

- Initial workspace created


## 2025-12-19

Created analysis document and diary documenting codebase exploration. Analyzed command structure, document model, workspace discovery, and document walking patterns. Identified implementation approach for skills feature.


## 2025-12-19

Created implementation documentation: codebase architecture guide and CLI verb implementation guide in pkg/doc/, following Glazed documentation style guide


## 2025-12-19

Enhanced codebase architecture documentation: added full explanatory paragraphs, visual diagrams, bullet points, and beginner-friendly content. Document expanded from 462 to 1,107 lines with detailed explanations of workspace discovery, SQLite indexing, ticket structure, document model, frontmatter parsing, document walking, query system, and design decisions.


## 2025-12-19

Updated analysis: removed skill discovery concept (same as list), added path filtering support (--file and --dir flags) to skill list command. Skills will use QueryDocs with DocType filter, leveraging existing RelatedFile/RelatedDir filters like doc search.


## 2025-12-19

Created design-doc implementation plan for skills (DocType=skill, WhatFor/WhenToUse, skill list with --file/--dir filtering via existing query layer). Added concrete implementation tasks to tasks.md and documented the key constraint: QueryDocs hydrates from SQLite, so skill fields must be indexed.


## 2025-12-19

Step 8: Added WhatFor and WhenToUse fields to Document model (commit e8e0341)

### Related Files

- /home/manuel/workspaces/2025-12-19/add-docmgr-skills/docmgr/pkg/models/document.go — Added optional WhatFor/WhenToUse fields for skills preamble

