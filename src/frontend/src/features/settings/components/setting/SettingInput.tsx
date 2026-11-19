import { useEffect, useState } from 'react'
import { Input } from '@/components/ui/input'
import { SettingField } from '@/features/settings/components/setting/SettingField'
import type { SettingField as SettingFieldModel, SettingValue } from '@/features/settings/schema'

type SettingInputProps = {
  field: SettingFieldModel
  onChange: (value: SettingValue) => void
}

// SettingInput keeps a local draft and commits it on blur. Typing therefore stays purely local: the
// interface is not repainted per keystroke, and a half typed value never reaches the other setters.
export function SettingInput({ field, onChange }: SettingInputProps) {
  const [draft, setDraft] = useState(() => String(field.value))

  // An external change, such as a reload of the user layer, wins over an untouched draft.
  useEffect(() => {
    setDraft(String(field.value))
  }, [field.value])

  return (
    <SettingField
      field={field}
      control={
        <Input
          id={field.path}
          value={draft}
          onChange={(event) => setDraft(event.target.value)}
          onBlur={() => {
            if (draft !== String(field.value)) {
              onChange(draft)
            }
          }}
          className="w-56"
        />
      }
    />
  )
}
