import { z } from 'zod'
import { Service as HelpBinding } from '../../bindings/github.com/wensboy/lightpc/backend/service/help'

// docsItemSchema validates one document entry of the table of contents.
const docsItemSchema = z.object({
  title: z.string(),
  path: z.string(),
})

// docsGroupSchema validates one section, keeping the order the backend preserved.
const docsGroupSchema = z.object({
  title: z.string(),
  items: z.array(docsItemSchema),
})

// docsVersionSchema validates one documentation set.
const docsVersionSchema = z.object({
  version: z.string(),
  description: z.string(),
  author: z.string(),
  date: z.string(),
  deprecated: z.boolean(),
  tags: z.array(z.string()).default([]),
  toc: z.array(docsGroupSchema),
})

const docsPayloadSchema = z.object({
  versions: z.array(docsVersionSchema),
})

/** DocsItem is one document listed in a group. */
export type DocsItem = z.infer<typeof docsItemSchema>

/** DocsGroup is one section of a documentation set. */
export type DocsGroup = z.infer<typeof docsGroupSchema>

/** DocsVersion is one documentation set as the reader shows it. */
export type DocsVersion = z.infer<typeof docsVersionSchema>

/** HelpService is the contract the reader depends on. */
export interface HelpService {
  /** payload returns every documentation set the backend ships. */
  payload(): Promise<DocsVersion[]>
  /** read returns the markdown of one document. */
  read(version: string, path: string): Promise<string>
}

// helpService asks the backend: the documentation ships with the application, so there is nothing to mock.
export const helpService: HelpService = {
  async payload() {
    const payload = await HelpBinding.Payload()
    const parsed = docsPayloadSchema.safeParse(payload)
    if (!parsed.success) {
      throw new Error(`help payload rejected: ${parsed.error.message}`)
    }
    return parsed.data.versions
  },

  async read(version, path) {
    const content = await HelpBinding.Read(version, path)
    return typeof content === 'string' ? content : ''
  },
}
