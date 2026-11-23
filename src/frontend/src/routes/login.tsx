import { createFileRoute, redirect } from '@tanstack/react-router'
import { isVerified, readAuth } from '@/features/auth/session'
import { LoginPage } from '@/features/auth/pages/LoginPage'

export const Route = createFileRoute('/login')({
  // An existing session skips the login form.
  beforeLoad: () => {
    if (readAuth() && isVerified()) {
      throw redirect({ to: '/' })
    }
  },
  component: LoginPage,
})
