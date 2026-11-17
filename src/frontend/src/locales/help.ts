// Help holds the documentation browser shown by the help dialog and its command palette.
export const help = {
  title: 'Help',
  description: 'Browse the documentation and search for answers.',
  switcher: {
    title: 'Documentation',
    versions: ['v1.0.1', 'v1.1.0-alpha', 'v2.0.0-beta1'],
  },
  search: {
    label: 'Search documents...',
    empty: 'No documents found.',
    placeholder: 'Select a document',
  },
  placeholder: {
    pending: (section: string) => `Open a document in ${section} to read it here.`,
  },
  sections: {
    gettingStarted: {
      title: 'Getting Started',
      items: {
        installation: 'Installation',
        structure: 'Project Structure',
        configuration: 'Configuration',
      },
    },
    build: {
      title: 'Build Your Application',
      items: {
        routing: 'Routing',
        dataFetching: 'Data Fetching',
        rendering: 'Rendering',
        styling: 'Styling',
      },
    },
    reference: {
      title: 'API Reference',
      items: {
        components: 'Components',
        functions: 'Functions',
        cli: 'CLI',
      },
    },
  },
  command: {
    title: 'Search',
    description: 'Search the documentation.',
    placeholder: 'Search documents...',
    empty: 'No results found.',
  },
}
