// JsonSchemaNode is the subset of JSON Schema the settings renderer understands.
export type JsonSchemaNode = {
  $ref?: string
  type?: string
  title?: string
  description?: string
  enum?: string[]
  pattern?: string
  minLength?: number
  maxLength?: number
  minimum?: number
  maximum?: number
  properties?: Record<string, JsonSchemaNode>
  items?: JsonSchemaNode
  $defs?: Record<string, JsonSchemaNode>
  /** Derived node: it backs other fields and is never rendered. */
  'x-drived'?: boolean
  /** Tells the renderer where a select takes its options from. */
  'x-options'?: { from: string; field?: string }
  /** Swaps the option source while a sibling boolean is enabled. */
  'x-options-when'?: { key: string; from: string }
}

// SettingControl is the control a leaf field renders as.
export type SettingControl = 'switch' | 'select' | 'input'

// SettingValue is the runtime payload a control can bind to.
export type SettingValue = boolean | string | number

// SettingField is one renderable setting: the schema told us how to draw it and the document holds the
// value to show.
export type SettingField = {
  path: string
  /** Segment path below the group, rendered with chevrons between the parts. */
  title: string[]
  description?: string
  control: SettingControl
  value: SettingValue
  options?: string[]
  pattern?: string
  /** True when the stored value violated the schema and the shown value is a fallback. */
  invalid: boolean
}

// SettingGroup is one top level section of the document, rendered as a page on the left rail.
export type SettingGroup = {
  key: string
  title: string
  description?: string
  fields: SettingField[]
}

// resolveRef follows a local $ref such as "#/$defs/size" and keeps the description of the use site.
function resolveRef(node: JsonSchemaNode, root: JsonSchemaNode): JsonSchemaNode {
  if (!node.$ref) {
    return node
  }
  const name = node.$ref.split('/').pop()
  const target = name ? root.$defs?.[name] : undefined
  if (!target) {
    return node
  }
  return { ...target, description: node.description ?? target.description }
}

// optionsOf resolves the options of a select: either the enum declared by the schema or the names pulled
// from a sibling array through the x-options extension.
function optionsOf(
  node: JsonSchemaNode,
  siblings: Record<string, unknown>,
  systemFonts: string[],
): string[] | undefined {
  if (node.enum && node.enum.length > 0) {
    return node.enum
  }
  // A conditional source wins while its sibling boolean is on, so one schema entry covers both modes.
  const condition = node['x-options-when']
  if (condition && siblings[condition.key] === true) {
    if (condition.from === 'systemFonts' && systemFonts.length > 0) {
      return systemFonts
    }
    return undefined
  }
  const source = node['x-options']
  if (!source) {
    return undefined
  }
  const list = siblings[source.from]
  if (!Array.isArray(list)) {
    return undefined
  }
  const options = list
    .map((entry) => {
      if (typeof entry === 'string') {
        return entry
      }
      if (!source.field || typeof entry !== 'object' || entry === null) {
        return undefined
      }
      const label = (entry as Record<string, unknown>)[source.field]
      return typeof label === 'string' ? label : undefined
    })
    .filter((value): value is string => typeof value === 'string' && value !== '')
  return options.length > 0 ? options : undefined
}

// controlOf picks the control for a leaf node: options become a select, a boolean a switch, anything else
// an input.
function controlOf(node: JsonSchemaNode, options: string[] | undefined): SettingControl {
  if (options && options.length > 0) {
    return 'select'
  }
  if (node.type === 'boolean') {
    return 'switch'
  }
  return 'input'
}

// titleOf drops the group segment and returns the rest as segments, so appearence.theme.default renders
// as "theme" chevron "default" while a direct child keeps its own name.
function titleOf(segments: string[]): string[] {
  return segments.length > 1 ? segments.slice(1) : segments
}

// isValidValue reports whether value satisfies the constraints the schema declares for a leaf. The
// default layer is authored against the schema, so a violation can only come from a hand edited user file.
function isValidValue(
  node: JsonSchemaNode,
  value: SettingValue,
  options: string[] | undefined,
): boolean {
  // Options come from an enum or from the x-options extension; both constrain the accepted values.
  if (options && options.length > 0) {
    return typeof value === 'string' && options.includes(value)
  }
  if (node.type === 'boolean') {
    return typeof value === 'boolean'
  }
  if (node.type === 'number' || node.type === 'integer') {
    return typeof value === 'number' && Number.isFinite(value)
  }
  if (node.type === 'string') {
    if (typeof value !== 'string') {
      return false
    }
    if (node.minLength !== undefined && value.length < node.minLength) {
      return false
    }
    if (node.maxLength !== undefined && value.length > node.maxLength) {
      return false
    }
    if (node.pattern) {
      try {
        return new RegExp(node.pattern).test(value)
      } catch {
        return true
      }
    }
    return true
  }
  return true
}

