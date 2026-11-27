import { useCallback, useEffect, useRef, useState } from 'react'
import { clearAuth, readAuth, writeAuth } from '@/features/auth/libs/session'
import {
  sessionService,
  type AuthMeta,
  type SessionLoginInputDto,
  type SessionRegisterInputDto,
  type SessionStatusDto,
} from '@/services/session'

/** SessionState is what the auth screens read and call. */
export type SessionState = {
  /** status is the last state the backend reported. */
  status: SessionStatusDto
  /** ready turns true once the first restore attempt finished, so a screen can avoid flashing the form. */
  ready: boolean
  /** error carries a failed attempt, empty when the last one succeeded. */
  error: string
  login: (input: SessionLoginInputDto) => Promise<boolean>
  register: (input: SessionRegisterInputDto) => Promise<boolean>
  logout: () => Promise<void>
  unregister: () => Promise<void>
}

const signedOut: SessionStatusDto = {
  authenticated: false,
  uid: '',
  account: '',
  uname: '',
  expiresAt: 0,
}

// useSession owns the sign in state.
//
// On mount it asks the backend about the pair stored locally. The backend answers from its cache first and from
// the history second, so a restart inside the window signs the user back in without any renewal. When the pair
// is missing the user signed out on purpose, and the call is skipped.
export function useSession(): SessionState {
  const [status, setStatus] = useState<SessionStatusDto>(signedOut)
  const [ready, setReady] = useState(false)
  const [error, setError] = useState('')
  const started = useRef(false)

  useEffect(() => {
    if (started.current) {
      return
    }
    started.current = true

    const stored = readAuth()
    if (!stored) {
      setReady(true)
      return
    }

    sessionService
      .current(stored)
      .then((next) => {
        setStatus(next)
        if (!next.authenticated) {
          clearAuth()
        }
      })
      .catch(() => {
        setStatus(signedOut)
      })
      .finally(() => setReady(true))
  }, [])

  const apply = useCallback((next: SessionStatusDto) => {
    setStatus(next)
    if (next.authenticated) {
      writeAuth({ uid: next.uid, account: next.account })
      return
    }
    clearAuth()
  }, [])

  const login = useCallback(
    async (input: SessionLoginInputDto) => {
      setError('')
      try {
        apply(await sessionService.login(input))
        return true
      } catch (cause: unknown) {
        setError(cause instanceof Error ? cause.message : 'login failed')
        return false
      }
    },
    [apply],
  )

  const register = useCallback(
    async (input: SessionRegisterInputDto) => {
      setError('')
      try {
        apply(await sessionService.register(input))
        return true
      } catch (cause: unknown) {
        setError(cause instanceof Error ? cause.message : 'registration failed')
        return false
      }
    },
    [apply],
  )

  // The pair is read before clearing: the backend needs it to record the action.
  const pairOf = useCallback((): AuthMeta | null => {
    if (status.authenticated) {
      return { uid: status.uid, account: status.account }
    }
    return readAuth()
  }, [status])

  const logout = useCallback(async () => {
    const auth = pairOf()
    try {
      if (auth) {
        await sessionService.logout(auth)
      }
    } catch {
      // A backend that cannot be reached must not keep the interface signed in.
    }
    clearAuth()
    setStatus(signedOut)
  }, [pairOf])

  const unregister = useCallback(async () => {
    const auth = pairOf()
    if (!auth) {
      setError('not signed in')
      return
    }
    try {
      await sessionService.unregister(auth)
      clearAuth()
      setStatus(signedOut)
    } catch (cause: unknown) {
      setError(cause instanceof Error ? cause.message : 'unregistration failed')
    }
  }, [pairOf])

  return { status, ready, error, login, register, logout, unregister }
}
