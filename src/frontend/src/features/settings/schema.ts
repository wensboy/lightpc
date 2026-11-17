import { z } from 'zod'

// A select option carries a stable value and a dictionary key for its label, so the text never lives
// in the data.
const settingOptionSchema = z.object({
  value: z.string().min(1),
  labelKey: z.string().min(1),
})

// Fields shared by every control type. labelKey and descriptionKey point into the text dictionary.
const baseFields = {
  id: z.string().min(1),
  labelKey: z.string().min(1),
  descriptionKey: z.string().min(1).optional(),
}

export const settingSchema = z.discriminatedUnion('type', [
  z.object({
    ...baseFields,
    type: z.literal('switch'),
    defaultValue: z.boolean(),
  }),
  z.object({
    ...baseFields,
    type: z.literal('input'),
    defaultValue: z.string(),
    placeholder: z.string().optional(),
  }),
  z.object({
    ...baseFields,
    type: z.literal('select'),
    defaultValue: z.string(),
    options: z.array(settingOptionSchema).min(1),
  }),
])

export type SettingSchema = z.infer<typeof settingSchema>
export type SettingOption = z.infer<typeof settingOptionSchema>

// A category groups settings the way the VSCode settings page does.
export const settingCategorySchema = z.object({
  id: z.string().min(1),
  labelKey: z.string().min(1),
  settings: z.array(settingSchema).min(1),
})

export type SettingCategory = z.infer<typeof settingCategorySchema>

// The document is the whole settings tree served to the page.
export const settingDocumentSchema = z.object({
  categories: z.array(settingCategorySchema).min(1),
})

export type SettingDocument = z.infer<typeof settingDocumentSchema>

// SettingValue is the runtime payload of a setting; the control type decides which variant it is.
export type SettingValue = boolean | string
