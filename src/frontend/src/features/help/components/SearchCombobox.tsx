import {
  Combobox,
  ComboboxCollection,
  ComboboxContent,
  ComboboxEmpty,
  ComboboxGroup,
  ComboboxInput,
  ComboboxItem,
  ComboboxLabel,
  ComboboxList,
  ComboboxSeparator,
} from '@/components/ui/combobox'
import { useText } from '@/locales'

// SearchSection is one group of the picker. Items are document paths, because that is what the reader
// loads; the visible text comes from labels.
export type SearchSection = {
  title: string
  items: string[]
}

type SearchComboboxProps = {
  /** value is the display text of the current document; empty means nothing is chosen. */
  value: string
  /** sections mirror the first level of the table of contents. */
  sections: readonly SearchSection[]
  /** labels maps a document path to the title shown in the list. */
  labels: Readonly<Record<string, string>>
  /** onSelect receives the path of the picked document. */
  onSelect: (path: string) => void
}

// SearchCombobox is the document picker of the help dialog. It follows the grouped combobox pattern: the
// list is rendered from a function so the primitive owns filtering and keyboard navigation. The field only
// reports a selection; the tree keeps showing every section regardless of what was typed.
export function SearchCombobox({ value, sections, labels, onSelect }: SearchComboboxProps) {
  const text = useText()

  return (
    <Combobox
      items={sections}
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
          {(group: SearchSection, index: number) => (
            <ComboboxGroup key={group.title} items={group.items}>
              <ComboboxLabel className="px-1.5 py-1 text-xs text-muted-foreground">
                {group.title}
              </ComboboxLabel>
              <ComboboxCollection>
                {(item: string) => (
                  <ComboboxItem key={item} value={item}>
                    <span className="truncate">{labels[item] ?? item}</span>
                  </ComboboxItem>
                )}
              </ComboboxCollection>
              {index < sections.length - 1 ? <ComboboxSeparator /> : null}
            </ComboboxGroup>
          )}
        </ComboboxList>
      </ComboboxContent>
    </Combobox>
  )
}
