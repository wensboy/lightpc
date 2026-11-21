import { z } from 'zod'
import { currentText } from '@/locales'

// loginSchema is the single source of truth for the login form and the payload check. It is built per call
// so the messages follow the active locale instead of freezing the one that happened to load first.
export function loginSchema() {
  const errors = currentText().auth.errors
  return z.object({
    email: z.email(errors.invalidEmail),
    password: z.string().min(8, errors.shortPassword),
  })
}

export type LoginSchema = ReturnType<typeof loginSchema>

export type LoginInput = z.infer<LoginSchema>

// AuthSession is what the caller keeps after a successful login.
export type AuthSession = {
  token: string
  email: string
}

// AuthService is the only contract business code depends on; the implementation is picked in mocks.
export interface AuthService {
  login(input: LoginInput): Promise<AuthSession>
  logout(): Promise<void>
}

// authService is the real implementation. It stays unimplemented until the Go backend exposes auth.
export const authService: AuthService = {
  login: () => Promise.reject(new Error(currentText().auth.errors.notConnected)),
  logout: () => Promise.resolve(),
}
