import { z } from 'zod'
import { currentText } from '@/locales'

const messages = currentText().account.errors

// minPasswordLength is the smallest password the backend accepts.
const minPasswordLength = 8

// profileSchema validates the three identity fields of the profile section.
export const profileSchema = z.object({
  uid: z.string().min(1, messages.uidRequired),
  uname: z.string().min(1, messages.nameRequired),
  uaccount: z.string().min(1, messages.accountRequired),
})

export type ProfileValues = z.infer<typeof profileSchema>

// securitySchema requires the new password to be long enough and repeated exactly once. The mismatch is
// reported on the confirmation field so the message sits next to the input that is wrong.
export const securitySchema = z
  .object({
    oldPassword: z.string().min(1, messages.oldPasswordRequired),
    newPassword: z.string().min(minPasswordLength, messages.shortPassword),
    checkPassword: z.string().min(minPasswordLength, messages.shortPassword),
  })
  .refine((values) => values.newPassword === values.checkPassword, {
    message: messages.passwordMismatch,
    path: ['checkPassword'],
  })

export type SecurityValues = z.infer<typeof securitySchema>
