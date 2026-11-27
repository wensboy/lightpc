// Console holds the text of the console sheet: its navigation labels, the log panel and the empty state.
export const consoleText = {
  title: 'Console',
  description: 'Inspect the running application from one place.',
  sections: {
    log: 'Log',
    performance: 'Performance',
  },
  log: {
    listener: 'Log client',
    filter: {
      level: 'Level',
      pattern: 'Filter (e.g. text, excludeText…)',
      patternLabel: 'Filter logs',
    },
    levels: {
      debug: 'DEBUG',
      info: 'INFO',
      warn: 'WARN',
      error: 'ERROR',
    },
    actions: {
      clear: 'Clear logs',
      autoScroll: 'Auto scroll',
    },
    // Each panel owns its empty state: the log view explains its own filters, and a panel that has no implementation
    // yet gets the pending text instead of borrowing one. Sharing a single empty block made the copy wrong for every
    // panel but the first.
    empty: {
      title: 'No log records',
      description:
        'Records appear here as the backend writes them. Switch the source or clear the filters if you expected output.',
    },
  },
  pending: {
    title: 'Nothing here yet',
    description: 'This panel lands with its own feature.',
  },
}
