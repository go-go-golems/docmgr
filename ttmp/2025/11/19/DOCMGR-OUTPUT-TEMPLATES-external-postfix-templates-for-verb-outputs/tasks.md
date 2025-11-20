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
- [ ] Document template data contracts more thoroughly (explicit documentation of available fields per verb)
- [ ] Add template debugging features (--debug-template flag showing resolved path, data, and errors)
- [ ] Create comprehensive template examples (advanced patterns: nested loops, conditionals, complex transformations)
- [ ] Consider adding templates to more verbs (status, search, guidelines, vocab list, etc.)
- [ ] Evaluate template composition/inheritance patterns (if needed based on usage feedback)
- [ ] Add unit tests for template FuncMap helpers (especially countBy with various data types)
