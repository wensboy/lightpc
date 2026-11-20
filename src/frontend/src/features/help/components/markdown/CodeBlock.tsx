import { WrapTextIcon } from 'lucide-react'
import { useState } from 'react'
import { Button } from '@/components/ui/button'
import {
  CodeBlock,
  CodeBlockActions,
  CodeBlockContent,
  CodeBlockCopyButton,
  CodeBlockHeader,
  CodeBlockLine,
  CodeBlockTitle,
} from '@/components/ui/code-block'
import { useHighlightedCode } from '@/hooks/use-highlighted-code'

type MarkdownCodeBlockProps = {
  /** language is the fence info string; an empty value means the block declared none. */
  language: string
  /** code is the verbatim fence body. */
  code: string
}

// MarkdownCodeBlock renders a fenced block. The fence language becomes the header title, the header offers
// wrapping and copying, and the body keeps its own horizontal scroll unless wrapping is on.
export function MarkdownCodeBlock({ language, code }: MarkdownCodeBlockProps) {
  const [wrap, setWrap] = useState(false)
  const lines = useHighlightedCode(code, language || 'text')

  return (
    <CodeBlock className="my-3">
      <CodeBlockHeader>
        <CodeBlockTitle>{language || 'text'}</CodeBlockTitle>
        <CodeBlockActions>
          <Button
            variant="ghost"
            size="icon-xs"
            aria-label="Wrap lines"
            aria-pressed={wrap}
            className="aria-pressed:text-foreground"
            onClick={() => setWrap(!wrap)}
          >
            <WrapTextIcon />
          </Button>
          <CodeBlockCopyButton value={code} />
        </CodeBlockActions>
      </CodeBlockHeader>
      <CodeBlockContent wrap={wrap}>
        {lines.map((line, index) => (
          <CodeBlockLine key={index}>{line}</CodeBlockLine>
        ))}
      </CodeBlockContent>
    </CodeBlock>
  )
}
