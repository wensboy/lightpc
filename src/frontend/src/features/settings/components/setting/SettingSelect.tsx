import {
  Combobox,
  ComboboxContent,
  ComboboxEmpty,
  ComboboxInput,
  ComboboxItem,
  ComboboxList,
} from '@/components/ui/combobox'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { SettingField } from '@/features/settings/components/setting/SettingField'
import type { SettingField as SettingFieldModel, SettingValue } from '@/features/settings/schema'
import { useText } from '@/locales'

// comboboxThreshold is where a list becomes too long to scan by scrolling. The system font list routinely
// holds hundreds of families, and those fields need filtering more than they need a dropdown.
const comboboxThreshold = 20

type SettingSelectProps = {
  field: SettingFieldModel
  onChange: (value: SettingValue) => void
}

// SettingSelect renders a field constrained by an option list. Short lists keep the dropdown; long lists
// switch to the combobox so the user can type instead of scrolling past hundreds of entries.
export function SettingSelect({ field, onChange }: SettingSelectProps) {
  const text = useText()
  const options = field.options ?? []

  if (options.length > comboboxThreshold) {
    return (
      <SettingField
        field={field}
        control={
          <Combobox
            items={options}
            value={String(field.value)}
            onValueChange={(next) => {
              if (typeof next === 'string' && next !== '') {
                onChange(next)
              }
            }}
          >
            <ComboboxInput id={field.path} placeholder={String(field.value)} className="w-72" />
            <ComboboxContent>
              <ComboboxEmpty>{text.settings.noResults}</ComboboxEmpty>
              <ComboboxList>
                {(item: string) => (
                  <ComboboxItem key={item} value={item}>
                    {item}
                  </ComboboxItem>
                )}
              </ComboboxList>
            </ComboboxContent>
          </Combobox>
        }
      />
    )
  }

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
          {/* Dropdown style: the popup aligns to the trigger edge instead of overlapping it. */}
          <SelectContent alignItemWithTrigger={false}>
            {options.map((option) => (
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
