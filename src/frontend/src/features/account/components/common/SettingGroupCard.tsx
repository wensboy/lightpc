import type { ReactNode } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

type SettingGroupCardProps = {
  title: string
  // actions renders to the left of the toggle; the component fixes the order so callers only pass the
  // icon buttons.
  actions?: ReactNode
  toggle?: ReactNode
  children: ReactNode
}

// SettingGroupCard wraps a group of settings. The header splits in half: the title truncates on the
// left, while the action panel docks to the right with icon buttons first and the toggle last.
export function SettingGroupCard({ title, actions, toggle, children }: SettingGroupCardProps) {
  return (
    <Card size="sm" className="gap-3 py-3">
      <CardHeader className="flex flex-row items-center justify-between gap-4 py-0">
        <CardTitle className="min-w-0 basis-1/2 truncate text-sm font-medium">{title}</CardTitle>
        <div className="flex basis-1/2 shrink-0 items-center justify-end gap-1">
          {actions}
          {toggle}
        </div>
      </CardHeader>
      <CardContent>{children}</CardContent>
    </Card>
  )
}
