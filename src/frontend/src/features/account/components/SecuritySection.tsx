import { zodResolver } from '@hookform/resolvers/zod'
import { EraserIcon, PencilIcon } from 'lucide-react'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { Button } from '@/components/ui/button'
import { FieldError } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Toggle } from '@/components/ui/toggle'
import { SettingGroupCard } from '@/features/account/components/common/SettingGroupCard'
import { securitySchema, type SecurityValues } from '@/features/account/schema'
import { useText } from '@/locales'

// SecuritySection collects the current password and its replacement. Validation runs on blur, so the
// mismatch between the new password and its confirmation shows up as soon as the field is left.
export function SecuritySection() {
  const text = useText()
  const [editMode, setEditMode] = useState(false)
  const { register, formState, reset } = useForm<SecurityValues>({
    resolver: zodResolver(securitySchema),
    defaultValues: { oldPassword: '', newPassword: '', checkPassword: '' },
    mode: 'onBlur',
  })

  return (
    <SettingGroupCard
      title={text.account.security.groupTitle}
      actions={
        <Button
          variant="ghost"
          size="icon-sm"
          aria-label={text.account.clear}
          disabled={!editMode}
          onClick={() => reset()}
        >
          <EraserIcon />
        </Button>
      }
      toggle={
        <Toggle
          size="sm"
          pressed={editMode}
          onPressedChange={setEditMode}
          aria-label={text.account.editMode}
        >
          <PencilIcon />
        </Toggle>
      }
    >
      <div className="flex max-w-sm flex-col gap-3">
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="oldPassword">{text.account.security.oldLabel}</Label>
          <Input
            id="oldPassword"
            type="password"
            autoComplete="current-password"
            placeholder={text.account.security.oldPlaceholder}
            disabled={!editMode}
            {...register('oldPassword')}
          />
          <FieldError errors={[formState.errors.oldPassword]} />
        </div>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="newPassword">{text.account.security.newLabel}</Label>
          <Input
            id="newPassword"
            type="password"
            autoComplete="new-password"
            placeholder={text.account.security.newPlaceholder}
            disabled={!editMode}
            {...register('newPassword')}
          />
          <FieldError errors={[formState.errors.newPassword]} />
        </div>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="checkPassword">{text.account.security.checkLabel}</Label>
          <Input
            id="checkPassword"
            type="password"
            autoComplete="new-password"
            placeholder={text.account.security.checkPlaceholder}
            disabled={!editMode}
            {...register('checkPassword')}
          />
          <FieldError errors={[formState.errors.checkPassword]} />
        </div>
      </div>
    </SettingGroupCard>
  )
}
