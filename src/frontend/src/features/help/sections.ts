import type { Dictionary } from '@/locales'

export type HelpSection = {
  title: string
  items: string[]
}

// toSections flattens the help dictionary into the shape the document tree and its combobox consume.
export function toSections(text: Dictionary): HelpSection[] {
  return Object.values(text.help.sections).map((section) => ({
    title: section.title,
    items: Object.values(section.items),
  }))
}
