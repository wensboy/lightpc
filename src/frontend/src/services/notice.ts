import { z } from 'zod'
import { NoticeService as NoticeBinding } from '../../bindings/github.com/wensboy/lightpc/backend/service/notice'
import { Level } from '../../bindings/github.com/wensboy/lightpc/backend/domain/service/notice/models'

// levelSchema mirrors the five severities the engine accepts. The interface draws one icon and one tone per value, so
// an unknown level would have nothing to render with.
const levelSchema = z.enum(['info', 'success', 'fail', 'error', 'warn'])

// itemSchema validates one notice the sheet shows.
const itemSchema = z.object({
  id: z.number(),
  level: levelSchema,
  title: z.string(),
  content: z.string().default(''),
  source: z.string().default(''),
  timestamp: z.number().default(0),
})

// historySchema validates one read. Limit travels with the list, so the sheet can show what it is capped at without
// asking a second question.
// summarySchema validates the counts the footer draws. They travel with the list because the backend already walked the
// rows: recounting here would only give the sheet a second opinion to disagree with.
const summarySchema = z.object({
  total: z.number().default(0),
  fail: z.number().default(0),
  error: z.number().default(0),
  warn: z.number().default(0),
})

const historySchema = z.object({
  items: z.array(itemSchema).default([]),
  limit: z.number().default(0),
  // The id the window starts at, zero when empty. It is what tells a capped answer from a complete one.
  leftEdge: z.number().default(0),
  summary: summarySchema.default({ total: 0, fail: 0, error: 0, warn: 0 }),
})

/** NoticeLevelDto is the severity a notice carries. */
export type NoticeLevelDto = z.infer<typeof levelSchema>

/** NoticeItemDto is one notice. */
export type NoticeItemDto = z.infer<typeof itemSchema>

/** NoticeSummaryDto counts the notices the answer carries, by severity. */
export type NoticeSummaryDto = z.infer<typeof summarySchema>

/** NoticeHistoryDto is the answer of one read. */
export type NoticeHistoryDto = z.infer<typeof historySchema>

/** NoticeSendInputDto is what the interface submits to raise a notice. */
export type NoticeSendInputDto = {
  level: NoticeLevelDto
  title: string
  content?: string
  source?: string
}

// levelMembers maps the string union this contract speaks onto the enum the binding declares.
//
// The generator turns a Go `type Level string` with constants into a TypeScript enum, whose members are not assignable
// from bare strings. Keeping the mapping here means a caller keeps working with plain values and only this file, the one
// place that already knows about the binding, deals with its shape.
const levelMembers: Record<NoticeLevelDto, Level> = {
  info: Level.LevelInfo,
  success: Level.LevelSuccess,
  fail: Level.LevelFail,
  error: Level.LevelError,
  warn: Level.LevelWarn,
}

/** NoticeService is the contract the notifications feature depends on. */
export interface NoticeService {
  /** history returns the live notices, newest first. */
  history(): Promise<NoticeHistoryDto>
  /** send raises a notice through the engine, so a backend toast and a front end one travel the same path. */
  send(input: NoticeSendInputDto): Promise<NoticeItemDto>
  /** clear hides everything currently shown, and anything older, and reports how many were covered. */
  clear(): Promise<number>
  /** remove drops the given notices and reports how many were there. */
  remove(ids: number[]): Promise<number>
}

// noticeService talks to the backend. The ring lives there, so both sides publish through it and the sheet always
// shows what the process actually recorded.
export const noticeService: NoticeService = {
  async history() {
    const parsed = historySchema.safeParse(await NoticeBinding.History())
    if (!parsed.success) {
      throw new Error(`notice history rejected: ${parsed.error.message}`)
    }
    return parsed.data
  },

  async send(input) {
    const parsed = itemSchema.safeParse(
      await NoticeBinding.Send({
        level: levelMembers[input.level],
        title: input.title,
        content: input.content ?? '',
        source: input.source ?? 'ui',
      }),
    )
    if (!parsed.success) {
      throw new Error(`notice send rejected: ${parsed.error.message}`)
    }
    return parsed.data
  },

  async clear() {
    return Number(await NoticeBinding.Clear())
  },

  async remove(ids) {
    return Number(await NoticeBinding.Remove({ ids }))
  },
}
