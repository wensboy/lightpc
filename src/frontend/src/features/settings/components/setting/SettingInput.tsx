import { useEffect, useState } from 'react'
import { Input } from '@/components/ui/input'
import { SettingField } from '@/features/settings/layouts/SettingField'
import type { SettingField as SettingFieldModel, SettingValue } from '@/features/settings/models/schema'

type SettingInputProps = {
  field: SettingFieldModel
  onChange: (value: SettingValue) => void
}

// SettingInput keeps a local draft and commits it on blur. Typing therefore stays purely local: the
// interface is not repainted per keystroke, and a half typed value never reaches the other setters.
//
// The draft is text because that is what an input holds, but the stored value has to match the declared type: an
// integer field committed as "4" is a string to the schema, and the field would immediately read back as invalid.
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
            const next = commit(field, draft)
            if (next === undefined) {
              // An unusable draft leaves the stored value alone, so a half typed number never replaces a good one.
              setDraft(String(field.value))
              return
            }
            if (next !== field.value) {
              onChange(next)
            }
          }}
          className="w-56"
        />
      }
    />
  )
}

// commit turns the draft into the type the schema declares. It returns undefined when the text cannot be read as that
// type, which the caller reports by restoring the stored value.
function commit(field: SettingFieldModel, draft: string): SettingValue | undefined {
  const text = draft.trim()
  if (field.type === 'integer' || field.type === 'number') {
    if (text === '') {
      return undefined
    }
    const value = Number(text)
    if (!Number.isFinite(value)) {
      return undefined
    }
    // An integer field rejects a fraction rather than rounding it: the person typed what they typed.
    if (field.type === 'integer' && !Number.isInteger(value)) {
      return undefined
    }
    return value
  }
  return draft
}
