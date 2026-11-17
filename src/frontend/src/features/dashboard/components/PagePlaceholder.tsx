// PagePlaceholder reserves the page body until the real feature content lands. The page title lives in
// the shared site header, so this only states the feature is pending.
export function PagePlaceholder({ title }: { title: string }) {
  return (
    <div className="flex flex-1 items-start p-6">
      <p className="text-sm text-muted-foreground">{title} is not implemented yet.</p>
    </div>
  )
}
