# Todo

> Project · Todo

Work that is agreed on but not started. Keep entries short and remove them once they land, so the list
never becomes an archive.

## Backend

- [ ] Settings service: expose the theme palettes already resolved, so the frontend stops converting keys.
- [ ] Store: support a transactional write so two services cannot interleave a flush.
- [ ] Log: honour the configured level per sink instead of the single global level.

## Frontend

- [ ] Settings: group fields by section once a document declares more than one level of nesting.
- [ ] Notifications: report a dismissed item to the backend so a restart does not resurrect it.
- [ ] Command palette: index the documentation so a command can jump straight to a page.

## Docs

- [ ] Fill every placeholder under `manual/` and `reference/`.
- [ ] Add a second version set once the format is stable, to exercise the version switcher.
