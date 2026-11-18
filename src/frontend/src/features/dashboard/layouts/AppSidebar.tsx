import { GalleryVerticalEndIcon } from 'lucide-react'
import { NavMain } from '@/components/nav-main'
import { NavSecondary } from '@/components/nav-secondary'
import { NavUser } from '@/components/nav-user'
import { WorkspaceSwitcher } from '@/components/workspace-switcher'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
} from '@/components/ui/sidebar'
import { useLogout } from '@/features/auth/useLogout'
import { buildPrimaryNav, buildSecondaryNav } from '@/features/dashboard/config/nav'
import { useNotificationStore } from '@/features/notifications/store/useNotificationStore'
import { useText } from '@/locales'
import { useDialogStore } from '@/stores/dialog'

// AppSidebar assembles the shell: team switcher on top, the primary and secondary navigation in the
// middle, and the user menu at the bottom. All labels come from the text dictionary.
export function AppSidebar() {
  const text = useText()
  const openSettings = useDialogStore((state) => state.openSettings)
  const openHelp = useDialogStore((state) => state.openHelp)
  const openCommand = useDialogStore((state) => state.openCommand)
  const logout = useLogout()
  const openAccount = useDialogStore((state) => state.openAccount)
  const openNotifications = useNotificationStore((state) => state.openSheet)
  const hasUnreadNotifications = useNotificationStore((state) => state.unread)
  const doNotDisturb = useNotificationStore((state) => state.doNotDisturb)

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
      <SidebarContent>
        <NavMain items={buildPrimaryNav(text)} />
        <NavSecondary
          items={buildSecondaryNav(text)}
          onSelect={(key) => {
            switch (key) {
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
