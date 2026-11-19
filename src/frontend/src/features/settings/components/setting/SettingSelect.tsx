import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { SettingField } from '@/features/settings/components/setting/SettingField'
import type { SettingField as SettingFieldModel, SettingValue } from '@/features/settings/schema'

type SettingSelectProps = {
  field: SettingFieldModel
  onChange: (value: SettingValue) => void
}

// SettingSelect renders a field constrained by a schema enum. The options are the enum values, so the
// control always mirrors what the schema allows.
export function SettingSelect({ field, onChange }: SettingSelectProps) {
  return (
    <SettingField
      field={field}
      control={
        <Select
          value={String(field.value)}
          onValueChange={(next) => {
            if (typeof next === 'string') {
              onChange(next)
            }
          }}
        >
          <SelectTrigger id={field.path} className="w-72">
            <SelectValue />
          </SelectTrigger>
          <SelectContent alignItemWithTrigger={false}>
            {(field.options ?? []).map((option) => (
              <SelectItem key={option} value={option}>
                {option}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      }
    />
  )
}
