import { BlocksIcon, CodeIcon, SettingsIcon, SlidersHorizontalIcon } from 'lucide-react'
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
import { useDialogStore } from '@/stores/dialog'

// groupIcons maps a document section to its glyph; an unknown section falls back to the generic icon.
const groupIcons: Record<string, ReactNode> = {
  appearence: <SlidersHorizontalIcon />,
  general: <SettingsIcon />,
  account: <CodeIcon />,
  advanced: <BlocksIcon />,
}

// SettingsDialog renders the settings document: the left rail lists the top level sections and the body
// renders one card per field, driven by the schema that ships with the document.
export function SettingsDialog() {
  const text = useText()
  const settingsOpen = useDialogStore((state) => state.settingsOpen)
  const closeSettings = useDialogStore((state) => state.closeSettings)
  // Opening the dialog triggers a reload, so an edit made to the user layer outside the UI shows up
  // immediately instead of requiring a restart.
  const { groups, error, setValue } = useSettings(settingsOpen)
  const [activeKey, setActiveKey] = useState(() => groups[0]?.key ?? '')

  const activeGroup = groups.find((group) => group.key === activeKey) ?? groups[0]

  return (
    <Dialog
      open={settingsOpen}
      onOpenChange={(open) => {
        if (!open) {
          closeSettings()
        }
      }}
    >
      <DialogContent className="overflow-hidden p-0 md:h-[500px] md:max-w-[700px] lg:max-w-[800px]">
        <DialogTitle className="sr-only">{text.settings.title}</DialogTitle>
        <DialogDescription className="sr-only">{text.settings.description}</DialogDescription>
        <SidebarProvider className="h-full min-h-0 items-stretch">
          <Sidebar collapsible="none" className="hidden border-r border-sidebar-border md:flex">
            <SidebarContent>
              <SidebarGroup>
                <SidebarGroupContent>
                  <SidebarMenu className="gap-1">
                    {groups.map((group) => (
                      <SidebarMenuItem key={group.key}>
                        <SidebarMenuButton
                          isActive={group.key === activeGroup?.key}
                          onClick={() => setActiveKey(group.key)}
                        >
                          {groupIcons[group.key] ?? <SettingsIcon />}
                          <span>{group.title}</span>
                        </SidebarMenuButton>
                      </SidebarMenuItem>
                    ))}
                  </SidebarMenu>
                </SidebarGroupContent>
              </SidebarGroup>
            </SidebarContent>
          </Sidebar>
          <main className="flex h-full flex-1 flex-col overflow-hidden">
            <header className="flex h-16 shrink-0 items-center gap-2">
              <div className="flex items-center gap-2 px-4">
                <Breadcrumb>
                  <BreadcrumbList>
                    <BreadcrumbItem className="hidden md:block">
                      <BreadcrumbLink href="#">{text.settings.title}</BreadcrumbLink>
                    </BreadcrumbItem>
                    <BreadcrumbSeparator className="hidden md:block" />
                    <BreadcrumbItem>
                      <BreadcrumbPage>{activeGroup?.title ?? text.settings.title}</BreadcrumbPage>
                    </BreadcrumbItem>
                  </BreadcrumbList>
                </Breadcrumb>
              </div>
            </header>
            <div className="flex flex-1 flex-col gap-1 overflow-y-auto p-4 pt-2">
              {activeGroup ? (
                activeGroup.fields.map((field) => (
                  <SettingRenderer
                    key={field.path}
                    field={field}
                    onChange={(next) => setValue(field.path, next)}
                  />
                ))
              ) : (
                <p className="p-4 text-xs text-muted-foreground">{error ?? text.settings.empty}</p>
              )}
            </div>
          </main>
        </SidebarProvider>
      </DialogContent>
    </Dialog>
  )
}
