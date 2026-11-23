import { z } from 'zod'
import { SessionService as SessionBinding } from '../../bindings/github.com/wensboy/lightpc/backend/service/session'

// statusSchema validates the signed in state the backend reports. Every call answers with it, so a caller can
// replace its local view from any response instead of asking twice.
const statusSchema = z.object({
  authenticated: z.boolean(),
  uid: z.string().default(''),
  account: z.string().default(''),
  uname: z.string().default(''),
  expiresAt: z.number().default(0),
})

/** SessionStatusDto is the signed in state of the current visitor. */
export type SessionStatusDto = z.infer<typeof statusSchema>

/** AuthMeta is the pair the front end keeps locally to identify the session across restarts. */
export type AuthMeta = {
  uid: string
  account: string
}

/** SessionRegisterInputDto is what a new account is created with. */
export type SessionRegisterInputDto = {
  uname: string
  account: string
  password: string
}

/** SessionLoginInputDto is what a sign in needs. */
export type SessionLoginInputDto = {
  account: string
  password: string
}

/** SessionService is the contract the auth feature depends on. */
export interface SessionService {
  /** register creates an account and signs it in. */
  register(input: SessionRegisterInputDto): Promise<SessionStatusDto>
  /** login verifies the credentials and signs in. */
  login(input: SessionLoginInputDto): Promise<SessionStatusDto>
  /** logout drops the live mark on the backend. */
  logout(auth: AuthMeta): Promise<SessionStatusDto>
  /** unregister removes the account of the given pair. */
  unregister(auth: AuthMeta): Promise<SessionStatusDto>
  /** current reports the state, restoring it from the history after a restart. */
  current(auth: AuthMeta): Promise<SessionStatusDto>
}

// parseStatus narrows the binding result, which is a plain object at the process boundary.
function parseStatus(value: unknown): SessionStatusDto {
  const parsed = statusSchema.safeParse(value)
  if (!parsed.success) {
    throw new Error(`session status rejected: ${parsed.error.message}`)
  }
  return parsed.data
}

// sessionService talks to the backend. Nothing is mocked here: the session lives in the database the store
// already opened, so a mock would only hide a broken migration.
export const sessionService: SessionService = {
  async register(input) {
    return parseStatus(await SessionBinding.Register(input))
  },

  async login(input) {
    return parseStatus(await SessionBinding.Login(input))
  },

  async logout(auth) {
    return parseStatus(await SessionBinding.Logout(auth))
  },

  async unregister(auth) {
    return parseStatus(await SessionBinding.Unregister(auth))
  },

  async current(auth) {
    return parseStatus(await SessionBinding.Current(auth))
  },
}
