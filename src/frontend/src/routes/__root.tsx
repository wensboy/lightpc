import { createRootRoute, Outlet, redirect } from '@tanstack/react-router'
import { Toaster } from '@/components/ui/sonner'
import { useNotificationPrefs } from '@/features/notifications/store/useNotificationPrefs'
import { readAuth } from '@/features/auth/session'

// The paths that stay reachable without a session.
const publicPaths = new Set(['/login', '/signup'])

// The root route renders the outlet: authenticated pages bring the sidebar shell through the "_app" layout
// while authentication pages stay standalone.
//
// The session check sits here rather than only in the layout below. It is a synchronous read, so a reload with
// no stored pair resolves straight to the form instead of walking into an asynchronous guard that could leave
// nothing to render. A stored pair is only trusted locally: the layout verifies it against the backend once.
export const Route = createRootRoute({
  beforeLoad: ({ location }) => {
    if (!readAuth() && !publicPaths.has(location.pathname)) {
      throw redirect({ to: '/login' })
    }
  },
  component: () => <RootLayout />,
  errorComponent: ({ error }) => (
    <div className="flex min-h-screen flex-col items-center justify-center gap-3 p-6 text-center">
      <p className="text-sm font-medium">Something went wrong</p>
      <p className="max-w-md text-xs break-words text-muted-foreground">
        {error instanceof Error ? error.message : String(error)}
      </p>
      <button
        type="button"
        className="rounded-md border border-border px-3 py-1.5 text-xs"
        onClick={() => window.location.reload()}
      >
        Reload
      </button>
    </div>
  ),
  notFoundComponent: () => (
    <div className="flex min-h-screen flex-col items-center justify-center gap-3 p-6 text-center">
      <p className="text-sm font-medium">Page not found</p>
      <a href="/" className="text-xs underline underline-offset-4">
        Go home
      </a>
    </div>
  ),
})

// RootLayout renders the outlet and the toast container.
//
// The container lives here rather than in the shell so a notice raised on any screen still has somewhere to appear. Its
// position and stack size come from the settings document: saving a new one publishes a change event, the hook
// re-reads, and the next toast lands in the new corner without a reload.
//
// visibleToasts is sonner's own stacking limit - the number of cards drawn behind the front one. Setting it here rather
// than dismissing from the handler keeps the two mechanisms from disagreeing: the library collapses the extras into the
// stack edge, while a manual dismiss would discard them and leave the count one short.
function RootLayout() {
  const prefs = useNotificationPrefs()

  return (
    <>
      <Outlet />
      <Toaster position={prefs.position} visibleToasts={prefs.fold} />
    </>
  )
}
