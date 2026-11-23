import type { AuthMeta } from '@/services/session'

// The session module is the single place that knows what survives a restart.
//
// Only the pair is stored: uid and account identify which history rows to look at, and the backend decides
// whether the window is still open. Neither value is a credential, so a copied storage snapshot grants nothing
// the account itself would not.
const storageKey = 'lightpc.session'

/** readAuth returns the stored pair, or null when the visitor signed out on purpose. */
export function readAuth(): AuthMeta | null {
  try {
    const raw = window.localStorage.getItem(storageKey)
    if (!raw) {
      return null
    }
    const parsed = JSON.parse(raw) as Partial<AuthMeta>
    if (typeof parsed.uid !== 'string' || typeof parsed.account !== 'string') {
      return null
    }
    if (parsed.uid === '' || parsed.account === '') {
      return null
    }
    return { uid: parsed.uid, account: parsed.account }
  } catch {
    return null
  }
}

/** writeAuth stores the pair after a successful sign in or restore. */
export function writeAuth(auth: AuthMeta): void {
  window.localStorage.setItem(storageKey, JSON.stringify({ uid: auth.uid, account: auth.account }))
}

/** clearAuth removes the pair, which is what a manual sign out means. */
export function clearAuth(): void {
  window.localStorage.removeItem(storageKey)
  verified = false
}

// verified remembers that the stored pair was checked against the backend in this process. The check is an IPC
// round trip, so it runs once rather than on every navigation; clearing the pair invalidates it.
let verified = false

/** isVerified reports whether the current pair already passed the backend check. */
export function isVerified(): boolean {
  return verified
}

/** markVerified records that the pair passed the backend check. */
export function markVerified(): void {
  verified = true
}
