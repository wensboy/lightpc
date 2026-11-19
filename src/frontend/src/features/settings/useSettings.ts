import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { applyAllSettingEffects, applySettingEffect } from '@/features/settings/effects'
import { buildGroups, setAt, type SettingGroup, type SettingValue } from '@/features/settings/schema'
import { loadSettings } from '@/mocks'
import { settingService, type SettingDocument } from '@/services/setting'

export type SettingsState = {
  groups: SettingGroup[]
  error: string | null
  setValue: (path: string, value: SettingValue) => void
}

// useSettings loads the settings document whenever the dialog opens. The document has to be re-read
// because the user layer can change outside the UI: a manual file edit, or a flush from another window.
export function useSettings(active = true): SettingsState {
  const [document, setDocument] = useState<SettingDocument | null>(null)
  const [values, setValues] = useState<Record<string, unknown>>({})
  // Fields the user touched are excluded from the default fallback until the document is reloaded.
  const edited = useRef(new Set<string>())
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!active) {
      return
    }
    let cancelled = false
    setError(null)
    loadSettings()
      .then((loaded) => {
        if (cancelled) {
          return
        }
        edited.current = new Set()
        repaired.current = new Set()
        setDocument(loaded)
        setValues(loaded.values)
        // The view replays every effect, so a stored theme applies before the first paint of the dialog.
        applyAllSettingEffects(loaded.values, loaded.themes)
      })
      .catch((cause: unknown) => {
        if (cancelled) {
          return
        }
        setError(cause instanceof Error ? cause.message : 'settings could not be loaded')
      })
    return () => {
      cancelled = true
    }
  }, [active])

  const groups = useMemo(
    () =>
      document
        ? buildGroups(document.schema, values, document.defaults, edited.current, document.systemFonts)
        : [],
    [document, values],
  )

  // A redisplayed value that breaks the schema is pushed into the runtime layer once, so the next flush
  // rewrites the user file and the view stops disagreeing with it. The stored value itself is untouched
  // until the user saves, which keeps the mismatch short lived instead of hiding it.
  const repaired = useRef(new Set<string>())
  useEffect(() => {
    for (const group of groups) {
      for (const field of group.fields) {
        if (!field.invalid || repaired.current.has(field.path)) {
          continue
        }
        repaired.current.add(field.path)
        void settingService.save(field.path, field.value)
      }
    }
  }, [groups])

  // Typing fires one change per keystroke, so writes are coalesced: the UI updates immediately while the
  // backend receives one save per quiet period instead of one per character.
  const pending = useRef(new Map<string, SettingValue>())
  const timer = useRef<number | null>(null)

  const flushPending = useCallback(() => {
    if (timer.current !== null) {
      window.clearTimeout(timer.current)
      timer.current = null
    }
    for (const [path, value] of pending.current) {
      void settingService.save(path, value)
    }
    pending.current.clear()
  }, [])

  useEffect(() => () => flushPending(), [flushPending])

  const setValue = useCallback(
    (path: string, value: SettingValue) => {
      // Frontend effects run first: a themed switch must repaint immediately instead of waiting for the
      // value to reach the backend and come back. Keys without an effect skip straight to the save.
      applySettingEffect(path, value)
      edited.current.add(path)
      setValues((previous) => setAt(previous, path, value))
      pending.current.set(path, value)
      if (timer.current !== null) {
        window.clearTimeout(timer.current)
      }
      timer.current = window.setTimeout(flushPending, 500)
    },
    [flushPending],
  )

  return { groups, error, setValue }
}
