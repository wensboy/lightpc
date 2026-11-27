import { useMemo, useState } from 'react'
import { Sheet, SheetContent, SheetDescription, SheetTitle } from '@/components/ui/sheet'
import { Sidebar, SidebarContent, SidebarProvider } from '@/components/ui/sidebar'
import { ConsoleHeader } from '@/features/console/components/ConsoleHeader'
import { ConsoleLogView } from '@/features/console/components/ConsoleLogView'
import { ConsoleNav } from '@/features/console/components/ConsoleNav'
import { buildConsoleNav, type ConsoleSectionKey } from '@/features/console/config/nav'
import { useText } from '@/locales'
import { useDialogStore } from '@/stores/dialog'

// ConsoleSheet is the console overlay. It docks to the bottom edge and is nothing but an icon rail next to a
// content area.
//
// The rail keeps the collapsed width and never expands, so the label lives in a tooltip. That tooltip only
// shows while the provider reports the collapsed state, which is why the provider opens collapsed. The active
// entry is marked by the same surface treatment the main sidebar uses.
//
// The views differ a lot, so each one currently renders an empty state under a shared header; the rail already
// carries the selection so a view can be filled in without touching this layout.
export function ConsoleSheet() {
  const text = useText()
  const consoleOpen = useDialogStore((state) => state.consoleOpen)
  const closeConsole = useDialogStore((state) => state.closeConsole)
  const [active, setActive] = useState<ConsoleSectionKey>('log')

  const items = useMemo(() => buildConsoleNav(text), [text])
  const current = items.find((item) => item.key === active) ?? items[0]

  return (
    <Sheet
      open={consoleOpen}
      onOpenChange={(next) => {
        if (!next) {
          closeConsole()
        }
      }}
    >
      <SheetContent side="bottom" showCloseButton={false} className="h-[40vh]! gap-0 p-0">
        <SheetTitle className="sr-only">{text.console.title}</SheetTitle>
        <SheetDescription className="sr-only">{text.console.description}</SheetDescription>
        <SidebarProvider defaultOpen={false} className="h-full min-h-0 w-full min-w-0 items-stretch">
          {/* The rail scrolls on its own and hides the scrollbar, like the main sidebar content. */}
          <Sidebar collapsible="none" className="h-full w-12 min-w-12 border-r border-sidebar-border">
            <SidebarContent className="items-center px-1 py-2">
              <ConsoleNav items={items} active={active} onSelect={setActive} />
            </SidebarContent>
          </Sidebar>
          <main className="flex h-full min-w-0 flex-1 flex-col overflow-hidden">
            {active === 'log' ? (
              <ConsoleLogView />
            ) : (
              <>
                <ConsoleHeader leading={<span className="truncate text-sm font-medium">{current?.title}</span>} />
                <div className="flex min-h-0 min-w-0 flex-1 flex-col items-center justify-center gap-1 p-4 text-center">
                  <p className="text-sm font-medium text-foreground">{text.console.pending.title}</p>
                  <p className="max-w-sm text-xs text-muted-foreground">{text.console.pending.description}</p>
                </div>
              </>
            )}
          </main>
        </SidebarProvider>
      </SheetContent>
    </Sheet>
  )
}
