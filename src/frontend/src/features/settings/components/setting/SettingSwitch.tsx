import { Switch } from '@/components/ui/switch'
import { SettingField } from '@/features/settings/components/setting/SettingField'
import type { SettingField as SettingFieldModel, SettingValue } from '@/features/settings/schema'

type SettingSwitchProps = {
  field: SettingFieldModel
  onChange: (value: SettingValue) => void
}

// SettingSwitch renders a boolean field. It is controlled: the value comes from the field and every
// toggle is reported through onChange.
export function SettingSwitch({ field, onChange }: SettingSwitchProps) {
  return (
    <SettingField
      field={field}
      control={
        <Switch
          id={field.path}
          checked={field.value === true}
          onCheckedChange={onChange}
        />
      }
    />
  )
}
