import { settingDocumentSchema, type SettingDocument } from '@/features/settings/schema'

// settingsMock mirrors the shape the real service will return. Every setting references dictionary
// keys instead of literal text, and the whole document is validated once at module load so a malformed
// fixture fails loudly during development instead of rendering a broken page.
const fixture = {
  categories: [
    {
      id: 'editor',
      labelKey: 'settings.categories.editor',
      settings: [
        {
          id: 'editor.fontSize',
          type: 'input',
          labelKey: 'settings.items.editorFontSize.label',
          descriptionKey: 'settings.items.editorFontSize.description',
          defaultValue: '14',
          placeholder: '14',
        },
        {
          id: 'editor.wordWrap',
          type: 'switch',
          labelKey: 'settings.items.editorWordWrap.label',
          descriptionKey: 'settings.items.editorWordWrap.description',
          defaultValue: true,
        },
        {
          id: 'editor.tabSize',
          type: 'select',
          labelKey: 'settings.items.editorTabSize.label',
          descriptionKey: 'settings.items.editorTabSize.description',
          defaultValue: '4',
          options: [
            { value: '2', labelKey: '2' },
            { value: '4', labelKey: '4' },
            { value: '8', labelKey: '8' },
          ],
        },
        {
          id: 'editor.minimap',
          type: 'switch',
          labelKey: 'settings.items.editorMinimap.label',
          descriptionKey: 'settings.items.editorMinimap.description',
          defaultValue: false,
        },
      ],
    },
    {
      id: 'terminal',
      labelKey: 'settings.categories.terminal',
      settings: [
        {
          id: 'terminal.shell',
          type: 'select',
          labelKey: 'settings.items.terminalShell.label',
          descriptionKey: 'settings.items.terminalShell.description',
          defaultValue: 'bash',
          options: [
            { value: 'bash', labelKey: 'bash' },
            { value: 'zsh', labelKey: 'zsh' },
            { value: 'powershell', labelKey: 'PowerShell' },
          ],
        },
        {
          id: 'terminal.fontSize',
          type: 'input',
          labelKey: 'settings.items.terminalFontSize.label',
          descriptionKey: 'settings.items.terminalFontSize.description',
          defaultValue: '13',
          placeholder: '13',
        },
        {
          id: 'terminal.cursorBlink',
          type: 'switch',
          labelKey: 'settings.items.terminalCursorBlink.label',
          descriptionKey: 'settings.items.terminalCursorBlink.description',
          defaultValue: true,
        },
      ],
    },
    {
      id: 'extensions',
      labelKey: 'settings.categories.extensions',
      settings: [
        {
          id: 'extensions.autoUpdate',
          type: 'switch',
          labelKey: 'settings.items.extensionsAutoUpdate.label',
          descriptionKey: 'settings.items.extensionsAutoUpdate.description',
          defaultValue: true,
        },
        {
          id: 'extensions.registry',
          type: 'input',
          labelKey: 'settings.items.extensionsRegistry.label',
          descriptionKey: 'settings.items.extensionsRegistry.description',
          defaultValue: 'https://registry.lightpc.local',
          placeholder: 'https://registry.example.com',
        },
      ],
    },
  ],
}

const parsed = settingDocumentSchema.safeParse(fixture)
if (!parsed.success) {
  throw new Error(`settings mock does not match the schema: ${parsed.error.message}`)
}

export const settingsMock: SettingDocument = parsed.data
