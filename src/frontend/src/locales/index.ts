import { useSyncExternalStore } from 'react'
import { account as accountEn } from './i18n/en-US/account'
import { auth as authEn } from './i18n/en-US/auth'
import { common as commonEn } from './i18n/en-US/common'
import { dashboard as dashboardEn } from './i18n/en-US/dashboard'
import { help as helpEn } from './i18n/en-US/help'
import { notifications as notificationsEn } from './i18n/en-US/notifications'
import { settings as settingsEn } from './i18n/en-US/settings'
import { account as accountZh } from './i18n/zh-CN/account'
import { auth as authZh } from './i18n/zh-CN/auth'
import { common as commonZh } from './i18n/zh-CN/common'
import { consoleText as consoleEn } from './i18n/en-US/console'
import { consoleText as consoleZh } from './i18n/zh-CN/console'
import { dashboard as dashboardZh } from './i18n/zh-CN/dashboard'
import { help as helpZh } from './i18n/zh-CN/help'
import { notifications as notificationsZh } from './i18n/zh-CN/notifications'
import { settings as settingsZh } from './i18n/zh-CN/settings'

// dictionaries keeps one directory per locale. English is the reference: every other locale has the same
// shape, so a missing translation is a type error rather than a runtime gap.
export const dictionaries = {
  'en-US': {
    account: accountEn,
    auth: authEn,
    common: commonEn,
    console: consoleEn,
    dashboard: dashboardEn,
    help: helpEn,
    notifications: notificationsEn,
    settings: settingsEn,
  },
  'zh-CN': {
    account: accountZh,
    auth: authZh,
    common: commonZh,
    console: consoleZh,
    dashboard: dashboardZh,
    help: helpZh,
    notifications: notificationsZh,
    settings: settingsZh,
  },
}

export type Locale = keyof typeof dictionaries
export type Dictionary = (typeof dictionaries)['en-US']

// The active locale lives in a module level store: switching it has to repaint every component that reads
// text, and the setting arrives from outside the React tree.
let activeLocale: Locale = 'en-US'
const listeners = new Set<() => void>()

// localeNames maps the display name a setting stores onto the locale it selects. The picker shows names, so
// the stored value is a name rather than an id.
const localeNames: Record<string, Locale> = {
  English: 'en-US',
  '简体中文': 'zh-CN',
}

// localeOf narrows an arbitrary string, accepting either an id or a display name. An unknown value falls back
// to the reference locale so it cannot blank the interface.
export function localeOf(value: string): Locale {
  if (value in dictionaries) {
    return value as Locale
  }
  return localeNames[value] ?? 'en-US'
}

// setLocale switches the dictionary and notifies the subscribers. It is a no-op when the value repeats.
export function setLocale(next: Locale): void {
  if (next === activeLocale) {
    return
  }
  activeLocale = next
  for (const listener of listeners) {
    listener()
  }
}

// subscribe registers a listener, which is what useSyncExternalStore needs.
export function subscribe(listener: () => void): () => void {
  listeners.add(listener)
  return () => {
    listeners.delete(listener)
  }
}

// currentText is the locale independent accessor used outside components, such as schemas and mocks. It has
// to be called at the point of use: a module level constant would freeze the first locale.
export function currentText(): Dictionary {
  return dictionaries[activeLocale] as Dictionary
}

// useText returns the dictionary for the active locale and rerenders the caller when it changes.
export function useText(): Dictionary {
  const locale = useSyncExternalStore(
    subscribe,
    () => activeLocale,
    () => activeLocale,
  )
  return dictionaries[locale] as Dictionary
}
