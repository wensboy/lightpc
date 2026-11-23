import { useNavigate } from '@tanstack/react-router'
import { useCallback } from 'react'
import { clearAuth, readAuth } from '@/features/auth/session'
import { sessionService } from '@/services/session'

// useLogout ends the session: the service call releases any server side state, the stored token is
// dropped and the router lands on the login page. The guard on the app layout would redirect anyway,
// but navigating keeps the transition immediate.
export function useLogout(): () => Promise<void> {
  const navigate = useNavigate()

  return useCallback(async () => {
    const auth = readAuth()
    if (auth) {
      await sessionService.logout(auth)
    }
    clearAuth()
    await navigate({ to: '/login' })
  }, [navigate])
}
