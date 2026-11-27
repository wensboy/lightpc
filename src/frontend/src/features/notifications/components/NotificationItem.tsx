import { ChevronDownIcon, ChevronUpIcon, XIcon } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { noticeLevelIcons, noticeLevelTones } from '@/features/notifications/notice-level'
import type { NotificationEntry } from '@/features/notifications/store/useNotificationStore'
import { useNotificationStore } from '@/features/notifications/store/useNotificationStore'
import { cn } from '@/lib/utils'
import { useText } from '@/locales'

// NotificationItem renders one notification as a card. Collapsed it keeps the title and the description
// on a single line each; expanded it shows them in full. The source line stays at the bottom of the card
// and is never part of the toggle.
export function NotificationItem({ notification }: { notification: NotificationEntry }) {
  const text = useText()
  const toggleExpanded = useNotificationStore((state) => state.toggleExpanded)
  const dismissNotification = useNotificationStore((state) => state.dismiss)
  const Icon = noticeLevelIcons[notification.level]
  const expanded = notification.isExpanded

  return (
    <div className="flex flex-col gap-1 rounded-lg border border-transparent p-4 transition-colors hover:border-border hover:bg-accent/60">
      <div className="flex items-start gap-2">
        <Icon className={cn('mt-0.5 size-4 shrink-0', noticeLevelTones[notification.level])} />
        <p className={cn('min-w-0 flex-1 text-sm font-medium break-words', !expanded && 'truncate')}>
          {notification.title}
        </p>
        <div className="-mt-1 flex shrink-0 items-center">
          <Button
            variant="ghost"
            size="icon-sm"
            aria-expanded={expanded}
            aria-label={expanded ? text.notifications.collapse : text.notifications.expand}
            onClick={() => toggleExpanded(notification.id)}
          >
            {expanded ? <ChevronUpIcon /> : <ChevronDownIcon />}
          </Button>
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label={text.notifications.dismiss}
            onClick={() => void dismissNotification(notification.id)}
          >
            <XIcon />
          </Button>
        </div>
      </div>
      <p className={cn('pl-6 text-xs break-words text-muted-foreground', !expanded && 'truncate')}>
        {notification.description}
      </p>
      {/* The source is an identifier, so one line with an ellipsis reads better than a wrapped one. */}
      <p className="truncate pl-6 text-xs text-muted-foreground" title={notification.source}>
        {text.notifications.source}: {notification.source}
      </p>
    </div>
  )
}
