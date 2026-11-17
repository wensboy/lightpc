// The session module is the single place that knows where the token lives. It is intentionally tiny so
// the real AuthService can replace it once the bridge is wired up.
const storageKey = 'lightpc.auth.token'

export function getToken(): string | null {
  return window.localStorage.getItem(storageKey)
}

export function setToken(token: string): void {
  window.localStorage.setItem(storageKey, token)
}

export function clearToken(): void {
  window.localStorage.removeItem(storageKey)
}
