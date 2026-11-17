import { currentText } from '@/locales'
import type { AuthService, AuthSession, LoginInput } from '@/services/auth'

// Development credentials live only here so no business component ever sees them.
const mockAccount = {
  email: 'user@test.com',
  password: 'user123456',
}

// artificialDelay mimics the IPC round trip, so loading and race conditions surface in development.
const artificialDelay = 300

function delay(): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, artificialDelay))
}

export const authMock: AuthService = {
  async login(input: LoginInput): Promise<AuthSession> {
    await delay()
    if (input.email !== mockAccount.email || input.password !== mockAccount.password) {
      throw new Error(currentText().auth.errors.invalidCredentials)
    }
    return { token: `mock.${Date.now()}`, email: input.email }
  },

  async logout(): Promise<void> {
    await delay()
  },
}
