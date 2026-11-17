import { create } from 'zustand'
import { loadNotifications } from '@/mocks'

export type NotificationType = 'info' | 'warning' | 'error'

// NotificationEntry is one row rendered by the sheet. isExpanded lives on the entry so the list keeps a
// single source of truth for its own presentation state.
export type NotificationEntry = {
  id: string
  title: string
  description: string
  source: string
  type: NotificationType
  isExpanded: boolean
}

type NotificationStore = {
  open: boolean
  doNotDisturb: boolean
  unread: boolean
  notifications: NotificationEntry[]
  openSheet: () => void
  closeSheet: () => void
  toggleDoNotDisturb: () => void
  clearNotifications: () => void
  dismissNotification: (id: string) => void
  toggleExpanded: (id: string) => void
}

// useNotificationStore owns the sheet visibility, the do not disturb flag and the notification list.
const initialNotifications = loadNotifications()

export const useNotificationStore = create<NotificationStore>((set) => ({
  open: false,
  doNotDisturb: false,
  // A non empty inbox starts unread, so the sidebar can advertise it before the sheet is opened.
  unread: initialNotifications.length > 0,
  notifications: initialNotifications,
  openSheet: () => set({ open: true, unread: false }),
  closeSheet: () => set({ open: false }),
  toggleDoNotDisturb: () => set((state) => ({ doNotDisturb: !state.doNotDisturb })),
  clearNotifications: () => {
    console.info('[notifications] clear')
    set({ notifications: [] })
  },
  dismissNotification: (id) =>
    set((state) => ({
      notifications: state.notifications.filter((entry) => entry.id !== id),
    })),
  toggleExpanded: (id) =>
    set((state) => ({
      notifications: state.notifications.map((entry) =>
        entry.id === id ? { ...entry, isExpanded: !entry.isExpanded } : entry,
      ),
    })),
}))
