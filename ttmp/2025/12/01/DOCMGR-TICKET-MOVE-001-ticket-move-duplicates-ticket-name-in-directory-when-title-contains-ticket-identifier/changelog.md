# Changelog

## 2025-12-01

- Initial workspace created


## 2025-12-01

Created bug ticket with reproduction steps and root cause analysis. Bug confirmed: when ticket title contains ticket identifier, directory names have duplicate ticket identifiers (e.g., TEST-9999-test-9999-...). Root cause: slug computation includes ticket identifier from title, then path template combines {{TICKET}} with {{SLUG}}.


## 2025-12-01

Fixed ticket name duplication bug. Added StripTicketFromTitle helper function in pkg/utils/slug.go to remove ticket identifier patterns from title before slugifying. Updated ticket_move.go and create_ticket.go to use the helper. Tested: tickets with titles containing ticket identifiers now create correct directory names (e.g., TEST-8888-another-test instead of TEST-8888-test-8888-another-test).


## 2025-12-02

Fixed ticket name duplication bug: (1) Added SlugifyTitleForTicket helper to strip ticket identifier from titles before slugifying, (2) Changed default path template to use double dash (-- instead of -) between ticket and slug for better visual separation, (3) Applied fix in both create-ticket and ticket-move commands, (4) Added comprehensive unit tests, (5) Updated documentation references. Fix verified: tickets with titles containing ticket identifiers now create correct directory names (e.g., TEST-9999--test-ticket-with-ticket-in-title instead of TEST-9999-test-9999-test-ticket-with-ticket-in-title).


## 2025-12-02

Fix implemented and verified: ticket name duplication bug resolved with SlugifyTitleForTicket helper and double-dash separator in path template

