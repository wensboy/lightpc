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
  },
  empty: {
    title: 'Nothing here yet',
    description: 'This view is reserved for the console panel. The content lands with its own feature.',
  },
}
