import { Fragment, type ReactNode } from 'react'
import {
  ContextMenu,
  ContextMenuCheckboxItem,
  ContextMenuContent,
  ContextMenuSeparator,
  ContextMenuTrigger,
} from '@/components/ui/context-menu'

// NavToggleItem is one entry the menu can hide or show.
export type NavToggleItem = {
  key: string
  title: string
  icon?: ReactNode
}

// NavToggleGroup is a block of the sidebar. Groups carry no title: the separator between them is the whole
// grouping, which is what a reader needs to tell the top block from the bottom one.
export type NavToggleGroup = {
  items: NavToggleItem[]
}

type NavContextMenuProps = {
  groups: readonly NavToggleGroup[]
  /** hidden lists the keys the sidebar currently omits. */
  hidden: ReadonlySet<string>
  onToggle: (key: string) => void
  children: ReactNode
}

// NavContextMenu turns the sidebar body into a context menu of checkboxes, one per navigation entry.
//
// Toggling only removes an entry: the remaining entries keep their order, so the two blocks stay in the same
// place relative to each other no matter what was hidden.
export function NavContextMenu({ groups, hidden, onToggle, children }: NavContextMenuProps) {
  return (
    <ContextMenu>
      <ContextMenuTrigger render={<div className="flex min-h-0 flex-1 flex-col" />}>
        {children}
      </ContextMenuTrigger>
      <ContextMenuContent className="min-w-48">
        {groups.map((group, index) => (
          <Fragment key={index}>
            {index > 0 ? <ContextMenuSeparator /> : null}
            {group.items.map((item) => (
              <ContextMenuCheckboxItem
                key={item.key}
                checked={!hidden.has(item.key)}
                onCheckedChange={() => onToggle(item.key)}
              >
                {item.icon}
                <span>{item.title}</span>
              </ContextMenuCheckboxItem>
            ))}
          </Fragment>
        ))}
      </ContextMenuContent>
    </ContextMenu>
  )
}
