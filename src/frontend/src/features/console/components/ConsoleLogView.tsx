import { useVirtualizer } from '@tanstack/react-virtual'
import { ArrowDownToLineIcon, EraserIcon } from 'lucide-react'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Button } from '@/components/ui/button'
import { ButtonGroup } from '@/components/ui/button-group'
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Toggle } from '@/components/ui/toggle'
import { ToggleGroup } from '@/components/ui/toggle-group'
import { ConsoleHeader } from '@/features/console/components/ConsoleHeader'
import { useText } from '@/locales'
import { loadSettings } from '@/mocks'
import { greetService } from '@/services/greet'
import { logService, type LogListenerDto, type LogRecordDto } from '@/services/log'

// levels are the four severities, in the order they escalate. The value is the slog.Level the backend stores, so the
// numbers travel through the binding unchanged and only the label is translated.
const levels = [
  { key: 'debug', value: -4 },
  { key: 'info', value: 0 },
  { key: 'warn', value: 4 },
  { key: 'error', value: 8 },
] as const

type LevelKey = (typeof levels)[number]['key']

// levelClass colours the marker of each severity, following the usual console order of blue, green, yellow and
// red. The 400 step stays legible on both the light and the dark surface.
const levelClass: Record<number, string> = {
  [-4]: 'text-blue-400',
  0: 'text-emerald-400',
  4: 'text-amber-400',
  8: 'text-red-400',
}

const levelLabel: Record<number, LevelKey> = {
  [-4]: 'debug',
  0: 'info',
  4: 'warn',
  8: 'error',
}

// fallbackInterval is what the view polls at until the stored preference arrives.
const fallbackInterval = 1000

// debounce is how long the view waits after a filter change before it asks the backend again. Typing a pattern
// fires one change per keystroke, and each one would otherwise cost a round trip.
const debounce = 250

// formatStamp renders the millisecond instant as the standard UTC form, which is what the records are compared by.
function formatStamp(milliseconds: number): string {
  if (!milliseconds) {
    return ''
  }
  return new Date(milliseconds).toISOString().replace('.000Z', 'Z')
}

type ConsoleLogViewProps = {
  /** leading replaces the listener picker, so a caller can attach a different source without touching the view. */
  leading?: React.ReactNode
}

