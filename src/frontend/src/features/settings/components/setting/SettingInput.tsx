import { Input } from '@/components/ui/input'
import { SettingField } from '@/features/settings/components/setting/SettingField'
import type { SettingSchema } from '@/features/settings/schema'

type SettingInputProps = {
  schema: Extract<SettingSchema, { type: 'input' }>
  categoryLabel: string
  value: string
  onChange: (value: string) => void
}

// SettingInput renders a free text setting. The width is capped so the field keeps the compact VSCode
// proportions instead of stretching across the panel.
export function SettingInput({ schema, categoryLabel, value, onChange }: SettingInputProps) {
  return (
    <SettingField
      id={schema.id}
      categoryLabel={categoryLabel}
      labelKey={schema.labelKey}
      descriptionKey={schema.descriptionKey}
      control={
        <Input
          id={schema.id}
          value={value}
          placeholder={schema.placeholder}
          onChange={(event) => onChange(event.target.value)}
          className="w-56"
        />
      }
    />
  )
}
