import { createRootRoute, Outlet } from '@tanstack/react-router'

// The root route only renders the outlet: authenticated pages bring the sidebar shell through the
// "_app" layout while authentication pages stay standalone.
export const Route = createRootRoute({
  component: () => <Outlet />,
})
