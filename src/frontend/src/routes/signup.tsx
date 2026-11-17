import { createFileRoute, redirect } from '@tanstack/react-router'
import { getToken } from '@/features/auth/session'
import { SignupPage } from '@/features/auth/pages/SignupPage'

export const Route = createFileRoute('/signup')({
  // An existing session skips the registration form.
  beforeLoad: () => {
    if (getToken()) {
      throw redirect({ to: '/' })
    }
  },
  component: SignupPage,
})
