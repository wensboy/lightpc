import type { ReactNode } from 'react'

type ConsoleHeaderProps = {
  /** leading sits on the left half and grows towards the right. */
  leading?: ReactNode
  /** trailing sits on the right half and grows towards the left. */
  trailing?: ReactNode
}

// ConsoleHeader is the bar above a console view. It owns nothing but the layout: the two halves are equally
// wide and each one grows away from its own edge, so a view can place a title, a filter or a set of actions
// without repeating the arrangement.
//
// The padding and the gap are the same value: a control then sits as far from the edge as it does from its
// neighbour, which is what makes the toolbar read as one row rather than a set of loosely placed items.
//
// The bar is taller than the controls it holds: at 36px a 32px control leaves only 2px above and below, which
// reads as an accident rather than a margin. 40px gives every control an even 4px.
export function ConsoleHeader({ leading, trailing }: ConsoleHeaderProps) {
  return (
    <header className="flex h-10 shrink-0 items-center gap-2 border-b border-border/60 px-2">
      <div className="flex min-w-0 flex-1 basis-1/2 items-center gap-2">{leading}</div>
      <div className="flex min-w-0 flex-1 basis-1/2 items-center justify-end gap-2">{trailing}</div>
    </header>
  )
}
