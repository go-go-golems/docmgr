# Changelog

## 2025-12-01

- Initial workspace created


## 2025-12-01

Created bug ticket with reproduction steps and root cause analysis. Bug confirmed: when ticket title contains ticket identifier, directory names have duplicate ticket identifiers (e.g., TEST-9999-test-9999-...). Root cause: slug computation includes ticket identifier from title, then path template combines {{TICKET}} with {{SLUG}}.


## 2025-12-01

Fixed ticket name duplication bug. Added StripTicketFromTitle helper function in pkg/utils/slug.go to remove ticket identifier patterns from title before slugifying. Updated ticket_move.go and create_ticket.go to use the helper. Tested: tickets with titles containing ticket identifiers now create correct directory names (e.g., TEST-8888-another-test instead of TEST-8888-test-8888-another-test).

