import type { ReactNode } from 'react'
import { Label } from '@/components/ui/label'
import { useText } from '@/locales'
import { resolveText } from '@/locales/text'

type SettingFieldProps = {
  id: string
  categoryLabel: string
  labelKey: string
  descriptionKey?: string
  control: ReactNode
}

// SettingField renders one setting as a VSCode style card: the title carries the category breadcrumb,
// the description sits right below it and the control takes the last row. Grouping comes from the
// breadcrumb, so no separator between items is needed.
export function SettingField({
  id,
  categoryLabel,
  labelKey,
  descriptionKey,
  control,
}: SettingFieldProps) {
  const text = useText()

  return (
    <div className="flex flex-col gap-2 rounded-lg p-3 transition-colors hover:bg-muted/60">
      <div className="flex flex-col gap-1">
        <Label htmlFor={id} className="text-sm font-medium">
          <span className="text-primary">{categoryLabel}</span>
          <span className="text-muted-foreground">: </span>
          {resolveText(text, labelKey)}
        </Label>
        {descriptionKey ? (
          <p className="text-xs text-muted-foreground">{resolveText(text, descriptionKey)}</p>
        ) : null}
      </div>
      {control}
    </div>
  )
}
