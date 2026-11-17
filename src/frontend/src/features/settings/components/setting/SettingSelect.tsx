import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { SettingField } from '@/features/settings/components/setting/SettingField'
import type { SettingSchema } from '@/features/settings/schema'
import { useText } from '@/locales'
import { resolveText } from '@/locales/text'

type SettingSelectProps = {
  schema: Extract<SettingSchema, { type: 'select' }>
  categoryLabel: string
  value: string
  onChange: (value: string) => void
}

// SettingSelect renders an enumerated setting. Option labels are resolved through the dictionary as
// well, so values stay stable identifiers while the text remains translatable.
export function SettingSelect({ schema, categoryLabel, value, onChange }: SettingSelectProps) {
  const text = useText()

  return (
    <SettingField
      id={schema.id}
      categoryLabel={categoryLabel}
      labelKey={schema.labelKey}
      descriptionKey={schema.descriptionKey}
      control={
        <Select
          value={value}
          onValueChange={(next) => {
            if (typeof next === 'string') {
              onChange(next)
            }
          }}
        >
          <SelectTrigger id={schema.id} className="w-72">
            <SelectValue />
          </SelectTrigger>
          {/* Dropdown style: the popup aligns to the trigger edge instead of overlapping it. */}
          <SelectContent alignItemWithTrigger={false}>
            {schema.options.map((option) => (
              <SelectItem key={option.value} value={option.value}>
                {resolveText(text, option.labelKey)}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      }
    />
  )
}
