import { BlocksIcon, CodeIcon, SettingsIcon, SquareTerminalIcon } from 'lucide-react'
import { useState } from 'react'
import type { ReactNode } from 'react'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import { Dialog, DialogContent, DialogDescription, DialogTitle } from '@/components/ui/dialog'
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
} from '@/components/ui/sidebar'
import { SettingRenderer } from '@/features/settings/components/setting/SettingRenderer'
import { useSettings } from '@/features/settings/useSettings'
import { useText } from '@/locales'
import { resolveText } from '@/locales/text'
import { useDialogStore } from '@/stores/dialog'

// categoryIcons maps a category id to its glyph; an unknown id falls back to the generic settings icon.
const categoryIcons: Record<string, ReactNode> = {
  editor: <CodeIcon />,
  terminal: <SquareTerminalIcon />,
  extensions: <BlocksIcon />,
}

// SettingsDialog mirrors the sidebar-13 layout: categories on the left, the controls of the selected
// category on the right, rendered from the schema in a VSCode style list.
export function SettingsDialog() {
  const text = useText()
  const settingsOpen = useDialogStore((state) => state.settingsOpen)
  const closeSettings = useDialogStore((state) => state.closeSettings)
  const { document, values, setValue } = useSettings()
  const [activeId, setActiveId] = useState(document.categories[0].id)

  const activeCategory =
    document.categories.find((category) => category.id === activeId) ?? document.categories[0]
  const activeTitle = resolveText(text, activeCategory.labelKey)

  return (
    <Dialog
      open={settingsOpen}
      onOpenChange={(open) => {
        if (!open) {
          closeSettings()
        }
      }}
    >
      <DialogContent className="overflow-hidden p-0 md:max-h-[500px] md:max-w-[700px] lg:max-w-[800px]">
        <DialogTitle className="sr-only">{text.settings.title}</DialogTitle>
        <DialogDescription className="sr-only">{text.settings.description}</DialogDescription>
        <SidebarProvider className="items-start">
          <Sidebar
            collapsible="none"
            className="hidden h-[480px] border-r border-sidebar-border md:flex"
          >
            <SidebarContent>
              <SidebarGroup>
                <SidebarGroupContent>
                  <SidebarMenu className="gap-1">
                    {document.categories.map((category) => (
                      <SidebarMenuItem key={category.id}>
                        <SidebarMenuButton
                          isActive={category.id === activeCategory.id}
                          onClick={() => setActiveId(category.id)}
                        >
                          {categoryIcons[category.id] ?? <SettingsIcon />}
                          <span>{resolveText(text, category.labelKey)}</span>
                        </SidebarMenuButton>
                      </SidebarMenuItem>
                    ))}
                  </SidebarMenu>
                </SidebarGroupContent>
              </SidebarGroup>
            </SidebarContent>
          </Sidebar>
          <main className="flex h-[480px] flex-1 flex-col overflow-hidden">
            <header className="flex h-16 shrink-0 items-center gap-2">
              <div className="flex items-center gap-2 px-4">
                <Breadcrumb>
                  <BreadcrumbList>
                    <BreadcrumbItem className="hidden md:block">
                      <BreadcrumbLink href="#">{text.settings.title}</BreadcrumbLink>
                    </BreadcrumbItem>
                    <BreadcrumbSeparator className="hidden md:block" />
                    <BreadcrumbItem>
                      <BreadcrumbPage>{activeTitle}</BreadcrumbPage>
                    </BreadcrumbItem>
                  </BreadcrumbList>
                </Breadcrumb>
              </div>
            </header>
            <div className="flex flex-1 flex-col gap-1 overflow-y-auto p-4 pt-2">
              {activeCategory.settings.map((setting) => (
                <SettingRenderer
                  key={setting.id}
                  schema={setting}
                  categoryLabel={activeTitle}
                  value={values[setting.id]}
                  onChange={(next) => setValue(setting.id, next)}
                />
              ))}
            </div>
          </main>
        </SidebarProvider>
      </DialogContent>
    </Dialog>
  )
}
