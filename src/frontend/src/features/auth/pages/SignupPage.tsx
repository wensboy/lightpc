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

type SignupValues = {
  name: string
  email: string
  password: string
  confirm: string
}

// SignupPage creates the account, which also signs it in: the backend answers with the same status a login
// does, so the screen only has to store the token and navigate.
export function SignupPage() {
  const text = useText()
  const navigate = useNavigate()
  const [message, setMessage] = useState('')
  const { register, handleSubmit } = useForm<SignupValues>({
    defaultValues: { name: '', email: '', password: '', confirm: '' },
  })

  const onSubmit = handleSubmit(async (values) => {
    setMessage('')
    if (values.password !== values.confirm) {
      setMessage(text.auth.errors.shortPassword)
      return
    }
    try {
      const session = await sessionService.register({
        uname: values.name,
        account: values.email,
        password: values.password,
      })
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
            <h1 className="text-2xl font-bold">{text.auth.signup.title}</h1>
            <p className="text-sm text-balance text-muted-foreground">
              {text.auth.signup.description}
            </p>
          </div>
          <Field>
            <FieldLabel htmlFor="name">{text.auth.signup.nameLabel}</FieldLabel>
            <Input
              id="name"
              type="text"
              placeholder={text.auth.signup.namePlaceholder}
              autoComplete="name"
              className="bg-background"
              {...register('name', { required: true })}
            />
          </Field>
          <Field>
            <FieldLabel htmlFor="email">{text.auth.signup.accountLabel}</FieldLabel>
            <Input
              id="email"
              type="email"
              placeholder={text.auth.signup.accountPlaceholder}
              autoComplete="email"
              className="bg-background"
              {...register('email', { required: true })}
            />
            <FieldDescription>{text.auth.signup.accountHint}</FieldDescription>
          </Field>
          <Field>
            <FieldLabel htmlFor="password">{text.auth.signup.passwordLabel}</FieldLabel>
            <Input
              id="password"
              type="password"
              autoComplete="new-password"
              className="bg-background"
              {...register('password', { required: true, minLength: 8 })}
            />
            <FieldDescription>{text.auth.signup.passwordHint}</FieldDescription>
          </Field>
          <Field>
            <FieldLabel htmlFor="confirm-password">{text.auth.signup.confirmLabel}</FieldLabel>
            <Input
              id="confirm-password"
              type="password"
              autoComplete="new-password"
              className="bg-background"
              {...register('confirm', { required: true, minLength: 8 })}
            />
            <FieldDescription>{text.auth.signup.confirmHint}</FieldDescription>
          </Field>
          {message ? <FieldError>{message}</FieldError> : null}
          <Field>
            <Button type="submit">{text.auth.signup.submit}</Button>
          </Field>
          <FieldSeparator>{text.auth.signup.separator}</FieldSeparator>
          <Field>
            <Button variant="outline" type="button">
              {text.auth.signup.github}
            </Button>
            <FieldDescription className="px-6 text-center">
              {text.auth.signup.hasAccount}{' '}
              <Link to="/login" className="underline underline-offset-4">
                {text.auth.signup.loginLink}
              </Link>
            </FieldDescription>
          </Field>
        </FieldGroup>
      </form>
    </AuthLayout>
  )
}
