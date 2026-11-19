import { MinusIcon, PlusIcon } from 'lucide-react'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
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
import { SearchCombobox } from '@/features/help/components/SearchCombobox'
import { toSections } from '@/features/help/sections'
import { useText } from '@/locales'
import { useDialogStore } from '@/stores/dialog'

// HelpDialog combines the sidebar-02 header (documentation switcher plus search) with the sidebar-05
// collapsible tree. The selected document lives in the shared store, so the combobox label, the panel
// body and the tree highlight always show the same page.
export function HelpDialog() {
  const text = useText()
  const helpOpen = useDialogStore((state) => state.helpOpen)
  const closeHelp = useDialogStore((state) => state.closeHelp)
  const helpDocument = useDialogStore((state) => state.helpDocument)
  const selectHelpDocument = useDialogStore((state) => state.selectHelpDocument)

  const sections = toSections(text)
  const currentSection =
    sections.find((section) => section.items.includes(helpDocument))?.title ?? sections[0].title

  return (
    <Dialog
      open={helpOpen}
      onOpenChange={(open) => {
        if (!open) {
          closeHelp()
        }
      }}
    >
      <DialogContent className="overflow-hidden p-0 md:h-[500px] md:max-w-[700px] lg:max-w-[800px]">
        <DialogTitle className="sr-only">{text.help.title}</DialogTitle>
        <DialogDescription className="sr-only">{text.help.description}</DialogDescription>
        <SidebarProvider className="h-full min-h-0 items-stretch">
          <Sidebar
            collapsible="none"
            className="hidden border-r border-sidebar-border md:flex"
          >
            <SidebarHeader>
              <DocumentationSwitcher versions={text.help.switcher.versions} />
              <SidebarGroup className="py-0">
                <SidebarGroupContent>
                  <SearchCombobox
                    value={helpDocument}
                    documents={sections.flatMap((section) => section.items)}
                    onSelect={selectHelpDocument}
                  />
                </SidebarGroupContent>
              </SidebarGroup>
            </SidebarHeader>
            <SidebarContent>
              <SidebarGroup>
                <SidebarMenu className="gap-1">
                  {sections.map((section, index) => (
                    <Collapsible
                      key={section.title}
                      defaultOpen={index === 0}
                      className="group/collapsible"
                    >
                      <SidebarMenuItem>
                        <SidebarMenuButton render={<CollapsibleTrigger />}>
                          {section.title}{' '}
                          <PlusIcon className="ml-auto group-aria-expanded/menu-button:hidden" />
                          <MinusIcon className="ml-auto hidden group-aria-expanded/menu-button:block" />
                        </SidebarMenuButton>
                        <CollapsibleContent>
                          <SidebarMenuSub>
                            {section.items.map((item) => (
                              <SidebarMenuSubItem key={item}>
                                <SidebarMenuSubButton
                                  isActive={item === helpDocument}
                                  onClick={() => selectHelpDocument(item)}
                                >
                                  {item}
                                </SidebarMenuSubButton>
                              </SidebarMenuSubItem>
                            ))}
                          </SidebarMenuSub>
                        </CollapsibleContent>
                      </SidebarMenuItem>
                    </Collapsible>
                  ))}
                </SidebarMenu>
              </SidebarGroup>
            </SidebarContent>
          </Sidebar>
          <main className="flex h-full flex-1 flex-col overflow-hidden">
            <header className="flex h-16 shrink-0 items-center gap-2 px-4">
              <h2 className="text-sm font-medium">{helpDocument || currentSection}</h2>
            </header>
            <div className="flex flex-1 flex-col gap-4 overflow-y-auto p-4 pt-2">
              <p className="text-sm text-muted-foreground">
                {text.help.placeholder.pending(currentSection)}
              </p>
            </div>
          </main>
        </SidebarProvider>
      </DialogContent>
    </Dialog>
  )
}
