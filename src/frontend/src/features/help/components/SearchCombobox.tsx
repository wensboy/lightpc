import {
  Combobox,
  ComboboxContent,
  ComboboxEmpty,
  ComboboxInput,
  ComboboxItem,
  ComboboxList,
} from '@/components/ui/combobox'
import { useText } from '@/locales'

type SearchComboboxProps = {
  value: string
  documents: string[]
  onSelect: (document: string) => void
}

// SearchCombobox is the document picker of the help dialog, built on the shadcn combobox so filtering,
// keyboard navigation and the clear affordance come from the primitive. It owns no state: the selected
// document comes from the caller, keeping the trigger, the panel body and the tree highlight identical.
export function SearchCombobox({ value, documents, onSelect }: SearchComboboxProps) {
  const text = useText()

  return (
    <Combobox
      items={documents}
      value={value}
      onValueChange={(next) => {
        if (typeof next === 'string' && next !== '') {
          onSelect(next)
        }
      }}
    >
      <ComboboxInput placeholder={text.help.search.placeholder} className="w-full" />
      <ComboboxContent>
        <ComboboxEmpty>{text.help.search.empty}</ComboboxEmpty>
        <ComboboxList>
          {(item: string) => (
            <ComboboxItem key={item} value={item}>
              {item}
            </ComboboxItem>
          )}
        </ComboboxList>
      </ComboboxContent>
    </Combobox>
  )
}
