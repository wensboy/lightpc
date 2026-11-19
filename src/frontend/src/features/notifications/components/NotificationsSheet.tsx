import { BellIcon, BellOffIcon, ListXIcon } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { Toggle } from '@/components/ui/toggle'
import { NotificationItem } from '@/features/notifications/components/NotificationItem'
import { useNotificationStore } from '@/features/notifications/store/useNotificationStore'
import { useText } from '@/locales'

// NotificationsSheet is the notification centre: a right side sheet without the default close button,
// a fixed header holding the title and its two actions, and a scrollable body below it.
export function NotificationsSheet() {
  const text = useText()
  const open = useNotificationStore((state) => state.open)
  const closeSheet = useNotificationStore((state) => state.closeSheet)
  const doNotDisturb = useNotificationStore((state) => state.doNotDisturb)
  const toggleDoNotDisturb = useNotificationStore((state) => state.toggleDoNotDisturb)
  const clearNotifications = useNotificationStore((state) => state.clearNotifications)
  const notifications = useNotificationStore((state) => state.notifications)

  return (
    <Sheet
      open={open}
      onOpenChange={(next) => {
        if (!next) {
          closeSheet()
        }
      }}
    >
      <SheetContent side="right" showCloseButton={false} className="flex w-96 flex-col gap-0 p-0">
        <SheetHeader className="flex flex-row items-center justify-between gap-2 border-b bg-card px-4 py-3">
          <SheetTitle className="text-sm font-medium">{text.notifications.title}</SheetTitle>
          <div className="flex items-center gap-1">
            <Button
              variant="ghost"
              size="icon-sm"
              aria-label={text.notifications.clear}
              onClick={clearNotifications}
            >
              <ListXIcon />
            </Button>
            <Toggle
              size="sm"
              pressed={doNotDisturb}
              onPressedChange={toggleDoNotDisturb}
              aria-label={text.notifications.doNotDisturb}
            >
              {doNotDisturb ? <BellOffIcon /> : <BellIcon />}
            </Toggle>
          </div>
        </SheetHeader>
        <SheetDescription className="sr-only">{text.notifications.description}</SheetDescription>
        <div className="flex flex-1 flex-col gap-1 overflow-y-auto p-2">
          {notifications.length === 0 ? (
            <p className="p-4 text-center text-xs text-muted-foreground">
              {text.notifications.empty}
            </p>
          ) : (
            notifications.map((notification) => (
              <NotificationItem key={notification.id} notification={notification} />
            ))
          )}
        </div>
      </SheetContent>
    </Sheet>
  )
}
