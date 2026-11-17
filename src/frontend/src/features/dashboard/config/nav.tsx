import {
  ChartBarIcon,
  CircleHelpIcon,
  CommandIcon,
  FolderIcon,
  LayoutDashboardIcon,
  ListIcon,
  Settings2Icon,
  UsersIcon,
} from 'lucide-react'
import type { ReactNode } from 'react'
import type { Dictionary } from '@/locales'

export type NavLink = {
  title: string
  to: '/' | '/lifecycle' | '/analytics' | '/projects' | '/team'
  icon: ReactNode
}

export type NavAction = {
  key: 'command' | 'help' | 'settings'
  title: string
  icon: ReactNode
}

// buildPrimaryNav keeps destinations, icons and labels together so the sidebar entries and the page
// header titles never drift apart.
export function buildPrimaryNav(text: Dictionary): NavLink[] {
  return [
    { title: text.dashboard.nav.dashboard, to: '/', icon: <LayoutDashboardIcon /> },
    { title: text.dashboard.nav.lifecycle, to: '/lifecycle', icon: <ListIcon /> },
    { title: text.dashboard.nav.analytics, to: '/analytics', icon: <ChartBarIcon /> },
    { title: text.dashboard.nav.projects, to: '/projects', icon: <FolderIcon /> },
    { title: text.dashboard.nav.team, to: '/team', icon: <UsersIcon /> },
  ]
}

// buildSecondaryNav returns the items stacked from the bottom of the sidebar upwards. Each entry
// carries a key so the shell opens the matching overlay instead of navigating.
export function buildSecondaryNav(text: Dictionary): NavAction[] {
  return [
    { key: 'command', title: text.dashboard.nav.command, icon: <CommandIcon /> },
    { key: 'help', title: text.dashboard.nav.help, icon: <CircleHelpIcon /> },
    { key: 'settings', title: text.dashboard.nav.settings, icon: <Settings2Icon /> },
  ]
}
