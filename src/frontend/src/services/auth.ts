import { z } from 'zod'

// loginSchema is the single source of truth for the login form and the payload check.
export const loginSchema = z.object({
  email: z.email('Enter a valid email address.'),
  password: z.string().min(8, 'Password must be at least 8 characters long.'),
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
  login: () => Promise.reject(new Error('Auth is not connected to the backend yet.')),
  logout: () => Promise.resolve(),
}
