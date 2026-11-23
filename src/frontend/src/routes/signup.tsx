import { createFileRoute, redirect } from '@tanstack/react-router'
import { isVerified, readAuth } from '@/features/auth/session'
import { SignupPage } from '@/features/auth/pages/SignupPage'

export const Route = createFileRoute('/signup')({
  // An existing session skips the registration form.
  beforeLoad: () => {
    if (readAuth() && isVerified()) {
      throw redirect({ to: '/' })
    }
  },
  component: SignupPage,
})