// ConsoleLogView is the log panel of the console.
//
// The backend owns the records, the boundaries and the filters; this view owns the toolbar state and the polling. A
// change to the toolbar is sent to the backend once it settles, then the list is read again immediately so the result
// appears without waiting for the next tick.
export function ConsoleLogView({ leading }: ConsoleLogViewProps) {
  const text = useText()
  const [listeners, setListeners] = useState<LogListenerDto[]>([])
  const [listener, setListener] = useState('')
  const [enabled, setEnabled] = useState<LevelKey[]>(levels.map((entry) => entry.key))
  const [pattern, setPattern] = useState('')
  const [autoScroll, setAutoScroll] = useState(true)
  const [records, setRecords] = useState<LogRecordDto[]>([])
  const [refreshInterval, setRefreshInterval] = useState(fallbackInterval)

  const scrollRef = useRef<HTMLDivElement>(null)
  const virtualizer = useVirtualizer({
    count: records.length,
    getScrollElement: () => scrollRef.current,
    // An estimate only seeds the layout; measureElement replaces it with the real height once a row is drawn.
    estimateSize: () => 26,
    overscan: 12,
  })

  // The stored preferences decide how often to poll and which levels to tick at first.
  useEffect(() => {
    let cancelled = false
    loadSettings()
      .then((document) => {
        if (cancelled) {
          return
        }
        const log = (document.values.general as Record<string, unknown> | undefined)?.log as
          | Record<string, unknown>
          | undefined
        if (!log) {
          return
        }
        setAutoScroll(log.autoScroll !== false)
        setRefreshInterval(readInterval(log.refreshInterval))
      })
      .catch(() => undefined)
    return () => {
      cancelled = true
    }
  }, [])

  // The panel seeds a sample batch before its first read. The boundary of every source is the moment the log service
  // came up, which is later than the records written during start up, so without this the list opens empty.
  useEffect(() => {
    void greetService.emitLogs().catch(() => undefined)
  }, [])

  // The sources come from the backend, so the picker lists whatever has actually written a record.
  useEffect(() => {
    let cancelled = false
    void logService
      .listeners()
      .then((entries) => {
        if (cancelled || entries.length === 0) {
          return
        }
        setListeners(entries)
        setListener((current) => current || entries[0].name)
      })
      .catch(() => undefined)
    return () => {
      cancelled = true
    }
  }, [])

  const load = useCallback(async (source: string) => {
    if (source === '') {
      return
    }
    try {
      const answer = await logService.records(source)
      setRecords(answer.records)
    } catch {
      // A failed read leaves the list as it was: a transient backend error should not blank the panel.
    }
  }, [])

  // Poll on the configured interval. The effect restarts when the source or the interval changes, so a switch never
  // leaves a timer pointing at the previous source.
  useEffect(() => {
    void load(listener)
    const timer = window.setInterval(() => void load(listener), refreshInterval)
    return () => window.clearInterval(timer)
  }, [listener, refreshInterval, load])

  // Filter changes are debounced: the parameters are sent once the user stops moving, then the list is read again so
  // the result is immediate rather than waiting for the next tick.
  useEffect(() => {
    const allowed = levels.filter((entry) => enabled.includes(entry.key)).map((entry) => entry.value)
    const timer = window.setTimeout(() => {
      void logService
        .setFilter({ levels: allowed, pattern })
        .then(() => load(listener))
        .catch(() => undefined)
    }, debounce)
    return () => window.clearTimeout(timer)
  }, [enabled, pattern, listener, load])

  // While auto scroll is on, every new record pulls the view to the last one. The effect runs after the list has been
  // measured, which is why it can address the final index rather than guessing an offset.
  useEffect(() => {
    if (!autoScroll || records.length === 0) {
      return
    }
    virtualizer.scrollToIndex(records.length - 1, { align: 'end' })
  }, [autoScroll, records.length, virtualizer])

  const toggleLevel = (level: LevelKey) => {
    setEnabled((previous) =>
      previous.includes(level) ? previous.filter((entry) => entry !== level) : [...previous, level],
    )
  }

  const clear = useCallback(() => {
    if (listener === '') {
      return
    }
    void logService
      .clear(listener)
      .then(() => load(listener))
      .catch(() => undefined)
  }, [listener, load])

  const current = useMemo(() => listeners.find((entry) => entry.name === listener), [listeners, listener])

  return (
    <div className="flex min-h-0 min-w-0 flex-1 flex-col">
      <ConsoleHeader
        leading={
          <>
            {leading ?? (
              <Select
                value={listener}
                onValueChange={(next) => setListener(String(next))}
                items={listeners.map((entry) => ({ value: entry.name, label: entry.title }))}
              >
                {/* Height and text size are declared here rather than through size="sm": the compact preset
                    carries its own padding, which left the label sitting higher than the neighbours. */}
                <SelectTrigger className="h-8 w-36 shrink-0 text-xs">
                  <SelectValue placeholder={current?.title ?? text.console.log.listener} />
                </SelectTrigger>
                <SelectContent alignItemWithTrigger={false}>
                  {listeners.map((entry) => (
                    <SelectItem key={entry.name} value={entry.name}>
                      {entry.title}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}

            <ButtonGroup className="min-w-0 flex-1">
              <DropdownMenu>
                <DropdownMenuTrigger
                  render={
                    <Button
                      variant="outline"
                      aria-label={text.console.log.filter.level}
                      className="h-8 w-8 shrink-0 rounded-r-none border-r-0 p-0"
                    >
                      <FilterIcon />
                    </Button>
                  }
                />
                <DropdownMenuContent align="start" className="min-w-40">
                  {levels.map((entry) => (
                    <DropdownMenuCheckboxItem
                      key={entry.key}
                      checked={enabled.includes(entry.key)}
                      onCheckedChange={() => toggleLevel(entry.key)}
                      closeOnClick={false}
                    >
                      {text.console.log.levels[entry.key]}
                    </DropdownMenuCheckboxItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>

              {/* One text size across the row: the input default is larger than the compact controls beside it,
                  which made the row look vertically off-centre even when every box shared a height. */}
              <Input
                value={pattern}
                onChange={(event) => setPattern(event.target.value)}
                placeholder={text.console.log.filter.pattern}
                aria-label={text.console.log.filter.patternLabel}
                className="h-8 min-w-0 flex-1 rounded-l-none text-xs"
              />
            </ButtonGroup>
          </>
        }
        trailing={
          <>
            <Button variant="ghost" aria-label={text.console.log.actions.clear} className="h-8 w-8 p-0" onClick={clear}>
              <EraserIcon />
            </Button>
            <ToggleGroup
              value={autoScroll ? ['auto-scroll'] : []}
              onValueChange={(next) => setAutoScroll(next.includes('auto-scroll'))}
            >
              {/* No size preset: the compact one carries its own padding, and the row is aligned by height alone so
                  every control shares one vertical centre. */}
              <Toggle value="auto-scroll" aria-label={text.console.log.actions.autoScroll} className="h-8 w-8">
                <ArrowDownToLineIcon />
              </Toggle>
            </ToggleGroup>
          </>
        }
      />

      {records.length === 0 ? (
        <div className="flex min-h-0 min-w-0 flex-1 flex-col items-center justify-center gap-1 p-4 text-center">
          <p className="text-sm font-medium text-foreground">{text.console.empty.title}</p>
          <p className="max-w-sm text-xs text-muted-foreground">{text.console.empty.description}</p>
        </div>
      ) : (
        <div ref={scrollRef} className="min-h-0 min-w-0 flex-1 overflow-y-auto overflow-x-hidden">
          {/* The spacer carries the full scroll height; each row is absolutely placed by the virtualiser, so the list
              never allocates a node per record. */}
          <div className="relative w-full" style={{ height: virtualizer.getTotalSize() }}>
            {virtualizer.getVirtualItems().map((item) => {
              const record = records[item.index]
              return (
                <div
                  key={record.id}
                  data-index={item.index}
                  ref={virtualizer.measureElement}
                  className="absolute inset-x-0 top-0 px-2 py-1 font-mono text-xs leading-5 break-words whitespace-pre-wrap hover:bg-accent/60"
                  style={{ transform: `translateY(${item.start}px)` }}
                >
                  {/* The stamp and the level share one line; the message and then the detail follow as blocks, so a
                      multi line message never pushes the identity line out of column. */}
                  <div className="flex items-baseline gap-2">
                    <span className="text-muted-foreground">{formatStamp(record.createdAt)}</span>
                    <span className={levelClass[record.level] ?? 'text-foreground'}>
                      {text.console.log.levels[levelLabel[record.level] ?? 'info']}
                    </span>
                  </div>
                  <div className="text-foreground">{record.msg}</div>
                  {record.detail ? <div className="text-muted-foreground">{record.detail}</div> : null}
                </div>
              )
            })}
          </div>
        </div>
      )}
    </div>
  )
}

// readInterval reads a value written as "<number> <unit>", falling back to one second.
function readInterval(value: unknown): number {
  const text = String(value ?? '').trim()
  const match = /^(\d+(?:\.\d+)?)\s*(ms|s|m)?$/i.exec(text)
  if (!match) {
    return fallbackInterval
  }
  const amount = Number(match[1])
  const unit = (match[2] ?? 's').toLowerCase()
  if (unit === 'ms') {
    return amount
  }
  if (unit === 'm') {
    return amount * 60_000
  }
  return amount * 1000
}

// FilterIcon is the funnel that marks the level menu. It is inline rather than imported so the component keeps one
// icon dependency for its own toolbar.
function FilterIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="size-4">
      <path d="M4 5h16l-6 7v6l-4 2v-8L4 5Z" strokeLinejoin="round" />
    </svg>
  )
}
