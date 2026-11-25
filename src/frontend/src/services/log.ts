import { z } from 'zod'
import { LogService as LogBinding } from '../../bindings/github.com/wensboy/lightpc/backend/service/log'

// recordSchema validates one row the backend returns. Level is the slog.Level value unchanged, so it stays a number
// here and is turned into a name only where it is rendered.
const recordSchema = z.object({
  id: z.number(),
  listener: z.string(),
  level: z.number(),
  msg: z.string(),
  detail: z.string().default(''),
  createdAt: z.number(),
})

// recordsSchema validates one read. Boundary travels with the rows so the view can show where "cleared" currently
// stands without asking a second question.
const recordsSchema = z.object({
  records: z.array(recordSchema).default([]),
  boundary: z.number().default(0),
})

// listenerSchema validates one source the picker offers.
const listenerSchema = z.object({
  name: z.string(),
  title: z.string(),
})

const listenersSchema = z.array(listenerSchema)

/** LogRecordDto is one record the console shows. */
export type LogRecordDto = z.infer<typeof recordSchema>

/** LogRecordsDto is the answer of one read. */
export type LogRecordsDto = z.infer<typeof recordsSchema>

/** LogListenerDto is one source the picker offers. */
export type LogListenerDto = z.infer<typeof listenerSchema>

/** LogFilterParamsDto carries every filter parameter in one payload. */
export type LogFilterParamsDto = {
  /** Levels allowed through. An empty list disables the level filter. */
  levels: number[]
  /** Search text. An empty string disables the pattern filter. */
  pattern: string
}

/** LogService is the contract the console feature depends on. */
export interface LogService {
  /** listeners returns every source that has written a record. */
  listeners(): Promise<LogListenerDto[]>
  /** records returns the rows of one source from its boundary onwards, oldest first. */
  records(listener: string, limit?: number): Promise<LogRecordsDto>
  /** clear moves the boundary of one source to now. */
  clear(listener: string): Promise<void>
  /** setFilter stores the parameters the backend filters read on the next call. */
  setFilter(params: LogFilterParamsDto): Promise<void>
}

// logService talks to the backend. The filters live there, so the view never decides what a level or a pattern
// means; it sends the toolbar state and renders whatever comes back.
export const logService: LogService = {
  async listeners() {
    const parsed = listenersSchema.safeParse(await LogBinding.Listeners())
    if (!parsed.success) {
      throw new Error(`log listeners rejected: ${parsed.error.message}`)
    }
    return parsed.data
  },

  async records(listener, limit = 0) {
    const parsed = recordsSchema.safeParse(await LogBinding.Records({ listener, limit }))
    if (!parsed.success) {
      throw new Error(`log records rejected: ${parsed.error.message}`)
    }
    return parsed.data
  },

  async clear(listener) {
    await LogBinding.Clear(listener)
  },

  async setFilter(params) {
    await LogBinding.SetFilter({ levels: params.levels, pattern: params.pattern })
  },
}
