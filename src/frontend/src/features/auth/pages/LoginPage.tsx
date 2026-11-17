import { zodResolver } from '@hookform/resolvers/zod'
import { Link, useNavigate } from '@tanstack/react-router'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { Button } from '@/components/ui/button'
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldSeparator,
} from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { AuthLayout } from '@/features/auth/layouts/AuthLayout'
import { setToken } from '@/features/auth/session'
import { auth } from '@/mocks'
import { loginSchema, type LoginInput } from '@/services/auth'

// LoginPage validates with Zod through React Hook Form and hands the payload to the auth service.
export function LoginPage() {
  const navigate = useNavigate()
  const [message, setMessage] = useState('')
  const { register, handleSubmit, formState } = useForm<LoginInput>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: '', password: '' },
  })

  const onSubmit = handleSubmit(async (values) => {
    setMessage('')
    try {
      const session = await auth.login(values)
      setToken(session.token)
      await navigate({ to: '/' })
    } catch (error) {
      setMessage(error instanceof Error ? error.message : 'Login failed.')
    }
  })

  return (
    <AuthLayout>
      <form onSubmit={onSubmit} className="flex flex-col gap-6" noValidate>
        <FieldGroup>
          <div className="flex flex-col items-center gap-1 text-center">
            <h1 className="text-2xl font-bold">Login to your account</h1>
            <p className="text-sm text-balance text-muted-foreground">
              Enter your email below to login to your account
            </p>
          </div>
          <Field>
            <FieldLabel htmlFor="email">Email</FieldLabel>
            <Input
              id="email"
              type="email"
              placeholder="m@example.com"
              autoComplete="email"
              {...register('email')}
            />
            <FieldError errors={[formState.errors.email]} />
          </Field>
          <Field>
            <div className="flex items-center">
              <FieldLabel htmlFor="password">Password</FieldLabel>
              <a href="#" className="ml-auto text-sm underline-offset-4 hover:underline">
                Forgot your password?
              </a>
            </div>
            <Input
              id="password"
              type="password"
              autoComplete="current-password"
              {...register('password')}
            />
            <FieldError errors={[formState.errors.password]} />
          </Field>
          {message ? <FieldError>{message}</FieldError> : null}
          <Field>
            <Button type="submit" disabled={formState.isSubmitting}>
              Login
            </Button>
          </Field>
          <FieldSeparator>Or continue with</FieldSeparator>
          <Field>
            <Button variant="outline" type="button">
              Login with GitHub
            </Button>
            <FieldDescription className="text-center">
              Don&apos;t have an account?{' '}
              <Link to="/signup" className="underline underline-offset-4">
                Sign up
              </Link>
            </FieldDescription>
          </Field>
        </FieldGroup>
      </form>
    </AuthLayout>
  )
}
