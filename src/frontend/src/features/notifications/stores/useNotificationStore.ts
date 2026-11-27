import { create } from 'zustand'
import {
  noticeService,
  type NoticeItemDto,
  type NoticeLevelDto,
  type NoticeSummaryDto,
} from '@/services/notice'

/** NotificationLevel is the severity a notice carries, mirroring the backend vocabulary. */
export type NotificationLevel = NoticeLevelDto

// NotificationEntry is one row rendered by the sheet. isExpanded lives on the entry so the list keeps a single source
// of truth for its own presentation state; everything else comes from the backend and is never edited here.
export type NotificationEntry = {
  id: number
  title: string
  description: string
  source: string
  level: NotificationLevel
  timestamp: number
  isExpanded: boolean
}

type NotificationStore = {
  open: boolean
  unread: boolean
  notifications: NotificationEntry[]
  /**
   * summary is what the backend counted while reading. It describes the rows the sheet holds, so a capped window shows
   * a capped count instead of a number the interface derived on its own.
   */
  summary: NoticeSummaryDto
  /**
   * limit is the size of the visible window, as the backend reports it. The ring keeps far more than this; the sheet
   * only ever shows the newest few, so a new notice pushes the oldest one out of view while it stays recorded.
   */
  limit: number
  openSheet: () => void
  closeSheet: () => void
  /** refresh re-reads the backend list, keeping the rows in the order the engine returns them. */
  refresh: () => Promise<void>
  /** clear removes every notice the sheet currently holds. */
  clear: () => Promise<void>
  /** dismiss removes one notice. */
  dismiss: (id: number) => Promise<void>
  toggleExpanded: (id: number) => void
  /** accept appends a notice that arrived over the bus, so the sheet reflects it before the next read. */
  accept: (item: NoticeItemDto) => void
}

// toEntry maps a backend notice onto the row shape. The engine already returns newest first, and the id grows with
// insertion order, so sorting by id descending keeps a late arrival at the top without consulting the timestamp.
function toEntry(item: NoticeItemDto): NotificationEntry {
  return {
    id: item.id,
    title: item.title,
    description: item.content,
    source: item.source,
    level: item.level,
    timestamp: item.timestamp,
    isExpanded: false,
  }
}

// byIdDescending keeps the newest notice at the top. Timestamps can collide within one millisecond, while the id is
// assigned under a lock and never repeats.
function byIdDescending(left: NotificationEntry, right: NotificationEntry): number {
  return right.id - left.id
}

// useNotificationStore owns the sheet visibility, the do not disturb flag and the list. The list is a mirror of the
// backend ring: every mutation is sent there first, then the store is updated from the answer, so a reload shows the
// same thing.
export const useNotificationStore = create<NotificationStore>((set, get) => ({
  open: false,
  unread: false,
  notifications: [],
  summary: { total: 0, fail: 0, error: 0, warn: 0 },
  limit: 0,

  openSheet: () => {
    set({ open: true, unread: false })
    void get().refresh()
  },
  closeSheet: () => set({ open: false }),


  refresh: async () => {
    try {
      const history = await noticeService.history()
      set({
        notifications: history.items.map(toEntry).sort(byIdDescending),
        summary: history.summary,
        limit: history.limit,
      })
    } catch {
      // A failed read leaves the list as it was: a transient backend error should not empty the sheet.
    }
  },

  clear: async () => {
    if (get().notifications.length === 0) {
      return
    }
    try {
      // One request, not one per row: clearing raises a floor on the backend, so the older notices outside the window
      // cannot come back on the next read either.
      await noticeService.clear()
    } finally {
      await get().refresh()
    }
  },

  dismiss: async (id) => {
    try {
      await noticeService.remove([id])
    } finally {
      await get().refresh()
    }
  },

  toggleExpanded: (id) =>
    set((state) => ({
      notifications: state.notifications.map((entry) =>
        entry.id === id ? { ...entry, isExpanded: !entry.isExpanded } : entry,
      ),
    })),

  accept: (item) => {
    const entry = toEntry(item)
    set((state) => {
      const next = state.notifications.filter((existing) => existing.id !== entry.id)
      next.push(entry)
      next.sort(byIdDescending)
      // The window slides: past the limit the oldest leaves the view, though the engine still holds it. Trimming here
      // rather than on the next read is what keeps a burst of notices from growing the list past the setting.
      return {
        notifications: state.limit > 0 ? next.slice(0, state.limit) : next,
        unread: state.open ? state.unread : true,
      }
    })

    // The counts are the backend's to keep. A read shortly after the burst replaces the appended row and refreshes them,
    // so several notices arriving together still cost one request.
    scheduleRefresh()
  },
}))

// refreshDelay is how long the store waits after a notice arrives before re-reading. A burst of notices arrives as
// several events in quick succession; one read afterwards settles both the list and the counts.
const refreshDelay = 400
let refreshTimer: number | undefined

// scheduleRefresh coalesces the reads that follow a burst into one.
function scheduleRefresh(): void {
  window.clearTimeout(refreshTimer)
  refreshTimer = window.setTimeout(() => {
    void useNotificationStore.getState().refresh()
  }, refreshDelay)
}
