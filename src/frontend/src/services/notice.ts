import { z } from 'zod'
import { NoticeService as NoticeBinding } from '../../bindings/github.com/wensboy/lightpc/backend/service/notice'

// levelSchema mirrors the four severities the engine accepts. The interface draws one icon and one tone per value, so
// an unknown level would have nothing to render with.
const levelSchema = z.enum(['info', 'success', 'warning', 'error'])

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
const historySchema = z.object({
  items: z.array(itemSchema).default([]),
  limit: z.number().default(0),
  // The id the window starts at, zero when empty. It is what tells a capped answer from a complete one.
  leftEdge: z.number().default(0),
})

/** NoticeLevelDto is the severity a notice carries. */
export type NoticeLevelDto = z.infer<typeof levelSchema>

/** NoticeItemDto is one notice. */
export type NoticeItemDto = z.infer<typeof itemSchema>

/** NoticeHistoryDto is the answer of one read. */
export type NoticeHistoryDto = z.infer<typeof historySchema>

/** NoticeSendInputDto is what the interface submits to raise a notice. */
export type NoticeSendInputDto = {
  level: NoticeLevelDto
  title: string
  content?: string
  source?: string
}

/** NoticeService is the contract the notifications feature depends on. */
export interface NoticeService {
  /** history returns the live notices, newest first. */
  history(): Promise<NoticeHistoryDto>
  /** send raises a notice through the engine, so a backend toast and a front end one travel the same path. */
  send(input: NoticeSendInputDto): Promise<NoticeItemDto>
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
        level: input.level,
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

  async remove(ids) {
    return Number(await NoticeBinding.Remove({ ids }))
  },
}
