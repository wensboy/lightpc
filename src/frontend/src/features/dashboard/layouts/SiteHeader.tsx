import { useText } from '@/locales'
import { useLocation } from '@tanstack/react-router'
import { Separator } from '@/components/ui/separator'
import { SidebarTrigger } from '@/components/ui/sidebar'

// routeTitles maps a pathname to its dictionary key; the header never carries a literal title.
const routeTitles = {
  '/': 'dashboard',
  '/lifecycle': 'lifecycle',
  '/analytics': 'analytics',
  '/projects': 'projects',
  '/team': 'team',
} as const

type RouteTitle = keyof typeof routeTitles

// SiteHeader holds the sidebar toggle and the current page title; it stays mounted across routes.
// Layout mirrors the sidebar-07 header so the separator keeps the intended height and alignment.
export function SiteHeader() {
  const text = useText()
  const { pathname } = useLocation()
  const key = routeTitles[pathname as RouteTitle]
  const title = key ? text.dashboard.nav[key] : pathname

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
