import {
  ChevronDownIcon,
  ChevronUpIcon,
  CircleAlertIcon,
  InfoIcon,
  TriangleAlertIcon,
  XIcon,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import type {
  NotificationEntry,
  NotificationType,
} from '@/features/notifications/store/useNotificationStore'
import { useNotificationStore } from '@/features/notifications/store/useNotificationStore'
import { cn } from '@/lib/utils'
import { useText } from '@/locales'

const typeIcons = {
  info: InfoIcon,
  warning: TriangleAlertIcon,
  error: CircleAlertIcon,
}

const typeTones: Record<NotificationType, string> = {
  info: 'text-muted-foreground',
  warning: 'text-foreground',
  error: 'text-destructive',
}

// NotificationItem renders one notification as a card. Collapsed it keeps the title and the description
// on a single line each; expanded it shows them in full. The source line stays at the bottom of the card
// and is never part of the toggle.
export function NotificationItem({ notification }: { notification: NotificationEntry }) {
  const text = useText()
  const toggleExpanded = useNotificationStore((state) => state.toggleExpanded)
  const dismissNotification = useNotificationStore((state) => state.dismissNotification)
  const Icon = typeIcons[notification.type]
  const expanded = notification.isExpanded

  return (
    <div className="flex flex-col gap-1 rounded-lg border border-transparent p-4 transition-colors hover:border-border hover:bg-accent/60">
      <div className="flex items-start gap-2">
        <Icon className={cn('mt-0.5 size-4 shrink-0', typeTones[notification.type])} />
        <p className={cn('min-w-0 flex-1 text-sm font-medium', !expanded && 'truncate')}>
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
            onClick={() => dismissNotification(notification.id)}
          >
            <XIcon />
          </Button>
        </div>
      </div>
      <p className={cn('pl-6 text-xs text-muted-foreground', !expanded && 'truncate')}>
        {notification.description}
      </p>
      <p className="pl-6 text-xs text-muted-foreground">
        {text.notifications.source}: {notification.source}
      </p>
    </div>
  )
}
