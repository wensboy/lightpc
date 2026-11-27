import { z } from 'zod'
import { SettingService as SettingBinding } from '../../bindings/github.com/wensboy/lightpc/backend/service/setting'
import { Kind } from '../../bindings/github.com/wensboy/lightpc/backend/domain/service/setting/models'
import type { JsonSchemaNode } from '@/features/settings/schema'

// settingPayloadSchema guards the boundary: the renderer only ever sees a payload that passed this check,
// so a backend shape change fails loudly instead of rendering an empty page.
const settingThemeSchema = z.object({
  name: z.string(),
  light: z.record(z.string(), z.string()),
  dark: z.record(z.string(), z.string()),
})

const settingPayloadSchema = z.object({
  schema: z.record(z.string(), z.unknown()),
  values: z.record(z.string(), z.unknown()),
  defaults: z.record(z.string(), z.unknown()),
  systemFonts: z.array(z.string()),
  themes: z.array(settingThemeSchema).default([]),
})

/** SettingThemeDto is one palette with both modes already keyed as CSS custom properties. */
export type SettingThemeDto = z.infer<typeof settingThemeSchema>

// SettingDocument is what the renderer needs: the schema describing every field and the values to show.
export type SettingDocument = {
  schema: JsonSchemaNode
  values: Record<string, unknown>
  /** Authored values from the default layer, used when a user value breaks the schema. */
  defaults: Record<string, unknown>
  /** Families installed on the machine, offered while font.useSystem is on. */
  systemFonts: string[]
  /** Selectable palettes, merged on the backend and keyed as custom properties. */
  themes: SettingThemeDto[]
}

// SettingService is the contract the feature depends on, independent of where the document comes from.
export interface SettingService {
  load(): Promise<SettingDocument>
  /** Persists one group of the document; the backend flushes it to the user layer in the background. */
  save(key: string, value: unknown, kind?: string): Promise<void>
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
      systemFonts: parsed.data.systemFonts,
      themes: parsed.data.themes,
    }
  },

  async save(key, value, kind) {
    // A field without x-kind is an ordinary setting, so the backend is told "service" rather than nothing: the routing
    // decision belongs there, and an empty string would only make it guess.
    //
    // The generator renders Kind as an enum, so the bare string has to be translated here - the same reason the notice
    // levels go through a table.
    await SettingBinding.Save(key, value, kind === 'engine' ? Kind.KindEngine : Kind.KindService)
  },
}
