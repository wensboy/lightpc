import { useNavigate } from '@tanstack/react-router'
import { useCallback } from 'react'
import { clearToken } from '@/features/auth/session'
import { auth } from '@/mocks'

// useLogout ends the session: the service call releases any server side state, the stored token is
// dropped and the router lands on the login page. The guard on the app layout would redirect anyway,
// but navigating keeps the transition immediate.
export function useLogout(): () => Promise<void> {
  const navigate = useNavigate()

  return useCallback(async () => {
    await auth.logout()
    clearToken()
    await navigate({ to: '/login' })
  }, [navigate])
}
