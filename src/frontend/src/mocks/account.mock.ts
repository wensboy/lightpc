import type { ProfileValues } from '@/features/account/schema'

// The mock profile mirrors the shape the real service will return.
export const accountProfileMock: ProfileValues = {
  uid: 'uid-8f2c1e',
  uname: 'moty',
  uaccount: 'moty@lightpc.local',
}

// accountAvatarMock reuses the bundled avatar so the placeholder renders offline.
export const accountAvatarMock = '/avatar-moty.png'
