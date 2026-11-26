import type { ProfileValues } from '@/features/account/schema'
import { settingService, type SettingDocument } from '@/services/setting'
import { accountAvatarMock, accountProfileMock } from './account.mock'

// mocksEnabled is resolved once at startup; the switch never leaks into business components.
const mocksEnabled = import.meta.env.VITE_ENABLE_MOCK === 'true'

// There is no auth mock: the session service talks to the database the store already opened, so a mock would
// only hide a broken migration.

// loadSettings always asks the backend: the schema and the default document are embedded resources the
// file store serves, so there is nothing to mock here.
export function loadSettings(): Promise<SettingDocument> {
  return settingService.load()
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
