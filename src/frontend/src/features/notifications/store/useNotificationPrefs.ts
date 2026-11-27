import { useCallback, useEffect, useState } from 'react'
import { settingService } from '@/services/setting'
import { useWailsEvent } from '@/hooks/use-wails-event'

/** ToastPosition is where a notice popup appears. */
export type ToastPosition = 'top-right' | 'top-left' | 'bottom-right' | 'bottom-left'

/** NotificationPrefs is the display half of the notification settings. */
export type NotificationPrefs = {
  /** doNotDisturb silences the toast only; the backend records every notice regardless. */
  doNotDisturb: boolean
  /** history caps how many rows the centre shows. The backend enforces it and reports the left edge. */
  history: number
  /** fold caps how many notices one toast stacks before the rest collapse into a count. */
  fold: number
  /** durationMs is how long one toast stays on screen. */
  durationMs: number
  /** position is where a toast appears. */
  position: ToastPosition
}

// The defaults mirror setting.default.json. They only apply while the document is loading or when a value is
// unusable; they never override what the document says.
const fallback: NotificationPrefs = {
  doNotDisturb: false,
  history: 32,
  fold: 3,
  durationMs: 5000,
  position: 'top-right',
}

// readDuration reads a value written as "<number> <unit>", falling back when it is absent or malformed.
function readDuration(value: unknown, fallbackMs: number): number {
  const text = String(value ?? '').trim()
  const match = /^(\d+(?:\.\d+)?)\s*(ms|s|m|h)?$/i.exec(text)
  if (!match) {
    return fallbackMs
  }
  const amount = Number(match[1])
  switch ((match[2] ?? 's').toLowerCase()) {
    case 'ms':
      return amount
    case 'm':
      return amount * 60_000
    case 'h':
      return amount * 3_600_000
    default:
      return amount * 1000
  }
}

// readPrefs narrows the notification section of the settings document.
function readPrefs(values: Record<string, unknown> | undefined): NotificationPrefs {
  const notification = (values?.general as Record<string, unknown> | undefined)?.notification as
    | Record<string, unknown>
    | undefined
  if (!notification) {
    return fallback
  }

  const fold = Number(notification.fold)
  const history = Number(notification.history)
  const position = String(notification.position ?? '')

  return {
    doNotDisturb: notification.doNotDisturb === true,
    history: Number.isFinite(history) && history > 0 ? history : fallback.history,
    // fold never exceeds history: a stack larger than the list would offer to expand into rows that are not there.
    fold:
      Number.isFinite(fold) && fold > 0 ? Math.min(fold, Number.isFinite(history) && history > 0 ? history : fallback.history) : fallback.fold,
    durationMs: readDuration(notification.duration, fallback.durationMs),
    position: ['top-right', 'top-left', 'bottom-right', 'bottom-left'].includes(position)
      ? (position as ToastPosition)
      : fallback.position,
  }
}

// useNotificationPrefs reads the notification settings and keeps them current.
//
// Every field here only affects how a notice is shown, never whether it is recorded: the backend stores and caps on its
// own, so a change reached here takes effect on the next toast without a round trip. That is why a settings save is
// enough - the change event arrives, the document is re-read, and the next notice renders with the new values.
export function useNotificationPrefs(): NotificationPrefs {
  const [prefs, setPrefs] = useState<NotificationPrefs>(fallback)

  const reload = useCallback(() => {
    void settingService.load()
      .then((document) => setPrefs(readPrefs(document.values)))
      .catch(() => undefined)
  }, [])

  useEffect(() => {
    reload()
  }, [reload])

  // A save lands on the service bus, so one subscription covers a position change just as well as a fold change.
  useWailsEvent('setting::changed', reload)

  return prefs
}
