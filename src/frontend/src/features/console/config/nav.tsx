import { GaugeIcon, ScrollTextIcon } from 'lucide-react'
import type { ReactNode } from 'react'
import type { Dictionary } from '@/locales'

export type ConsoleSectionKey = 'log' | 'performance'

export type ConsoleNavItem = {
  key: ConsoleSectionKey
  title: string
  icon: ReactNode
}

// buildConsoleNav keeps the destinations, icons and labels of the console in one place. The order here is the
// order the icon rail renders, so the two never drift apart.
export function buildConsoleNav(text: Dictionary): ConsoleNavItem[] {
  return [
    { key: 'log', title: text.console.sections.log, icon: <ScrollTextIcon /> },
    { key: 'performance', title: text.console.sections.performance, icon: <GaugeIcon /> },
  ]
}
