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
import { writeAuth } from '@/features/auth/session'
import { useText } from '@/locales'
import { sessionService } from '@/services/session'
import { loginSchema, type LoginInput } from '@/services/auth'

// LoginPage validates with Zod through React Hook Form and hands the payload to the auth service.
export function LoginPage() {
  const text = useText()
  const navigate = useNavigate()
  const [message, setMessage] = useState('')
  const { register, handleSubmit, formState } = useForm<LoginInput>({
    resolver: zodResolver(loginSchema()),
    defaultValues: { email: '', password: '' },
  })

  const onSubmit = handleSubmit(async (values) => {
    setMessage('')
    try {
      const session = await sessionService.login({ account: values.email, password: values.password })
      writeAuth({ uid: session.uid, account: session.account })
      await navigate({ to: '/' })
    } catch (error) {
      setMessage(error instanceof Error ? error.message : text.auth.errors.loginFailed)
    }
  })

  return (
    <AuthLayout>
      <form onSubmit={onSubmit} className="flex flex-col gap-6" noValidate>
        <FieldGroup>
          <div className="flex flex-col items-center gap-1 text-center">
            <h1 className="text-2xl font-bold">{text.auth.login.title}</h1>
            <p className="text-sm text-balance text-muted-foreground">
              {text.auth.login.description}
            </p>
          </div>
          <Field>
            <FieldLabel htmlFor="email">{text.auth.login.accountLabel}</FieldLabel>
            <Input
              id="email"
              type="email"
              placeholder={text.auth.login.accountPlaceholder}
              autoComplete="email"
              className="bg-background"
              {...register('email')}
            />
            <FieldError errors={[formState.errors.email]} />
          </Field>
          <Field>
            <div className="flex items-center">
              <FieldLabel htmlFor="password">{text.auth.login.passwordLabel}</FieldLabel>
              <a href="#" className="ml-auto text-sm underline-offset-4 hover:underline">
                {text.auth.login.forgotPassword}
              </a>
            </div>
            <Input
              id="password"
              type="password"
              autoComplete="current-password"
              className="bg-background"
              {...register('password')}
            />
            <FieldError errors={[formState.errors.password]} />
          </Field>
          {message ? <FieldError>{message}</FieldError> : null}
          <Field>
            <Button type="submit" disabled={formState.isSubmitting}>
              {text.auth.login.submit}
            </Button>
          </Field>
          <FieldSeparator>{text.auth.login.separator}</FieldSeparator>
          <Field>
            <Button variant="outline" type="button">
              {text.auth.login.github}
            </Button>
            <FieldDescription className="text-center">
              {text.auth.login.noAccount}{' '}
              <Link to="/signup" className="underline underline-offset-4">
                {text.auth.login.signupLink}
              </Link>
            </FieldDescription>
          </Field>
        </FieldGroup>
      </form>
    </AuthLayout>
  )
}
