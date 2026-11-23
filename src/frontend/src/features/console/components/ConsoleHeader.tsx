import type { ReactNode } from 'react'

type ConsoleHeaderProps = {
  /** leading sits on the left and grows towards the right. */
  leading?: ReactNode
  /** trailing sits on the right and grows towards the left. */
  trailing?: ReactNode
}

// ConsoleHeader is the bar above a console view. It owns nothing but the layout: the two halves are pushed to
// the opposite edges and each one grows away from its edge, so a view can place a title, a filter or a set of
// actions without repeating the arrangement.
export function ConsoleHeader({ leading, trailing }: ConsoleHeaderProps) {
  return (
    <header className="flex h-9 shrink-0 items-center justify-between gap-3 border-b border-border/60 px-3">
      <div className="flex min-w-0 flex-1 items-center gap-2">{leading}</div>
      <div className="flex min-w-0 items-center justify-end gap-2">{trailing}</div>
    </header>
  )
}
