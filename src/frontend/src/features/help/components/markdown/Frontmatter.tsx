import { Table, TableBody, TableCell, TableRow } from '@/components/ui/table'

// FrontmatterEntry is one key/value pair of the block at the top of a page.
export type FrontmatterEntry = { key: string; value: string }

type FrontmatterProps = {
  entries: readonly FrontmatterEntry[]
}

// Frontmatter renders the metadata block as a two column table without a header: the key on the left, the
// value on the right. A page without metadata renders nothing.
export function Frontmatter({ entries }: FrontmatterProps) {
  if (entries.length === 0) {
    return null
  }

  return (
    <Table className="my-3 text-xs">
      <TableBody>
        {entries.map((entry) => (
          <TableRow key={entry.key} className="hover:bg-transparent">
            <TableCell className="w-px whitespace-nowrap py-1 pr-4 pl-0 font-medium text-muted-foreground">
              {entry.key}
            </TableCell>
            <TableCell className="py-1 pr-0 break-words">{entry.value}</TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}
