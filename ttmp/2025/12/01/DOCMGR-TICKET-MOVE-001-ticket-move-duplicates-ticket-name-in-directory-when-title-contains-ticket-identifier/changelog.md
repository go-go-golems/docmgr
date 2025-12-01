# Changelog

## 2025-12-01

- Initial workspace created


## 2025-12-01

Created bug ticket with reproduction steps and root cause analysis. Bug confirmed: when ticket title contains ticket identifier, directory names have duplicate ticket identifiers (e.g., TEST-9999-test-9999-...). Root cause: slug computation includes ticket identifier from title, then path template combines {{TICKET}} with {{SLUG}}.

