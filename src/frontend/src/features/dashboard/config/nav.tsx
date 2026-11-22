import {
  ChartBarIcon,
  CircleHelpIcon,
  CommandIcon,
  SquareTerminalIcon,
  GaugeIcon,
  LayoutDashboardIcon,
  ScrollTextIcon,
  Settings2Icon,
} from 'lucide-react'
import type { ReactNode } from 'react'
import type { Dictionary } from '@/locales'

export type NavLink = {
  key: 'dashboard' | 'analytics' | 'log' | 'performance'
  title: string
  to: '/' | '/analytics' | '/log' | '/performance'
  icon: ReactNode
}

export type NavAction = {
  key: 'console' | 'command' | 'help' | 'settings'
  title: string
  icon: ReactNode
}

// buildPrimaryNav keeps destinations, icons and labels together so the sidebar entries and the page
// header titles never drift apart.
export function buildPrimaryNav(text: Dictionary): NavLink[] {
  return [
    { key: 'dashboard', title: text.dashboard.nav.dashboard, to: '/', icon: <LayoutDashboardIcon /> },
    { key: 'analytics', title: text.dashboard.nav.analytics, to: '/analytics', icon: <ChartBarIcon /> },
    { key: 'log', title: text.dashboard.nav.log, to: '/log', icon: <ScrollTextIcon /> },
    { key: 'performance', title: text.dashboard.nav.performance, to: '/performance', icon: <GaugeIcon /> },
  ]
}

// buildSecondaryNav returns the items stacked from the bottom of the sidebar upwards. Each entry
// carries a key so the shell opens the matching overlay instead of navigating.
export function buildSecondaryNav(text: Dictionary): NavAction[] {
  return [
    { key: 'console', title: text.dashboard.nav.console, icon: <SquareTerminalIcon /> },
    { key: 'command', title: text.dashboard.nav.command, icon: <CommandIcon /> },
    { key: 'help', title: text.dashboard.nav.help, icon: <CircleHelpIcon /> },
    { key: 'settings', title: text.dashboard.nav.settings, icon: <Settings2Icon /> },
  ]
}
