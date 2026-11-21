import { Link, useLocation } from '@tanstack/react-router'
import type { ReactNode } from 'react'
import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from '@/components/ui/sidebar'

export type NavMainItem = {
  key: string
  title: string
  icon?: ReactNode
  to: '/' | '/analytics' | '/log' | '/performance'
}

// NavMain renders the primary navigation. Items navigate through the router so the sidebar keeps its
// state while pages swap, and the active route is highlighted.
export function NavMain({ items }: { items: NavMainItem[] }) {
  const { pathname } = useLocation()

  return (
    <SidebarGroup>
      <SidebarGroupContent className="flex flex-col gap-2">
        <SidebarMenu className="gap-1">
          {items.map((item) => (
            <SidebarMenuItem key={item.key}>
              <SidebarMenuButton
                tooltip={item.title}
                isActive={pathname === item.to}
                render={<Link to={item.to} />}
              >
                {item.icon}
                <span>{item.title}</span>
              </SidebarMenuButton>
            </SidebarMenuItem>
          ))}
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  )
}
