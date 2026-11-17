import { Link } from '@tanstack/react-router'
import { Button } from '@/components/ui/button'
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
  FieldSeparator,
} from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { AuthLayout } from '@/features/auth/layouts/AuthLayout'
import { useText } from '@/locales'

// SignupPage is the registration shell; the form is wired to AuthService once the backend exposes it.
export function SignupPage() {
  const text = useText()

  return (
    <AuthLayout>
      <form className="flex flex-col gap-6">
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
              required
              className="bg-background"
            />
          </Field>
          <Field>
            <FieldLabel htmlFor="email">{text.auth.signup.emailLabel}</FieldLabel>
            <Input
              id="email"
              type="email"
              placeholder="m@example.com"
              required
              className="bg-background"
            />
            <FieldDescription>{text.auth.signup.emailHint}</FieldDescription>
          </Field>
          <Field>
            <FieldLabel htmlFor="password">{text.auth.signup.passwordLabel}</FieldLabel>
            <Input id="password" type="password" required className="bg-background" />
            <FieldDescription>{text.auth.signup.passwordHint}</FieldDescription>
          </Field>
          <Field>
            <FieldLabel htmlFor="confirm-password">{text.auth.signup.confirmLabel}</FieldLabel>
            <Input id="confirm-password" type="password" required className="bg-background" />
            <FieldDescription>{text.auth.signup.confirmHint}</FieldDescription>
          </Field>
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
