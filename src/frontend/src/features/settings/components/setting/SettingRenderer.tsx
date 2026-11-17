import { SettingInput } from '@/features/settings/components/setting/SettingInput'
import { SettingSelect } from '@/features/settings/components/setting/SettingSelect'
import { SettingSwitch } from '@/features/settings/components/setting/SettingSwitch'
import type { SettingSchema, SettingValue } from '@/features/settings/schema'

type SettingRendererProps = {
  schema: SettingSchema
  categoryLabel: string
  value: SettingValue
  onChange: (value: SettingValue) => void
}

// SettingRenderer picks the control for one setting. The switch on schema.type narrows the schema and
// the stored value together, so each control receives exactly the types it declares.
export function SettingRenderer({ schema, categoryLabel, value, onChange }: SettingRendererProps) {
  switch (schema.type) {
    case 'switch':
      return (
        <SettingSwitch
          schema={schema}
          categoryLabel={categoryLabel}
          value={value === true}
          onChange={onChange}
        />
      )
    case 'input':
      return (
        <SettingInput
          schema={schema}
          categoryLabel={categoryLabel}
          value={typeof value === 'string' ? value : ''}
          onChange={onChange}
        />
      )
    case 'select':
      return (
        <SettingSelect
          schema={schema}
          categoryLabel={categoryLabel}
          value={typeof value === 'string' ? value : ''}
          onChange={onChange}
        />
      )
  }
}
