import { ChevronRightIcon } from 'lucide-react'
import { Fragment } from 'react'
import type { ReactNode } from 'react'
import { Label } from '@/components/ui/label'
import type { SettingField as SettingFieldModel } from '@/features/settings/schema'
import { useText } from '@/locales'

type SettingFieldProps = {
  field: SettingFieldModel
  control: ReactNode
}

// SettingField renders one setting as a VSCode style card: the title carries the nested path separated by
// chevrons, the description comes straight from the schema and the control takes the last row.
export function SettingField({ field, control }: SettingFieldProps) {
  const text = useText()

  return (
    <div className="flex flex-col gap-2 rounded-lg p-3 transition-colors hover:bg-muted/60">
      <div className="flex flex-col gap-1">
        <Label htmlFor={field.path} className="flex items-center gap-1 text-sm font-medium">
          {field.title.map((segment, index) => (
            <Fragment key={`${segment}-${index}`}>
              {index > 0 ? (
                <ChevronRightIcon className="size-3.5 shrink-0 text-muted-foreground" />
              ) : null}
              <span>{segment}</span>
            </Fragment>
          ))}
        </Label>
        {field.description ? (
          <p className="text-xs text-muted-foreground">{field.description}</p>
        ) : null}
        {field.invalid ? (
          <p className="text-xs text-destructive">{text.settings.invalidValue}</p>
        ) : null}
      </div>
      {control}
    </div>
  )
}
