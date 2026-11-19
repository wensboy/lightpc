import { Input } from '@/components/ui/input'
import { SettingField } from '@/features/settings/components/setting/SettingField'
import type { SettingField as SettingFieldModel, SettingValue } from '@/features/settings/schema'

type SettingInputProps = {
  field: SettingFieldModel
  onChange: (value: SettingValue) => void
}

// SettingInput renders a free text field. The width is capped so the list keeps the compact VSCode
// proportions instead of stretching across the panel.
export function SettingInput({ field, onChange }: SettingInputProps) {
  return (
    <SettingField
      field={field}
      control={
        <Input
          id={field.path}
          value={String(field.value)}
          onChange={(event) => onChange(event.target.value)}
          className="w-56"
        />
      }
    />
  )
}
