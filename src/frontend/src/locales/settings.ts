// Settings holds the shell text of the settings dialog. Categories and items mirror the schema keys
// used by the settings document, so adding a setting means adding its labels here.
export const settings = {
  title: 'Settings',
  description: 'Customize your settings here.',
  categories: {
    editor: 'Editor',
    terminal: 'Terminal',
    extensions: 'Extensions',
  },
  items: {
    editorFontSize: {
      label: 'Font Size',
      description: 'Controls the font size in pixels.',
    },
    editorWordWrap: {
      label: 'Word Wrap',
      description: 'Wrap lines that exceed the viewport width.',
    },
    editorTabSize: {
      label: 'Tab Size',
      description: 'The number of spaces a tab is equal to.',
    },
    editorMinimap: {
      label: 'Minimap',
      description: 'Show a code overview on the right edge of the editor.',
    },
    terminalShell: {
      label: 'Default Shell',
      description: 'Shell used when opening a new terminal.',
    },
    terminalFontSize: {
      label: 'Font Size',
      description: 'Controls the terminal font size in pixels.',
    },
    terminalCursorBlink: {
      label: 'Cursor Blinking',
      description: 'Blink the terminal cursor.',
    },
    extensionsAutoUpdate: {
      label: 'Auto Update',
      description: 'Update installed extensions automatically.',
    },
    extensionsRegistry: {
      label: 'Registry',
      description: 'Source used to look up extensions.',
    },
  },
}
