import { auth } from './auth'
import { common } from './common'
import { dashboard } from './dashboard'

// dictionaries keeps a locale dimension from day one so a real i18n runtime can be plugged in later
// without touching a single call site. zh-CN reuses the English entries until it is translated.
export const dictionaries = {
  'en-US': { common, auth, dashboard },
  'zh-CN': { common, auth, dashboard },
}

export type Locale = keyof typeof dictionaries
export type Dictionary = (typeof dictionaries)['en-US']

const activeLocale: Locale = 'en-US'

// currentText is the locale independent accessor used outside components, such as schemas and mocks.
export function currentText(): Dictionary {
  return dictionaries[activeLocale]
}

// useText returns the dictionary for the active locale. Components read text through it instead of
// hard-coding strings, which keeps the keys typed and discoverable.
export function useText(): Dictionary {
  return currentText()
}
