import {
  ChartBarIcon,
  CircleHelpIcon,
  FolderIcon,
  LayoutDashboardIcon,
  ListIcon,
  SearchIcon,
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
  title: string
  url: string
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

// buildSecondaryNav returns the items stacked from the bottom of the sidebar upwards.
export function buildSecondaryNav(text: Dictionary): NavAction[] {
  return [
    { title: text.dashboard.nav.search, url: '#', icon: <SearchIcon /> },
    { title: text.dashboard.nav.help, url: '#', icon: <CircleHelpIcon /> },
    { title: text.dashboard.nav.settings, url: '#', icon: <Settings2Icon /> },
  ]
}
