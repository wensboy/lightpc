import { CircleAlertIcon, CircleCheckIcon, InfoIcon, TriangleAlertIcon, XIcon } from 'lucide-react'
import { useCallback, useEffect } from 'react'
import { toast } from 'sonner'
import { levelTones, type levelIcons } from '@/features/notifications/components/NotificationItem'
import { useNotificationStore } from '@/features/notifications/store/useNotificationStore'
import { useNotificationPrefs } from '@/features/notifications/store/useNotificationPrefs'
import { useWailsEvent } from '@/hooks/use-wails-event'
import { useText } from '@/locales'
import type { NoticeItemDto, NoticeLevelDto } from '@/services/notice'

// The tone and icon tables live with the sheet item, but a toast needs its own copies: importing the component's maps
// would mean the toast pulls the whole card in, and the two render very differently.
const toneOf: Record<NoticeLevelDto, string> = levelTones

const iconOf: Record<NoticeLevelDto, typeof InfoIcon> = {
  info: InfoIcon,
  success: CircleCheckIcon,
  warning: TriangleAlertIcon,
  error: CircleAlertIcon,
} satisfies typeof levelIcons

// payloadOf narrows what the bus delivered. The event carries the engine's record, so an unexpected shape is dropped
// rather than rendered as an empty toast.
function payloadOf(value: unknown): NoticeItemDto | null {
  if (typeof value !== 'object' || value === null) {
    return null
  }
  const candidate = value as Partial<NoticeItemDto>
  if (typeof candidate.id !== 'number' || typeof candidate.title !== 'string') {
    return null
  }
  return {
    id: candidate.id,
    title: candidate.title,
    content: candidate.content ?? '',
    source: candidate.source ?? '',
    level: candidate.level ?? 'info',
    timestamp: candidate.timestamp ?? 0,
  }
}

// NoticeToaster subscribes to the notice event and renders each notice as a toast.
//
// It is the only place that decides whether a popup appears. The backend records every notice regardless of this
// choice: do not disturb silences the toast, while the entry still lands in the sheet and survives switching the mode
// off, because the ring never learned about it in the first place.
export function NoticeToaster() {
  const text = useText()
  const accept = useNotificationStore((state) => state.accept)
  // Everything below only decides how a notice is shown. Whether it is recorded was settled by the backend before the
  // event arrived, so a preference change never needs a second call.
  const prefs = useNotificationPrefs()

  const handle = useCallback(
    (raw: unknown) => {
      const notice = payloadOf(raw)
      if (notice === null) {
        return
      }

      // The sheet is updated either way: a silenced notice is still a notice, and opening the centre must show it.
      accept(notice)
      if (prefs.doNotDisturb) {
        return
      }

      const Icon = iconOf[notice.level]
      toast.custom(
        (id) => (
          <div className="flex w-full items-start gap-2 rounded-lg border bg-popover p-3 text-popover-foreground shadow-lg">
            <Icon className={`mt-0.5 size-4 shrink-0 ${toneOf[notice.level]}`} />
            <div className="flex min-w-0 flex-1 flex-col gap-0.5">
              <p className="truncate text-sm font-medium">{notice.title}</p>
              {notice.content ? (
                <p className="line-clamp-3 text-xs text-muted-foreground">{notice.content}</p>
              ) : null}
              {notice.source ? (
                <p className="text-xs text-muted-foreground/70">
                  {text.notifications.source}: {notice.source}
                </p>
              ) : null}
            </div>
            {/* Only the dismiss control: a notice is informational, so there is no action to take and no need for the
                confirm/cancel pair a dialog would draw. */}
            <button
              type="button"
              aria-label={text.notifications.dismiss}
              className="-mr-1 -mt-1 rounded p-1 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
              onClick={() => toast.dismiss(id)}
            >
              <XIcon className="size-3.5" />
            </button>
          </div>
        ),
        { duration: prefs.durationMs, closeButton: false },
      )
    },
    [accept, prefs.doNotDisturb, prefs.durationMs, text],
  )

  useWailsEvent('notice::send', handle)

  // One read on mount, so the sidebar badge and the sheet start from what the engine already holds rather than from
  // an empty list that only fills in when the sheet is opened.
  useEffect(() => {
    void useNotificationStore.getState().refresh()
  }, [])

  return null
}
