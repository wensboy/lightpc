import { createFileRoute, Outlet, redirect, useNavigate } from '@tanstack/react-router'
import { useCallback } from 'react'
import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar'
import { AccountDialog } from '@/features/account/components/AccountDialog'
import { clearAuth, isVerified, markVerified, readAuth } from '@/features/auth/session'
import { CommandPalette } from '@/features/command/components/CommandPalette'
import { ConsoleSheet } from '@/features/console/components/ConsoleSheet'
import { AppSidebar } from '@/features/dashboard/layouts/AppSidebar'
import { SiteHeader } from '@/features/dashboard/layouts/SiteHeader'
import { HelpDialog } from '@/features/help/components/HelpDialog'
import { NotificationsSheet } from '@/features/notifications/components/NotificationsSheet'
import { SettingsDialog } from '@/features/settings/components/SettingsDialog'
import { NoticeToaster } from '@/features/notifications/components/NoticeToaster'
import { useWailsEvent } from '@/hooks/use-wails-event'
import { sessionService } from '@/services/session'

export const Route = createFileRoute('/_app')({
  // The root already refused an anonymous visit. What is left is deciding whether the stored pair is still
  // live: the backend answers from its timer first and from the history second, which is what makes a restart
  // inside the window sign the user back in and an expired one land on the login page.
  beforeLoad: async () => {
    const auth = readAuth()
    if (!auth) {
      throw redirect({ to: '/login' })
    }
    if (isVerified()) {
      return { auth }
    }

    let authenticated = false
    try {
      const status = await sessionService.current(auth)
      authenticated = status.authenticated
    } catch {
      // An unreachable backend must not keep the interface signed in.
      authenticated = false
    }
    if (!authenticated) {
      clearAuth()
      throw redirect({ to: '/login' })
    }

    markVerified()
    return { auth }
  },
  component: AppLayout,
})

// AppLayout owns the authenticated shell: the sidebar and header stay mounted while pages swap in the
// inset.
function AppLayout() {
  const navigate = useNavigate()

  // The backend publishes this when a live pair expires while the interface is open, so a short window closes
  // the session without waiting for the next request.
  useWailsEvent(
    'session::expired',
    useCallback(() => {
      clearAuth()
      void navigate({ to: '/login', replace: true })
    }, [navigate]),
  )

  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset>
        <SiteHeader />
        <Outlet />
      </SidebarInset>
      <SettingsDialog />
      <HelpDialog />
      <ConsoleSheet />
      <CommandPalette />
      <NotificationsSheet />
      <NoticeToaster />
      <AccountDialog />
    </SidebarProvider>
  )
}