// asValue narrows a stored value into something a control can bind to.
function asValue(value: unknown, control: SettingControl): SettingValue {
  if (control === 'switch') {
    return value === true
  }
  if (typeof value === 'number') {
    return value
  }
  return typeof value === 'string' ? value : ''
}

// atPath reads a dotted path out of a nested document.
function atPath(document: Record<string, unknown>, path: string): unknown {
  let cursor: unknown = document
  for (const segment of path.split('.')) {
    if (typeof cursor !== 'object' || cursor === null) {
      return undefined
    }
    cursor = (cursor as Record<string, unknown>)[segment]
  }
  return cursor
}

// isSettingValue narrows unknown into a value a control can bind to.
function isSettingValue(value: unknown): value is SettingValue {
  return typeof value === 'string' || typeof value === 'number' || typeof value === 'boolean'
}

// buildGroups walks the document alongside the schema. Nodes that declare properties are treated as
// nested objects and flattened, while scalars become fields bound to their current value. A user value
// that violates the schema falls back to the authored default, because a zero value may be invalid too.
export function buildGroups(
  schema: JsonSchemaNode,
  values: Record<string, unknown>,
  defaults: Record<string, unknown> = {},
  edited: ReadonlySet<string> = new Set(),
  systemFonts: string[] = [],
): SettingGroup[] {
  const groups: SettingGroup[] = []

  for (const [key, rawNode] of Object.entries(schema.properties ?? {})) {
    const node = resolveRef(rawNode, schema)
    const value = values[key]
    if (!node.properties || typeof value !== 'object' || value === null || Array.isArray(value)) {
      continue
    }
    const fields: SettingField[] = []
    collectFields(node, schema, defaults, edited, systemFonts, [key], value as Record<string, unknown>, fields)
    if (fields.length === 0) {
      continue
    }
    // Sorted by the whole path so siblings of the same branch stay adjacent and the order is stable.
    fields.sort((a, b) => a.title.join('.').localeCompare(b.title.join('.')))
    groups.push({
      key,
      title: node.title ?? key,
      description: node.description,
      fields,
    })
  }
  return groups
}

// collectFields descends one node and appends a field for every scalar it reaches.
function collectFields(
  node: JsonSchemaNode,
  root: JsonSchemaNode,
  defaults: Record<string, unknown>,
  edited: ReadonlySet<string>,
  systemFonts: string[],
  segments: string[],
  values: Record<string, unknown>,
  out: SettingField[],
) {
  for (const [key, rawChild] of Object.entries(node.properties ?? {})) {
    const child = resolveRef(rawChild, root)
    const value = values[key]
    const path = [...segments, key]

    if (child.properties && typeof value === 'object' && value !== null && !Array.isArray(value)) {
      collectFields(child, root, defaults, edited, systemFonts, path, value as Record<string, unknown>, out)
      continue
    }
    // Arrays (palettes, font lists) are edited by dedicated components later on.
    if (value === undefined || Array.isArray(value)) {
      continue
    }

    // A derived node backs other fields; rendering it would duplicate information.
    if (child['x-drived'] === true) {
      continue
    }
    if (Array.isArray(value)) {
      continue
    }

    const options = optionsOf(child, values, systemFonts)
    const control = controlOf(child, options)
    const dotted = path.join('.')
    const stored = asValue(value, control)
    const authored = atPath(defaults, dotted)
    const invalid = value !== undefined && !isValidValue(child, stored, options)
    // A value that breaks the schema is replaced by the authored default as soon as it is seen, so the
    // view never shows an impossible state. Fields the user is editing are left alone: a half typed value
    // is not a stored value and must stay under the cursor.
    const accepted =
      invalid && !edited.has(dotted) && isSettingValue(authored) ? authored : stored
    out.push({
      path: dotted,
      title: titleOf(path),
      description: child.description,
      control,
      value: accepted,
      options,
      pattern: child.pattern,
      invalid,
    })
  }
}

// setAt returns a copy of values with the dotted path replaced, leaving sibling branches untouched.
export function setAt(
  values: Record<string, unknown>,
  path: string,
  next: SettingValue,
): Record<string, unknown> {
  const segments = path.split('.')
  const copy: Record<string, unknown> = { ...values }
  let cursor = copy
  for (const segment of segments.slice(0, -1)) {
    const current = cursor[segment]
    const child = typeof current === 'object' && current !== null ? { ...(current as object) } : {}
    cursor[segment] = child
    cursor = child as Record<string, unknown>
  }
  cursor[segments[segments.length - 1]] = next
  return copy
}
