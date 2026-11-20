import { useMemo } from 'react'
import { MarkdownCodeBlock } from '@/features/help/components/markdown/CodeBlock'
import { prose, renderMarkdown, sanitizeHtml, splitBlocks } from '@/features/help/render'

// HelpMarkdown renders one document. Parsing, sanitising and splitting live in render so this component only
// maps a segment to the component or container that paints it.
export function HelpMarkdown({ source }: { source: string }) {
  const segments = useMemo(() => splitBlocks(sanitizeHtml(renderMarkdown(source))), [source])

  return (
    <div className="flex min-w-0 flex-col gap-3">
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
