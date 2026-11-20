import DOMPurify from 'dompurify'
import { load } from 'js-yaml'
import { Marked } from 'marked'

// The reader renders markdown, so it needs GitHub flavoured tables, task lists and strikethrough. Parsing,
// sanitising and splitting live here so components only decide how to paint a segment.
const marked = new Marked({ gfm: true, breaks: false })

// External links open in a new tab so the reader is never navigated away from the dialog.
DOMPurify.addHook('afterSanitizeAttributes', (node) => {
  if (node.tagName === 'A') {
    node.setAttribute('target', '_blank')
    node.setAttribute('rel', 'noreferrer noopener')
  }
})

// Segment is one piece of the rendered document. A fenced block is pulled out because it renders as a
// component; everything else stays as sanitised html.
export type Segment =
  | { kind: 'html'; html: string }
  | { kind: 'code'; language: string; code: string }

// prose holds the typography the reader applies. Tailwind has no prose plugin here, so the rules are
// written against the generated elements instead of adding a dependency. Code fences are excluded: they are
// rendered by MarkdownCodeBlock.
export const prose = [
  'min-w-0 max-w-full break-words text-sm leading-relaxed text-foreground',
  '[&_h1]:mt-0 [&_h1]:mb-3 [&_h1]:text-xl [&_h1]:font-semibold',
  '[&_h2]:mt-6 [&_h2]:mb-2 [&_h2]:text-lg [&_h2]:font-semibold',
  '[&_h3]:mt-5 [&_h3]:mb-2 [&_h3]:text-base [&_h3]:font-semibold',
  '[&_h4]:mt-4 [&_h4]:mb-1 [&_h4]:text-sm [&_h4]:font-semibold',
  '[&_p]:my-3 [&_p]:break-words',
  '[&_a]:break-all [&_a]:text-primary [&_a]:underline [&_a]:underline-offset-2',
  '[&_ul]:my-3 [&_ul]:list-disc [&_ul]:pl-5 [&_ol]:my-3 [&_ol]:list-decimal [&_ol]:pl-5',
  '[&_li]:my-1 [&_li]:break-words [&_li]:marker:text-muted-foreground',
  '[&_blockquote]:my-3 [&_blockquote]:border-l-2 [&_blockquote]:border-border [&_blockquote]:pl-3 [&_blockquote]:text-muted-foreground',
  '[&_code]:rounded [&_code]:bg-muted [&_code]:px-1 [&_code]:py-0.5 [&_code]:text-[0.85em]',
  '[&_table]:my-3 [&_table]:w-full [&_table]:border-collapse [&_table]:text-xs',
  '[&_th]:border [&_th]:border-border [&_th]:bg-muted [&_th]:px-2 [&_th]:py-1 [&_th]:text-left [&_th]:font-medium [&_th]:break-words',
  '[&_td]:border [&_td]:border-border [&_td]:px-2 [&_td]:py-1 [&_td]:break-words',
  '[&_hr]:my-5 [&_hr]:border-border',
  '[&_kbd]:rounded [&_kbd]:border [&_kbd]:border-border [&_kbd]:bg-muted [&_kbd]:px-1 [&_kbd]:py-0.5 [&_kbd]:text-[0.8em]',
  '[&_input[type=checkbox]]:mr-1 [&_input[type=checkbox]]:align-middle',
].join(' ')

// FrontmatterEntry is one row of the metadata block shown above a document.
export type FrontmatterEntry = { key: string; value: string }

// readFrontmatter splits the optional block at the top of a page from its body. Values are flattened to text
// because the block is rendered as a key/value table, not as data.
//
// The fence is located by hand and only the block is handed to the YAML parser: pulling in a wrapper that
// accepts the whole document would drag a Node Buffer shim into the browser bundle.
export function readFrontmatter(source: string): { entries: FrontmatterEntry[]; body: string } {
  const block = splitFence(source)
  if (!block) {
    return { entries: [], body: source }
  }

  let data: unknown
  try {
    data = load(block.raw)
  } catch {
    return { entries: [], body: block.body }
  }
  if (typeof data !== 'object' || data === null) {
    return { entries: [], body: block.body }
  }

  const entries: FrontmatterEntry[] = []
  for (const [key, value] of Object.entries(data as Record<string, unknown>)) {
    const text = Array.isArray(value) ? value.join(', ') : value === null || value === undefined ? '' : String(value)
    entries.push({ key, value: text.trim() })
  }
  return { entries, body: block.body }
}

// splitFence finds the leading "---" block and returns its YAML plus the remaining body. A page without a
// closed fence is treated as plain markdown.
function splitFence(source: string): { raw: string; body: string } | null {
  if (!source.startsWith('---')) {
    return null
  }
  const close = source.indexOf('\n---', 3)
  if (close === -1) {
    return null
  }
  const raw = source.slice(3, close)
  const next = source.indexOf('\n', close + 1)
  return { raw, body: next === -1 ? '' : source.slice(next + 1) }
}

// renderMarkdown turns markdown into html. Parsing is synchronous, so the result is a string ready to be
// sanitised.
export function renderMarkdown(source: string): string {
  return marked.parse(source, { async: false })
}

// sanitizeHtml removes everything the reader does not need. Only the tags markdown emits pass through, plus
// the attributes added for external links.
export function sanitizeHtml(html: string): string {
  return DOMPurify.sanitize(html, {
    ADD_ATTR: ['target', 'rel'],
    FORBID_TAGS: ['script', 'style', 'iframe', 'object', 'embed', 'form'],
  })
}

// splitBlocks separates the document into html runs and fenced blocks. A fence is identified by its <pre>
// element, and its language comes from the class marked writes on the inner <code>.
export function splitBlocks(html: string): Segment[] {
  const template = document.createElement('template')
  template.innerHTML = html

  const segments: Segment[] = []
  let buffer = ''

  const flush = () => {
    if (buffer.trim() !== '') {
      segments.push({ kind: 'html', html: buffer })
    }
    buffer = ''
  }

  for (const node of Array.from(template.content.childNodes)) {
    if (node instanceof HTMLElement && node.tagName === 'PRE') {
      flush()
      const code = node.querySelector('code')
      const language = code?.className.match(/language-([\w-]+)/)?.[1] ?? ''
      segments.push({ kind: 'code', language, code: code?.textContent ?? '' })
      continue
    }
    if (node instanceof HTMLElement) {
      buffer += node.outerHTML
      continue
    }
    buffer += node.textContent ?? ''
  }
  flush()

  return segments
}
