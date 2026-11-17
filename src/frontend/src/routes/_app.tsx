import { createFileRoute, Outlet, redirect } from '@tanstack/react-router'
import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar'
import { getToken } from '@/features/auth/session'
import { AppSidebar } from '@/features/dashboard/layouts/AppSidebar'
import { SiteHeader } from '@/features/dashboard/layouts/SiteHeader'

export const Route = createFileRoute('/_app')({
  // Every page below this layout requires a session; without one the user lands on the login page.
  beforeLoad: () => {
    if (!getToken()) {
      throw redirect({ to: '/login' })
    }
  },
  component: AppLayout,
})

// AppLayout owns the authenticated shell: the sidebar and header stay mounted while pages swap in the
// inset.
function AppLayout() {
  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset>
        <SiteHeader />
        <Outlet />
      </SidebarInset>
    </SidebarProvider>
  )
}
