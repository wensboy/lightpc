import { z } from 'zod'
import { currentText } from '@/locales'

// loginSchema is the single source of truth for the login form. It is built per call so the messages follow
// the active locale instead of freezing the one that happened to load first.
//
// The sign in itself goes through sessionService: this file only describes the shape the form validates.
export function loginSchema() {
  const errors = currentText().auth.errors
  return z.object({
    email: z.email(errors.invalidAccount),
    password: z.string().min(8, errors.shortPassword),
  })
}

export type LoginSchema = ReturnType<typeof loginSchema>

export type LoginInput = z.infer<LoginSchema>
