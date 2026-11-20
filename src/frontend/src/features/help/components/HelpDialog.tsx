import { ChevronRightIcon, MinusIcon, PlusIcon } from 'lucide-react'
import { useMemo, useState } from 'react'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'
import { Dialog, DialogContent, DialogDescription, DialogTitle } from '@/components/ui/dialog'
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
  SidebarProvider,
} from '@/components/ui/sidebar'
import { DocumentationSwitcher } from '@/features/help/components/DocumentationSwitcher'
import { HelpMarkdown } from '@/features/help/components/HelpMarkdown'
import { SearchCombobox, type SearchSection } from '@/features/help/components/SearchCombobox'
import { useHelp } from '@/features/help/useHelp'
import { useText } from '@/locales'
import type { DocsItem, DocsVersion } from '@/services/help'
import { useDialogStore } from '@/stores/dialog'

// documentValue is what the picker shows for a document: its path without the markdown suffix, so
// "quick-guide/usage.md" reads as "quick-guide/usage".
function documentValue(path: string): string {
  return path.replace(/\.md$/i, '')
}

// HelpDialog is the documentation browser. The version picker, the tree and the reader all read from one
// hook, so switching a version clears the article and picking a document moves the breadcrumb with it.
export function HelpDialog() {
  const text = useText()
  const helpOpen = useDialogStore((state) => state.helpOpen)
  const closeHelp = useDialogStore((state) => state.closeHelp)
  const { versions, version, content, error, selectVersion, selectDocument } = useHelp(helpOpen)

  const [active, setActive] = useState<{ group: string; item: DocsItem } | null>(null)

  const current = useMemo(
    () => versions.find((entry) => entry.version === version) ?? null,
    [versions, version],
  )
  const groups = current?.toc ?? []

  // Each section of the table of contents becomes a group of the picker, so the first level stays visible
  // while the second level is what gets searched.
  const sections = useMemo<SearchSection[]>(
    () => groups.map((group) => ({ title: group.title, items: group.items.map((item) => item.path) })),
    [groups],
  )

  // labels lets the picker show a title while the primitive filters on the path.
  const labels = useMemo<Record<string, string>>(() => {
    const table: Record<string, string> = {}
    for (const group of groups) {
      for (const item of group.items) {
        table[item.path] = item.title
      }
    }
    return table
  }, [groups])

  const open = (group: string, item: DocsItem) => {
    setActive({ group, item })
    selectDocument(item.path)
  }

  const openByPath = (path: string) => {
    for (const group of groups) {
      for (const item of group.items) {
        if (item.path === path) {
          open(group.title, item)
          return
        }
      }
    }
  }

  return (
    <Dialog
      open={helpOpen}
      onOpenChange={(next) => {
        if (!next) {
          closeHelp()
        }
      }}
    >
      <DialogContent className="grid-rows-[1fr] overflow-hidden p-0 md:h-[500px] md:max-h-[500px] md:max-w-[700px] lg:max-w-[800px]">
        <DialogTitle className="sr-only">{text.help.title}</DialogTitle>
        <DialogDescription className="sr-only">{text.help.description}</DialogDescription>
        <SidebarProvider className="h-full min-h-0 w-full min-w-0 items-stretch">
          <Sidebar collapsible="none" className="hidden border-r border-sidebar-border md:flex">
            <SidebarHeader>
              <DocumentationSwitcher
                versions={versions}
                active={version}
                onChange={(next) => {
                  setActive(null)
                  selectVersion(next)
                }}
              />
              <SidebarGroup className="py-0">
                <SearchCombobox
                  value={active ? documentValue(active.item.path) : ''}
                  sections={sections}
                  labels={labels}
                  onSelect={openByPath}
                />
              </SidebarGroup>
            </SidebarHeader>
            <SidebarContent>
              <SidebarGroup>
                <SidebarGroupContent>
                  <SidebarMenu className="gap-1">
                    {groups.map((group, index) => (
                      <Collapsible
                        key={group.title}
                        defaultOpen={index === 0}
                        className="group/collapsible"
                      >
                        <SidebarMenuItem>
                          <SidebarMenuButton
                            tooltip={group.title}
                            render={<CollapsibleTrigger />}
                          >
                            <span>{group.title}</span>
                            <PlusIcon className="ml-auto size-4 group-aria-expanded/menu-button:hidden" />
                            <MinusIcon className="ml-auto hidden size-4 group-aria-expanded/menu-button:block" />
                          </SidebarMenuButton>
                          <CollapsibleContent>
                            <SidebarMenuSub>
                              {group.items.map((item) => (
                                <SidebarMenuSubItem key={item.path}>
                                  <SidebarMenuSubButton
                                    isActive={active?.item.path === item.path}
                                    onClick={() => open(group.title, item)}
                                  >
                                    <span className="truncate">{item.title}</span>
                                  </SidebarMenuSubButton>
                                </SidebarMenuSubItem>
                              ))}
                            </SidebarMenuSub>
                          </CollapsibleContent>
                        </SidebarMenuItem>
                      </Collapsible>
                    ))}
                  </SidebarMenu>
                </SidebarGroupContent>
              </SidebarGroup>
            </SidebarContent>
          </Sidebar>
          <main className="flex h-full min-w-0 flex-1 flex-col overflow-hidden">
            <header className="flex h-12 shrink-0 items-center gap-2 border-b border-border/60 px-4">
              {active ? (
                <Breadcrumb>
                  <BreadcrumbList>
                    <BreadcrumbItem className="text-muted-foreground">{active.group}</BreadcrumbItem>
                    <BreadcrumbSeparator>
                      <ChevronRightIcon />
                    </BreadcrumbSeparator>
                    <BreadcrumbItem>
                      <BreadcrumbPage>{active.item.title}</BreadcrumbPage>
                    </BreadcrumbItem>
                  </BreadcrumbList>
                </Breadcrumb>
              ) : (
                <span className="truncate text-sm text-muted-foreground">{version || text.help.title}</span>
              )}
            </header>
            <div className="flex min-h-0 min-w-0 flex-1 flex-col gap-3 overflow-y-auto overflow-x-hidden p-4 pt-2">
              {error ? (
                <p className="text-xs text-destructive">{error}</p>
              ) : content ? (
                <HelpMarkdown source={content} />
              ) : current ? (
                <VersionSummary version={current} />
              ) : (
                <p className="text-xs text-muted-foreground">{text.help.reader.empty}</p>
              )}
            </div>
          </main>
        </SidebarProvider>
      </DialogContent>
    </Dialog>
  )
}

// VersionSummary is what the reader shows before a document is opened: the metadata of the active set,
// so the panel is not blank while the user chooses.
function VersionSummary({ version }: { version: DocsVersion }) {
  const text = useText()

  return (
    <div className="flex flex-col gap-3">
      <p className="text-sm leading-relaxed text-muted-foreground">{version.description}</p>
      <dl className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-1 text-xs">
        <dt className="text-muted-foreground">{text.help.reader.author}</dt>
        <dd>{version.author}</dd>
        <dt className="text-muted-foreground">{text.help.reader.date}</dt>
        <dd>{version.date}</dd>
      </dl>
      <div className="flex flex-wrap gap-1">
        {version.tags?.map((tag) => (
          <span key={tag} className="rounded bg-muted px-1.5 py-0.5 text-[0.7rem] text-muted-foreground">
            {tag}
          </span>
        ))}
      </div>
    </div>
  )
}
