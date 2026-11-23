import { createRootRoute, Outlet, redirect } from '@tanstack/react-router'
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
  component: () => <Outlet />,
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
