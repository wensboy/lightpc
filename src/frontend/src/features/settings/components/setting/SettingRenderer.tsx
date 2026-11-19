import { SettingInput } from '@/features/settings/components/setting/SettingInput'
import { SettingSelect } from '@/features/settings/components/setting/SettingSelect'
import { SettingSwitch } from '@/features/settings/components/setting/SettingSwitch'
import type { SettingField, SettingValue } from '@/features/settings/schema'

type SettingRendererProps = {
  field: SettingField
  onChange: (value: SettingValue) => void
}

// SettingRenderer maps a field to its control. The control comes from the schema, so adding a field to
// the document never requires a change here.
export function SettingRenderer({ field, onChange }: SettingRendererProps) {
  switch (field.control) {
    case 'switch':
      return <SettingSwitch field={field} onChange={onChange} />
    case 'select':
      return <SettingSelect field={field} onChange={onChange} />
    case 'input':
      return <SettingInput field={field} onChange={onChange} />
  }
}
