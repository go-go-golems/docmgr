# Tasks

## TODO

- [x] Add tasks here

- [x] Survey verbs and typed outputs
- [x] Design template resolution: templates//.templ
- [x] Define typed data contracts per verb
- [x] Add rendering hook after classic output
- [x] Implement for 3 verbs: list docs, list tickets, doctor
- [x] Implement canonical template path resolution (no fallbacks)
- [x] Update list_docs Run to build struct and render templates/doc/list.templ
- [x] Update list_tickets Run to build struct and render templates/list/tickets.templ
- [x] Update doctor Run to build struct and render templates/doctor.templ
- [x] Add safe template FuncMap helpers for postfix templates
- [x] Create example templates under ttmp/templates for verification
- [x] Update docs to describe postfix templates (human-only)
- [x] Add changelog entry after implementation
- [ ] Add template validation tooling (docmgr template validate command to check syntax before runtime)
- [x] Document template data contracts more thoroughly (explicit documentation of available fields per verb)
- [ ] Add template debugging features (--debug-template flag showing resolved path, data, and errors)
- [x] Create comprehensive template examples (advanced patterns: nested loops, conditionals, complex transformations)
- [ ] Consider adding templates to more verbs (status, search, guidelines, vocab list, etc.)
- [ ] Evaluate template composition/inheritance patterns (if needed based on usage feedback)
- [x] Add unit tests for template FuncMap helpers (especially countBy with various data types)
- [x] Add vocabulary entries for new topics used in tickets (cli, templates, glaze)
- [ ] Implement template support for status command - build template data struct with TicketsTotal, TicketsStale, DocsTotal, doc type counts, and per-ticket status info. Render templates/status.templ after human output.
- [ ] Implement template support for search command - refactor to collect results before printing, build template data struct with Query, Results (Path, Title, Ticket, Snippet), TotalResults. Render templates/doc/search.templ after human output.
- [ ] Implement template support for tasks list command - build template data struct with TotalTasks, OpenTasks, DoneTasks, Tasks array. Render templates/tasks/list.templ after human output.
- [ ] Implement template support for vocab list command - build template data struct with Category filter and vocabulary items by category (Topics, DocTypes, Intent, Status). Render templates/vocab/list.templ after human output.
- [ ] Implement template support for guidelines command - build template data struct with DocType, GuidelineText, Source, Path. Render templates/doc/guidelines.templ after human output. (Low priority - mostly static content)
- [ ] Replace custom FuncMap helpers with glazed templating helpers (sprig and co) - investigate glazed template formatter FuncMap, migrate existing templates to use standard helpers, update documentation
- [x] Add --print-template-schema flag to all verbs with templating - output JSON schema and documentation for template data structures using introspection, show available fields, types, and example values
- [ ] Create documentation/tutorial for postfix templates - user guide explaining how to create templates, available data structures, function helpers, common patterns, and examples
- [ ] Create example template for status (ttmp/templates/status.templ)
- [ ] Create example template for tasks list (ttmp/templates/tasks/list.templ)
- [ ] Create example template for search (ttmp/templates/doc/search.templ)
- [ ] Create example template for vocab list (ttmp/templates/vocab/list.templ)
- [ ] Create example template for guidelines (ttmp/templates/doc/guidelines.templ)
- [ ] Extend data contracts reference with status/tasks/search/vocab/guidelines
- [ ] Add integration tests: --print-template-schema prints only schema (docs/tickets/doctor)
- [ ] Document --print-template-schema usage in how-to-use and tutorial
- [ ] Migrate FuncMap in internal/templates/verb_output.go to Glazed/Sprig helpers; update templates
- [ ] Implement docmgr template validate (lint templates; pre-runtime checks)
- [ ] Add --debug-template flag: show resolved path, data keys preview, and errors
- [ ] Add CI step: validate templates and run schema-only tests
- [ ] Evaluate template composition/inheritance patterns; propose recommendations
- [ ] Standardize template field names across verbs; write conventions doc
