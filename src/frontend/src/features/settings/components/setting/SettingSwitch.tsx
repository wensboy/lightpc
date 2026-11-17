import { Switch } from '@/components/ui/switch'
import { SettingField } from '@/features/settings/components/setting/SettingField'
import type { SettingSchema } from '@/features/settings/schema'

type SettingSwitchProps = {
  schema: Extract<SettingSchema, { type: 'switch' }>
  categoryLabel: string
  value: boolean
  onChange: (value: boolean) => void
}

// SettingSwitch renders a boolean setting. It is controlled: the value comes from the caller and every
// toggle is reported through onChange.
export function SettingSwitch({ schema, categoryLabel, value, onChange }: SettingSwitchProps) {
  return (
    <SettingField
      id={schema.id}
      categoryLabel={categoryLabel}
      labelKey={schema.labelKey}
      descriptionKey={schema.descriptionKey}
      control={<Switch id={schema.id} checked={value} onCheckedChange={onChange} />}
    />
  )
}
