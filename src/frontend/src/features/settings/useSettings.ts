import { useCallback, useState } from 'react'
import type { SettingDocument, SettingValue } from '@/features/settings/schema'
import { loadSettings } from '@/mocks'

// initialValues flattens every setting default into a single id keyed map, which is what the renderer
// reads from.
function initialValues(document: SettingDocument): Record<string, SettingValue> {
  const values: Record<string, SettingValue> = {}
  for (const category of document.categories) {
    for (const setting of category.settings) {
      values[setting.id] = setting.defaultValue
    }
  }
  return values
}

export type SettingsState = {
  document: SettingDocument
  values: Record<string, SettingValue>
  setValue: (id: string, value: SettingValue) => void
}

// useSettings owns the settings document and the edited values. The document is loaded once; every edit
// replaces the single entry so React sees a new object.
export function useSettings(): SettingsState {
  const [document] = useState<SettingDocument>(loadSettings)
  const [values, setValues] = useState<Record<string, SettingValue>>(() => initialValues(document))

  const setValue = useCallback((id: string, value: SettingValue) => {
    setValues((previous) => ({ ...previous, [id]: value }))
  }, [])

  return { document, values, setValue }
}
