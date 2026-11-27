import type { SettingValue } from '@/features/settings/models/schema'
import { setColorMode, type ColorMode } from '@/lib/color-mode'
import { localeOf, setLocale } from '@/locales'

// SettingEffect applies the visible consequence of a value. Registering every key that changes the
// interface in one place keeps the handling uniform, instead of scattering branches across components.
type SettingEffect = (value: SettingValue) => void

// SettingThemeDto is one palette the backend already merged and keyed as custom properties.
type SettingThemeDto = {
  name: string
  light: Record<string, string>
  dark: Record<string, string>
}

// themePalettes caches the palettes the backend resolved, so picking a name can paint it without the
// effect having to reach back into the document.
let themePalettes: Record<string, SettingThemeDto> = {}

// colorModes maps the stored value onto the colour mode helper, which owns the class and the system
// listener. Anything unexpected is treated as the system default rather than as a hard failure.
const colorModes: Record<string, ColorMode> = { light: 'light', dark: 'dark', system: 'system' }

// applyLocale switches the dictionary. Every component that reads text subscribes to the store, so the
// whole interface repaints without remounting.
function applyLocale(value: SettingValue): void {
  setLocale(localeOf(String(value)))
}

// settingEffects is the single registry of keys with a frontend effect. A key without an entry only
// affects the backend, so it is written through without touching the view.
const settingEffects: Record<string, SettingEffect> = {
  'appearence.mode': applyColorMode,
  'general.language': applyLocale,
  'appearence.radius': applyRadius,
  'appearence.font.size': applyFontSize,
  'appearence.font.default': applyFontFamily,
  'appearence.theme.default': applyTheme,
}

// isDark reports whether the dark class is currently applied.
function isDark(): boolean {
  return document.documentElement.classList.contains('dark')
}

// applyPalette writes the custom properties the backend resolved.
function applyPalette(palette: Record<string, string> | undefined): void {
  if (!palette) {
    return
  }
  const style = document.documentElement.style
  for (const [property, value] of Object.entries(palette)) {
    style.setProperty(property, value)
  }
}

// applyTheme paints the palette of the active colour mode.
function applyTheme(value: SettingValue): void {
  const theme = themePalettes[String(value)]
  if (!theme) {
    return
  }
  applyPalette(isDark() ? theme.dark : theme.light)
}

// repaintTheme replays the active theme after the colour mode changed, so the palette follows the mode.
function repaintTheme(): void {
  for (const [path, effect] of Object.entries(settingEffects)) {
    if (path === 'appearence.theme.default') {
      effect(currentThemeName())
    }
  }
}

// currentThemeName reads the selected theme name off the document.
function currentThemeName(): string {
  const name = document.documentElement.dataset.theme
  return name ?? ''
}

// applyColorMode repaints the interface as soon as the value changes.
function applyColorMode(value: SettingValue): void {
  setColorMode(colorModes[String(value)] ?? 'system')
  repaintTheme()
}

// compactLength drops the separator the document uses between a number and its unit, because CSS rejects
// "12 px" while accepting "12px". An intermediate keystroke produces an invalid length that the browser
// ignores, leaving the previous value in place.
function compactLength(value: SettingValue): string {
  return String(value).replace(/\s+/g, '')
}

// applyRadius writes the corner radius onto the theme variable. Every radius utility is derived from it
// through calc, so one property repaints the whole surface.
function applyRadius(value: SettingValue): void {
  document.documentElement.style.setProperty('--radius', compactLength(value))
}

// applyFontSize drives the text scale variable that every text utility derives from. Spacing and control
// sizes stay on their own rem values, so the copy grows without moving the layout.
function applyFontSize(value: SettingValue): void {
  document.documentElement.style.setProperty('--font-size-body', compactLength(value))
}

// applyFontFamily writes the chosen family into the stack the body reads. Quotes are stripped so a name
// coming from a configuration file cannot break out of the declaration.
function applyFontFamily(value: SettingValue): void {
  const family = String(value).replace(/['"]/g, '').trim()
  document.documentElement.style.setProperty('--font-sans', `'${family}', sans-serif`)
}

// readPath reads a dotted path out of a nested document.
function readPath(document: Record<string, unknown>, path: string): unknown {
  let cursor: unknown = document
  for (const segment of path.split('.')) {
    if (typeof cursor !== 'object' || cursor === null) {
      return undefined
    }
    cursor = (cursor as Record<string, unknown>)[segment]
  }
  return cursor
}

// setThemes records the palettes the backend resolved. The conversion from the document shape already
// happened on the backend, so this is only a lookup table for the active theme.
export function setThemes(themes: readonly SettingThemeDto[]): void {
  const table: Record<string, SettingThemeDto> = {}
  for (const theme of themes) {
    table[theme.name] = theme
  }
  themePalettes = table
}

// applySettingEffect runs the effect of one key when the key has one.
export function applySettingEffect(path: string, value: SettingValue): void {
  if (path === 'appearence.theme.default') {
    document.documentElement.dataset.theme = String(value)
  }
  settingEffects[path]?.(value)
}

// applyAllSettingEffects replays every effect for a freshly loaded document, so the view reflects the
// stored values right after a load or a reload from the user layer.
export function applyAllSettingEffects(values: Record<string, unknown>, themes: readonly SettingThemeDto[]): void {
  setThemes(themes)
  for (const [path, effect] of Object.entries(settingEffects)) {
    const value = readPath(values, path)
    if (path === 'appearence.theme.default' && typeof value === 'string') {
      document.documentElement.dataset.theme = value
    }
    if (typeof value === 'string' || typeof value === 'number' || typeof value === 'boolean') {
      effect(value)
    }
  }
}
