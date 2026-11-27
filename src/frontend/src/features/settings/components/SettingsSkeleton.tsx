import { Skeleton } from '@/components/ui/skeleton'

// SettingsSkeleton stands in for the panel while the document is on its way.
//
// The shapes follow the real layout rather than being generic bars: a group list on the left, and cards on the right
// with the title, the description and the control stacked the way SettingField stacks them. Parsing the schema is
// quick but not instant, and a skeleton that matches the result means the panel does not jump when it arrives.
export function SettingsSkeleton() {
  return (
    <>
      {/* Sidebar: one row per group, each a label beside its icon. */}
      <div className="hidden w-52 shrink-0 flex-col gap-3 border-r border-sidebar-border p-2 md:flex">
        {Array.from({ length: 4 }, (_, index) => (
          <div key={index} className="flex items-center gap-2 px-2 py-1.5">
            <Skeleton className="size-4 shrink-0 rounded-sm" />
            <Skeleton className="h-3.5" style={{ width: `${[5, 6, 4, 7][index] * 8}px` }} />
          </div>
        ))}
      </div>

      <main className="flex h-full flex-1 flex-col overflow-hidden">
        <header className="flex h-16 shrink-0 items-center gap-2 px-4">
          <Skeleton className="h-3.5 w-16" />
          <Skeleton className="size-3 rounded-sm" />
          <Skeleton className="h-3.5 w-20" />
        </header>
        <div className="flex flex-1 flex-col gap-1 overflow-hidden p-4 pt-2">
          {Array.from({ length: 5 }, (_, index) => (
            <div key={index} className="flex flex-col gap-2 rounded-lg p-3">
              <div className="flex flex-col gap-1.5">
                <Skeleton className="h-4" style={{ width: `${[96, 120, 80, 136, 104][index]}px` }} />
                <Skeleton className="h-3" style={{ width: `${[224, 256, 192, 240, 208][index]}px` }} />
              </div>
              {/* The control row is the tallest part of a card: an input, a quantity pair or a switch. */}
              <Skeleton className="h-8 w-56" />
            </div>
          ))}
        </div>
      </main>
    </>
  )
}
