import { SidebarMenu, SidebarMenuButton, SidebarMenuItem } from '@/components/ui/sidebar'
import type { ConsoleNavItem, ConsoleSectionKey } from '@/features/console/config/nav'

type ConsoleNavProps = {
  items: readonly ConsoleNavItem[]
  active: ConsoleSectionKey
  onSelect: (key: ConsoleSectionKey) => void
}

// ConsoleNav is the icon rail of the console. It never expands: the label lives in the tooltip, which is the
// same shape the main sidebar takes when it is collapsed to icons.
export function ConsoleNav({ items, active, onSelect }: ConsoleNavProps) {
  return (
    <SidebarMenu className="items-center gap-1">
      {items.map((item) => (
        <SidebarMenuItem key={item.key}>
          <SidebarMenuButton
            tooltip={item.title}
            isActive={active === item.key}
            onClick={() => onSelect(item.key)}
            className="justify-center"
          >
            {item.icon}
            <span className="sr-only">{item.title}</span>
          </SidebarMenuButton>
        </SidebarMenuItem>
      ))}
    </SidebarMenu>
  )
}
