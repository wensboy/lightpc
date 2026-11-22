import { ArchiveXIcon, FileTextIcon, InboxIcon, SendIcon, TrashIcon } from 'lucide-react'
import type { ReactNode } from 'react'
import type { Dictionary } from '@/locales'

export type ConsoleSectionKey = 'inbox' | 'drafts' | 'sent' | 'archive' | 'trash'

export type ConsoleNavItem = {
  key: ConsoleSectionKey
  title: string
  icon: ReactNode
}

// buildConsoleNav keeps the destinations, icons and labels of the console in one place. The order here is the
// order the icon rail renders, so the two never drift apart.
export function buildConsoleNav(text: Dictionary): ConsoleNavItem[] {
  return [
    { key: 'inbox', title: text.console.sections.inbox, icon: <InboxIcon /> },
    { key: 'drafts', title: text.console.sections.drafts, icon: <FileTextIcon /> },
    { key: 'sent', title: text.console.sections.sent, icon: <SendIcon /> },
    { key: 'archive', title: text.console.sections.archive, icon: <ArchiveXIcon /> },
    { key: 'trash', title: text.console.sections.trash, icon: <TrashIcon /> },
  ]
}
