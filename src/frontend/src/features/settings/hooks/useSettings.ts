import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { applyAllSettingEffects, applySettingEffect } from '@/features/settings/libs/effects'
import { buildGroups, setAt, type SettingGroup, type SettingValue } from '@/features/settings/models/schema'
import { settingService, type SettingDocument } from '@/services/setting'

export type SettingsState = {
  groups: SettingGroup[]
  /**
   * loading is true between opening the dialog and the document arriving. The view draws a skeleton of the final
   * layout while it holds, so the panel does not flash an empty state it is about to replace.
   */
  loading: boolean
  error: string | null
  setValue: (path: string, value: SettingValue, kind?: string) => void
}

// useSettings loads the settings document whenever the dialog opens. The document has to be re-read
// because the user layer can change outside the UI: a manual file edit, or a flush from another window.
export function useSettings(active = true): SettingsState {
  const [document, setDocument] = useState<SettingDocument | null>(null)
  const [loading, setLoading] = useState(active)
  // hydrated records whether a document has ever arrived. Reopening the dialog re-reads it, but the previous content is
  // still on screen, so the skeleton only belongs to the first load.
  const hydrated = useRef(false)
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
    // Only the first load leaves the panel empty; a reload keeps what is drawn and swaps it when the answer arrives.
    if (!hydrated.current) {
      setLoading(true)
    }
    settingService.load()
      .then((loaded) => {
        if (cancelled) {
          return
        }
        edited.current = new Set()
        repaired.current = new Set()
        hydrated.current = true
        setDocument(loaded)
        setValues(loaded.values)
        setLoading(false)
        // The view replays every effect, so a stored theme applies before the first paint of the dialog.
        applyAllSettingEffects(loaded.values, loaded.themes)
      })
      .catch((cause: unknown) => {
        if (cancelled) {
          return
        }
        setError(cause instanceof Error ? cause.message : 'settings could not be loaded')
        setLoading(false)
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
        void settingService.save(field.path, field.value, field.kind)
      }
    }
  }, [groups])

  // Typing fires one change per keystroke, so writes are coalesced: the UI updates immediately while the
  // backend receives one save per quiet period instead of one per character.
  // Each pending write keeps the kind it was made with: the field is gone by the time the batch flushes, and the
  // backend needs to know which bus the change belongs on.
  const pending = useRef(new Map<string, { value: SettingValue; kind?: string }>())
  const timer = useRef<number | null>(null)

  const flushPending = useCallback(() => {
    if (timer.current !== null) {
      window.clearTimeout(timer.current)
      timer.current = null
    }
    for (const [path, entry] of pending.current) {
      void settingService.save(path, entry.value, entry.kind)
    }
    pending.current.clear()
  }, [])

  useEffect(() => () => flushPending(), [flushPending])

  const setValue = useCallback(
    (path: string, value: SettingValue, kind?: string) => {
      // Frontend effects run first: a themed switch must repaint immediately instead of waiting for the
      // value to reach the backend and come back. Keys without an effect skip straight to the save.
      applySettingEffect(path, value)
      edited.current.add(path)
      setValues((previous) => setAt(previous, path, value))
      pending.current.set(path, { value, kind })
      if (timer.current !== null) {
        window.clearTimeout(timer.current)
      }
      timer.current = window.setTimeout(flushPending, 500)
    },
    [flushPending],
  )

  return { groups, loading, error, setValue }
}
