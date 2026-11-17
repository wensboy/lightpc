import { useLocation } from '@tanstack/react-router'
import { Separator } from '@/components/ui/separator'
import { SidebarTrigger } from '@/components/ui/sidebar'

const titles: Record<string, string> = {
  '/': 'Dashboard',
  '/lifecycle': 'Lifecycle',
  '/analytics': 'Analytics',
  '/projects': 'Projects',
  '/team': 'Team',
}

// SiteHeader holds the sidebar toggle and the current page title; it stays mounted across routes.
// Layout mirrors the sidebar-07 header so the separator keeps the intended height and alignment.
export function SiteHeader() {
  const { pathname } = useLocation()
  const title = titles[pathname] ?? pathname

  return (
    <header className="flex h-14 shrink-0 items-center gap-2 border-b transition-[width,height] ease-linear">
      <div className="flex w-full items-center gap-1 px-4 lg:gap-2 lg:px-6">
        <SidebarTrigger className="-ml-1" />
        <Separator orientation="vertical" className="mx-2 h-4 data-vertical:self-auto" />
        <h1 className="text-base font-medium">{title}</h1>
      </div>
    </header>
  )
}
