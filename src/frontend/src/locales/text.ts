// resolveText walks a dotted dictionary key such as "settings.items.editorFontSize.label" and returns
// the string it points at. A missing path falls back to the key itself so gaps are visible in the UI
// instead of rendering nothing.
export function resolveText(dictionary: unknown, key: string): string {
  const value = key.split('.').reduce<unknown>((current, part) => {
    if (current && typeof current === 'object' && part in current) {
      return (current as Record<string, unknown>)[part]
    }
    return undefined
  }, dictionary)

  return typeof value === 'string' ? value : key
}
