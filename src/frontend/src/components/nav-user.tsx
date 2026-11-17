import {
  Avatar,
  AvatarFallback,
  AvatarImage,
} from "@/components/ui/avatar"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@/components/ui/sidebar"
import { EllipsisVerticalIcon, CircleUserRoundIcon, CreditCardIcon, BellIcon, BellDotIcon, BellOffIcon, LogOutIcon } from "lucide-react"

export type NavUserLabels = {
  account: string
  billing: string
  notifications: string
  logout: string
}

// NotificationBadge picks the notification glyph: unread when something has not been seen, muted while
// do not disturb is on, and the plain bell otherwise.
export type NotificationBadge = 'default' | 'unread' | 'muted'

const notificationIcons = {
  default: BellIcon,
  unread: BellDotIcon,
  muted: BellOffIcon,
}

export function NavUser({
  user,
  labels,
  notificationBadge = 'default',
  onAccount,
  onNotifications,
}: {
  user: {
    name: string
    email: string
    avatar: string
  }
  labels: NavUserLabels
  notificationBadge?: NotificationBadge
  onAccount?: () => void
  onNotifications?: () => void
}) {
  const { isMobile } = useSidebar()
  const NotificationIcon = notificationIcons[notificationBadge]

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger
            render={
              <SidebarMenuButton size="lg" className="aria-expanded:bg-muted" />
            }
          >
            <Avatar className="size-8 rounded-lg grayscale">
              <AvatarImage src={user.avatar} alt={user.name} />
              <AvatarFallback className="rounded-lg">CN</AvatarFallback>
            </Avatar>
            <div className="grid flex-1 text-left text-sm leading-tight group-data-[collapsible=icon]:hidden">
              <span className="truncate font-medium">{user.name}</span>
              <span className="truncate text-xs text-foreground/70">
                {user.email}
              </span>
            </div>
            <EllipsisVerticalIcon className="ml-auto size-4 group-data-[collapsible=icon]:hidden" />
          </DropdownMenuTrigger>
          <DropdownMenuContent
            className="min-w-56"
            side={isMobile ? "bottom" : "right"}
            align="end"
            sideOffset={4}
          >
            <DropdownMenuGroup>
              <DropdownMenuLabel className="p-0 font-normal">
                <div className="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
                  <Avatar className="size-8">
                    <AvatarImage src={user.avatar} alt={user.name} />
                    <AvatarFallback className="rounded-lg">CN</AvatarFallback>
                  </Avatar>
                  <div className="grid flex-1 text-left text-sm leading-tight">
                    <span className="truncate font-medium">{user.name}</span>
                    <span className="truncate text-xs text-muted-foreground">
                      {user.email}
                    </span>
                  </div>
                </div>
              </DropdownMenuLabel>
            </DropdownMenuGroup>
            <DropdownMenuSeparator />
            <DropdownMenuGroup>
              <DropdownMenuItem onClick={onAccount}>
                <CircleUserRoundIcon
                />
                {labels.account}
              </DropdownMenuItem>
              <DropdownMenuItem>
                <CreditCardIcon
                />
                {labels.billing}
              </DropdownMenuItem>
              <DropdownMenuItem onClick={onNotifications}>
                <NotificationIcon />
                {labels.notifications}
              </DropdownMenuItem>
            </DropdownMenuGroup>
            <DropdownMenuSeparator />
            <DropdownMenuItem>
              <LogOutIcon
              />
              {labels.logout}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  )
}
