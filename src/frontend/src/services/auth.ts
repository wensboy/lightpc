import { z } from 'zod'
import { currentText } from '@/locales'

const messages = currentText().auth.errors

// loginSchema is the single source of truth for the login form and the payload check.
export const loginSchema = z.object({
  email: z.email(messages.invalidEmail),
  password: z.string().min(8, messages.shortPassword),
})

export type LoginInput = z.infer<typeof loginSchema>

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
  login: () => Promise.reject(new Error(messages.notConnected)),
  logout: () => Promise.resolve(),
}
