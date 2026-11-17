import { authService, type AuthService } from '@/services/auth'
import { authMock } from './auth.mock'

// mocksEnabled is resolved once at startup; the switch never leaks into business components.
const mocksEnabled = import.meta.env.VITE_ENABLE_MOCK === 'true'

export const auth: AuthService = mocksEnabled ? authMock : authService
