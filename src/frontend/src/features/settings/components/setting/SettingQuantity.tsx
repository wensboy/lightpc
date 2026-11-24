import { useEffect, useMemo, useState } from 'react'
import { ButtonGroup } from '@/components/ui/button-group'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { SettingField } from '@/features/settings/components/setting/SettingField'
import type { SettingField as SettingFieldModel, SettingValue } from '@/features/settings/schema'
import { useText } from '@/locales'

type SettingQuantityProps = {
  field: SettingFieldModel
  onChange: (value: SettingValue) => void
}

// fallbackNumber is what a field accepts when the schema declares no grammar of its own: a plain decimal.
const fallbackNumber = '^\\d+(\\.\\d+)?$'

// typing matches a value as it is being typed: digits, at most one decimal point, nothing else. The unit lives in
// the picker beside the input, so a pasted "14 px" is refused here rather than reaching commit as a number and
// producing "14 px px".
const typing = /^\d*\.?\d*$/

// splitQuantity separates the number from its unit. The separator is optional and may be whitespace, so the
// same code reads "8 px", "8px" and "8".
function splitQuantity(text: string): { number: string; unit: string } {
  const trimmed = text.trim()
  const index = trimmed.search(/[^\d.\-+]\s*$|[a-z%µ]/i)
  if (index <= 0) {
    return { number: trimmed, unit: '' }
  }
  return { number: trimmed.slice(0, index).trim(), unit: trimmed.slice(index).trim() }
}

// SettingQuantity edits a value that carries a unit: the number stays an input, the unit becomes a picker fed
// by the set the schema names.
//
// The number is checked against the grammar the schema declares before anything is committed. A rejected draft
// leaves the stored value and the interface untouched and shows the reason underneath, because a colour or a
// radius that silently changed to something invalid is far harder to trace than a field that refused the edit.
export function SettingQuantity({ field, onChange }: SettingQuantityProps) {
  const text = useText()
  const units = useMemo(() => field.units ?? [], [field.units])
  const grammar = useMemo(() => new RegExp(field.number ?? fallbackNumber), [field.number])
  const initial = useMemo(() => splitQuantity(String(field.value)), [field.value])

  const [number, setNumber] = useState(initial.number)
  const [unit, setUnit] = useState(initial.unit || units[0] || '')
  const [rejected, setRejected] = useState(false)

  // A value arriving from outside, such as a reload of the user layer, wins over an untouched draft.
  useEffect(() => {
    const next = splitQuantity(String(field.value))
    setNumber(next.number)
    setUnit(next.unit || units[0] || '')
    setRejected(false)
  }, [field.value, units])

  // commit writes the canonical form: the number, one space, the unit. The document is authored by hand too, so
  // the stored text has to be predictable rather than whatever the two controls happened to hold.
  const commit = (nextNumber: string, nextUnit: string) => {
    const value = nextNumber.trim()
    if (!grammar.test(value)) {
      setRejected(true)
      return
    }

    setRejected(false)
    const next = nextUnit ? `${value} ${nextUnit}` : value
    if (next !== String(field.value)) {
      onChange(next)
    }
  }

  return (
    <SettingField
      field={field}
      control={
        <div className="flex flex-col gap-1">
          <ButtonGroup className="w-56">
            <Input
              id={field.path}
              value={number}
              inputMode="decimal"
              aria-invalid={rejected}
              onChange={(event) => {
                const next = event.target.value
                if (!typing.test(next)) {
                  setRejected(true)
                  return
                }
                setNumber(next)
                setRejected(false)
              }}
              onBlur={() => commit(number, unit)}
              className="min-w-0 flex-1 rounded-r-none"
            />
            {units.length > 0 ? (
              <Select
                value={unit}
                onValueChange={(next) => {
                  const picked = String(next)
                  setUnit(picked)
                  commit(number, picked)
                }}
                items={units.map((entry) => ({ value: entry, label: entry }))}
              >
                <SelectTrigger className="w-16 shrink-0 rounded-l-none border-l-0" aria-label={field.path}>
                  <SelectValue placeholder={units[0]} />
                </SelectTrigger>
                <SelectContent alignItemWithTrigger={false}>
                  {units.map((entry) => (
                    <SelectItem key={entry} value={entry}>
                      {entry}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            ) : null}
          </ButtonGroup>
          {rejected ? (
            <p className="text-xs text-destructive">{text.settings.invalidQuantity}</p>
          ) : null}
        </div>
      }
    />
  )
}
