import { GalleryVerticalEndIcon } from 'lucide-react'
import { useMemo, useState } from 'react'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
} from '@/components/ui/sidebar'
import { useLogout } from '@/features/auth/useLogout'
import { NavContextMenu, type NavToggleGroup } from '@/features/dashboard/components/NavContextMenu'
import { NavMain } from '@/features/dashboard/components/NavMain'
import { NavSecondary } from '@/features/dashboard/components/NavSecondary'
import { NavUser } from '@/features/dashboard/components/NavUser'
import { WorkspaceSwitcher } from '@/features/dashboard/components/WorkspaceSwitcher'
import { buildPrimaryNav, buildSecondaryNav } from '@/features/dashboard/config/nav'
import { useNotificationPrefs } from '@/features/notifications/store/useNotificationPrefs'
import { useNotificationStore } from '@/features/notifications/store/useNotificationStore'
import { useText } from '@/locales'
import { useDialogStore } from '@/stores/dialog'

// AppSidebar assembles the shell: team switcher on top, the primary and secondary navigation in the
// middle, and the user menu at the bottom. All labels come from the text dictionary.
//
// The body answers a right click with a checkbox per entry. Hiding only removes the entry, so the remaining
// ones keep their order and the two blocks stay where they are.
export function AppSidebar() {
  const text = useText()
  const openSettings = useDialogStore((state) => state.openSettings)
  const openHelp = useDialogStore((state) => state.openHelp)
  const openCommand = useDialogStore((state) => state.openCommand)
  const openConsole = useDialogStore((state) => state.openConsole)
  const logout = useLogout()
  const openAccount = useDialogStore((state) => state.openAccount)
  const openNotifications = useNotificationStore((state) => state.openSheet)
  const hasUnreadNotifications = useNotificationStore((state) => state.unread)
  // The mode lives in the settings document now, so the sidebar reads it from the same place the toast does.
  const { doNotDisturb } = useNotificationPrefs()

  const [hidden, setHidden] = useState<ReadonlySet<string>>(() => new Set())

  const workspaces = [
    {
      name: text.dashboard.workspace.primary,
      logo: <GalleryVerticalEndIcon className="size-4" />,
      plan: text.dashboard.workspace.primaryPlan,
    },
    {
      name: text.dashboard.workspace.secondary,
      logo: <GalleryVerticalEndIcon className="size-4" />,
      plan: text.dashboard.workspace.secondaryPlan,
    },
  ]

  const user = {
    name: text.dashboard.user.name,
    email: text.dashboard.user.email,
    avatar: '/avatar-moty.png',
  }

  const primary = useMemo(() => buildPrimaryNav(text), [text])
  const secondary = useMemo(() => buildSecondaryNav(text), [text])

  // The menu lists every entry of both blocks, including the hidden ones, so a hidden entry can come back.
  const menuGroups = useMemo<NavToggleGroup[]>(
    () => [
      { items: primary.map((item) => ({ key: item.key, title: item.title, icon: item.icon })) },
      { items: secondary.map((item) => ({ key: item.key, title: item.title, icon: item.icon })) },
    ],
    [primary, secondary],
  )

  const visiblePrimary = primary.filter((item) => !hidden.has(item.key))
  const visibleSecondary = secondary.filter((item) => !hidden.has(item.key))

  const toggle = (key: string) => {
    setHidden((previous) => {
      const next = new Set(previous)
      if (next.has(key)) {
        next.delete(key)
      } else {
        next.add(key)
      }
      return next
    })
  }

  return (
    <Sidebar collapsible="icon">
      <SidebarHeader>
        <WorkspaceSwitcher
          workspaces={workspaces}
          labels={{
            workspaces: text.dashboard.workspace.workspaces,
            addWorkspace: text.dashboard.workspace.addWorkspace,
          }}
        />
      </SidebarHeader>
      <NavContextMenu groups={menuGroups} hidden={hidden} onToggle={toggle}>
        <SidebarContent className="flex-1">
          <NavMain items={visiblePrimary} />
          <NavSecondary
            items={visibleSecondary}
            onSelect={(key) => {
              switch (key) {
                case 'console':
                  openConsole()
                  break
                case 'settings':
                  openSettings()
                  break
                case 'help':
                  openHelp()
                  break
                case 'command':
                  openCommand()
                  break
              }
            }}
            className="mt-auto"
          />
        </SidebarContent>
      </NavContextMenu>
      <SidebarFooter>
        <NavUser
          user={user}
          labels={{
            account: text.dashboard.user.account,
            notifications: text.dashboard.user.notifications,
            logout: text.dashboard.user.logout,
          }}
          notificationBadge={
            hasUnreadNotifications ? 'unread' : doNotDisturb ? 'muted' : 'default'
          }
          onAccount={openAccount}
          onNotifications={openNotifications}
          onLogout={() => void logout()}
        />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
