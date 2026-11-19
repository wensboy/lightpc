import { z } from 'zod'
import { SettingService as SettingBinding } from '../../bindings/github.com/wensboy/lightpc/backend/service/setting'
import type { JsonSchemaNode } from '@/features/settings/schema'

// settingPayloadSchema guards the boundary: the renderer only ever sees a payload that passed this check,
// so a backend shape change fails loudly instead of rendering an empty page.
const settingPayloadSchema = z.object({
  schema: z.record(z.string(), z.unknown()),
  values: z.record(z.string(), z.unknown()),
  defaults: z.record(z.string(), z.unknown()),
})

// SettingDocument is what the renderer needs: the schema describing every field and the values to show.
export type SettingDocument = {
  schema: JsonSchemaNode
  values: Record<string, unknown>
  /** Authored values from the default layer, used when a user value breaks the schema. */
  defaults: Record<string, unknown>
}

// SettingService is the contract the feature depends on, independent of where the document comes from.
export interface SettingService {
  load(): Promise<SettingDocument>
  /** Persists one group of the document; the backend flushes it to the user layer in the background. */
  save(key: string, value: unknown): Promise<void>
}

// settingService asks the backend for the document, since the schema and the default document are
// embedded resources the file store already serves.
export const settingService: SettingService = {
  async load() {
    const payload = await SettingBinding.Payload()
    const parsed = settingPayloadSchema.safeParse(payload)
    if (!parsed.success) {
      throw new Error(`setting payload rejected: ${parsed.error.message}`)
    }
    return {
      schema: parsed.data.schema as JsonSchemaNode,
      values: parsed.data.values,
      defaults: parsed.data.defaults,
    }
  },

  async save(key, value) {
    await SettingBinding.Save(key, value)
  },
}
