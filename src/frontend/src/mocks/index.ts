import type { ProfileValues } from '@/features/account/schema'
import type { NotificationEntry } from '@/features/notifications/store/useNotificationStore'
import type { SettingDocument } from '@/features/settings/schema'
import { authService, type AuthService } from '@/services/auth'
import { accountAvatarMock, accountProfileMock } from './account.mock'
import { authMock } from './auth.mock'
import { notificationsMock } from './notifications.mock'
import { settingsMock } from './settings.mock'

// mocksEnabled is resolved once at startup; the switch never leaks into business components.
const mocksEnabled = import.meta.env.VITE_ENABLE_MOCK === 'true'

export const auth: AuthService = mocksEnabled ? authMock : authService

// loadSettings returns the settings document. Only the mock source exists today; the real branch throws
// until the backend exposes the endpoint.
export function loadSettings(): SettingDocument {
  if (mocksEnabled) {
    return settingsMock
  }
  throw new Error('Settings are not connected to the backend yet.')
}

// loadNotifications returns the initial notification list. Without mocks the list starts empty, which is
// what a fresh backend connection would report.
export function loadNotifications(): NotificationEntry[] {
  return mocksEnabled ? notificationsMock : []
}

// loadAccountProfile returns the profile form defaults.
export function loadAccountProfile(): ProfileValues {
  if (mocksEnabled) {
    return accountProfileMock
  }
  return { uid: '', uname: '', uaccount: '' }
}

// loadAccountAvatar returns the avatar shown before the user picks a file.
export function loadAccountAvatar(): string {
  return mocksEnabled ? accountAvatarMock : ''
}
