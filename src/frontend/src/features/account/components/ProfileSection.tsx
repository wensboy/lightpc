import { zodResolver } from '@hookform/resolvers/zod'
import { EraserIcon, PencilIcon } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import type { ChangeEvent } from 'react'
import { useForm } from 'react-hook-form'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import { FieldError } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Toggle } from '@/components/ui/toggle'
import { SettingGroupCard } from '@/features/account/layouts/SettingGroupCard'
import { profileSchema, type ProfileValues } from '@/features/account/models/schema'
import { useText } from '@/locales'
import { loadAccountAvatar, loadAccountProfile } from '@/mocks'

// ProfileSection pairs the identity fields on the left with the avatar on the right. The card header
// switches the section between read only and editable, and the clear action empties the editable fields
// while keeping the immutable uid untouched.
export function ProfileSection() {
  const text = useText()
  const [editMode, setEditMode] = useState(false)
  const [avatar, setAvatar] = useState(loadAccountAvatar)
  const previewRef = useRef<string | null>(null)
  const fileRef = useRef<HTMLInputElement | null>(null)

  const { register, formState, getValues, reset } = useForm<ProfileValues>({
    resolver: zodResolver(profileSchema),
    defaultValues: loadAccountProfile(),
    mode: 'onBlur',
  })

  // Release the last object URL when the section unmounts or a new preview replaces it.
  useEffect(
    () => () => {
      if (previewRef.current) {
        URL.revokeObjectURL(previewRef.current)
      }
    },
    [],
  )

  const handleFileChange = (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (!file) {
      return
    }
    if (previewRef.current) {
      URL.revokeObjectURL(previewRef.current)
    }
    const next = URL.createObjectURL(file)
    previewRef.current = next
    setAvatar(next)
  }

  const handleClear = () => {
    // uid identifies the account, so a clear keeps it and only empties the editable fields.
    reset({ uid: getValues('uid'), uname: '', uaccount: '' })
  }

  return (
    <SettingGroupCard
      title={text.account.profile.groupTitle}
      actions={
        <Button
          variant="ghost"
          size="icon-sm"
          aria-label={text.account.clear}
          disabled={!editMode}
          onClick={handleClear}
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
      <div className="flex flex-col gap-6 md:flex-row md:items-start">
        <div className="flex flex-1 flex-col gap-3">
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="uid">{text.account.profile.uidLabel}</Label>
            <Input
              id="uid"
              placeholder={text.account.profile.uidPlaceholder}
              disabled={!editMode}
              {...register('uid')}
            />
            <FieldError errors={[formState.errors.uid]} />
          </div>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="uname">{text.account.profile.nameLabel}</Label>
            <Input
              id="uname"
              placeholder={text.account.profile.namePlaceholder}
              disabled={!editMode}
              {...register('uname')}
            />
            <FieldError errors={[formState.errors.uname]} />
          </div>
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="uaccount">{text.account.profile.accountLabel}</Label>
            <Input
              id="uaccount"
              placeholder={text.account.profile.accountPlaceholder}
              disabled={!editMode}
              {...register('uaccount')}
            />
            <FieldError errors={[formState.errors.uaccount]} />
          </div>
        </div>
        <div className="flex shrink-0 flex-col items-center gap-2">
          <button
            type="button"
            aria-label={text.account.profile.avatarLabel}
            disabled={!editMode}
            className="rounded-full outline-none focus-visible:ring-3 focus-visible:ring-ring/50 disabled:cursor-not-allowed disabled:opacity-60"
            onClick={() => fileRef.current?.click()}
          >
            <Avatar className="size-28">
              <AvatarImage src={avatar} alt={text.account.profile.avatarLabel} />
              <AvatarFallback className="text-xl">
                {text.account.profile.avatarFallback}
              </AvatarFallback>
            </Avatar>
          </button>
          <input
            ref={fileRef}
            type="file"
            accept="image/*"
            className="hidden"
            onChange={handleFileChange}
          />
          <p className="max-w-32 text-center text-xs text-muted-foreground">
            {text.account.profile.avatarHint}
          </p>
        </div>
      </div>
    </SettingGroupCard>
  )
}
