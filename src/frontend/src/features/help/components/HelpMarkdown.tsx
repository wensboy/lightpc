import { useMemo } from 'react'
import { Frontmatter } from '@/features/help/components/markdown/Frontmatter'
import { MarkdownCodeBlock } from '@/features/help/components/markdown/CodeBlock'
import { prose, readFrontmatter, renderMarkdown, sanitizeHtml, splitBlocks } from '@/features/help/render'

// HelpMarkdown renders one document. The frontmatter block becomes the metadata table above the body, and
// the rest is split so fenced blocks render as components.
export function HelpMarkdown({ source }: { source: string }) {
  const { entries, body } = useMemo(() => readFrontmatter(source), [source])
  const segments = useMemo(() => splitBlocks(sanitizeHtml(renderMarkdown(body))), [body])

  return (
    <div className="flex min-w-0 flex-col gap-3">
      <Frontmatter entries={entries} />
      {segments.map((segment, index) =>
        segment.kind === 'code' ? (
          <MarkdownCodeBlock key={index} language={segment.language} code={segment.code} />
        ) : (
          <div key={index} className={prose} dangerouslySetInnerHTML={{ __html: segment.html }} />
        ),
      )}
    </div>
  )
}
