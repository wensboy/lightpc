import { BellIcon, BellOffIcon, CircleAlertIcon, CircleXIcon, LayersIcon, ListXIcon, TriangleAlertIcon } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Empty, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
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
import { noticeLevelTones } from '@/features/notifications/notice-level'
import { useNotificationPrefs } from '@/features/notifications/store/useNotificationPrefs'
import { useNotificationStore } from '@/features/notifications/store/useNotificationStore'
import { settingService } from '@/services/setting'
import { useText } from '@/locales'

// NotificationsSheet is the notification centre: a right side sheet without the default close button,
// a fixed header holding the title and its two actions, and a scrollable body below it.
export function NotificationsSheet() {
  const text = useText()
  const open = useNotificationStore((state) => state.open)
  const closeSheet = useNotificationStore((state) => state.closeSheet)
  const prefs = useNotificationPrefs()
  const clearNotifications = useNotificationStore((state) => state.clear)
  const notifications = useNotificationStore((state) => state.notifications)
  const summary = useNotificationStore((state) => state.summary)

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
              onClick={() => void clearNotifications()}
            >
              <ListXIcon />
            </Button>
            <Toggle
              size="sm"
              pressed={prefs.doNotDisturb}
              onPressedChange={(next) => {
                // The switch writes the setting rather than a local flag: the document is what the toast reads, so
                // storing it here would give the two views separate ideas of the mode.
                void settingService.save('general.notification.doNotDisturb', next, 'service')
              }}
              aria-label={text.notifications.doNotDisturb}
            >
              {prefs.doNotDisturb ? <BellOffIcon /> : <BellIcon />}
            </Toggle>
          </div>
        </SheetHeader>
        <SheetDescription className="sr-only">{text.notifications.description}</SheetDescription>
        <div className="flex flex-1 flex-col gap-1 overflow-y-auto p-2">
          {notifications.length === 0 ? (
            <Empty className="border-0 p-4">
              <EmptyHeader>
                <EmptyTitle className="text-sm font-normal text-muted-foreground">
                  {text.notifications.empty}
                </EmptyTitle>
              </EmptyHeader>
            </Empty>
          ) : (
            notifications.map((notification) => (
              <NotificationItem key={notification.id} notification={notification} />
            ))
          )}
        </div>

        {/* The footer describes the list above it, so its numbers come from the same read that produced the rows. Four
            cells share one row: a total, then the three severities worth a badge. */}
        <div className="flex items-center gap-2 border-t bg-card px-3 py-2">
          <FooterCell
            icon={<LayersIcon />}
            label={text.notifications.total}
            count={summary.total}
          />
          <FooterCell
            icon={<CircleXIcon className={noticeLevelTones.fail} />}
            label={text.notifications.fail}
            count={summary.fail}
          />
          <FooterCell
            icon={<CircleAlertIcon className={noticeLevelTones.error} />}
            label={text.notifications.error}
            count={summary.error}
          />
          <FooterCell
            icon={<TriangleAlertIcon className={noticeLevelTones.warn} />}
            label={text.notifications.warn}
            count={summary.warn}
          />
        </div>
      </SheetContent>
    </Sheet>
  )
}

// FooterCell is one counter in the sheet footer: a badge carrying an icon and a number, sharing the row with the others.
//
// The label is not drawn - four badges in one row cannot afford words - but it is announced and shown on hover, so the
// icon is never the only thing saying which count this is.
function FooterCell({ icon, label, count }: { icon: React.ReactNode; label: string; count: number }) {
  return (
    <div className="flex min-w-0 flex-1 items-center justify-center" title={label}>
      <Badge variant="ghost" aria-label={`${label}: ${count}`} className="gap-1">
        {icon}
        <span className="tabular-nums">{count}</span>
      </Badge>
    </div>
  )
}
