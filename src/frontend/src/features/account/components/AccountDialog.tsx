import { ShieldCheckIcon, UserRoundIcon } from 'lucide-react'
import { useState } from 'react'
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
import { ProfileSection } from '@/features/account/components/ProfileSection'
import { SecuritySection } from '@/features/account/components/SecuritySection'
import { useText } from '@/locales'
import { useDialogStore } from '@/stores/dialog'

// accountTabs mirrors the sidebar-13 navigation: one entry per section shown on the right.
const accountTabs = [
  { key: 'profile', icon: <UserRoundIcon /> },
  { key: 'security', icon: <ShieldCheckIcon /> },
] as const

type AccountTab = (typeof accountTabs)[number]['key']

// AccountDialog hosts the account sections. The left rail selects a tab, the breadcrumb follows the
// selection and the body is reserved for the section bodies added in the next steps.
export function AccountDialog() {
  const text = useText()
  const accountOpen = useDialogStore((state) => state.accountOpen)
  const closeAccount = useDialogStore((state) => state.closeAccount)
  const [activeTab, setActiveTab] = useState<AccountTab>('profile')

  return (
    <Dialog
      open={accountOpen}
      onOpenChange={(open) => {
        if (!open) {
          closeAccount()
        }
      }}
    >
      <DialogContent className="overflow-hidden p-0 md:max-h-[500px] md:max-w-[700px] lg:max-w-[800px]">
        <DialogTitle className="sr-only">{text.account.title}</DialogTitle>
        <DialogDescription className="sr-only">{text.account.description}</DialogDescription>
        <SidebarProvider className="items-start">
          <Sidebar
            collapsible="none"
            className="hidden h-[480px] w-[250px] border-r border-sidebar-border md:flex"
          >
            <SidebarContent>
              <SidebarGroup>
                <SidebarGroupContent>
                  <SidebarMenu className="gap-1">
                    {accountTabs.map((tab) => (
                      <SidebarMenuItem key={tab.key}>
                        <SidebarMenuButton
                          isActive={tab.key === activeTab}
                          onClick={() => setActiveTab(tab.key)}
                        >
                          {tab.icon}
                          <span>{text.account.tabs[tab.key]}</span>
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
                      <BreadcrumbLink href="#">{text.account.title}</BreadcrumbLink>
                    </BreadcrumbItem>
                    <BreadcrumbSeparator className="hidden md:block" />
                    <BreadcrumbItem>
                      <BreadcrumbPage>{text.account.tabs[activeTab]}</BreadcrumbPage>
                    </BreadcrumbItem>
                  </BreadcrumbList>
                </Breadcrumb>
              </div>
            </header>
            <div className="flex flex-1 flex-col gap-4 overflow-y-auto p-4 pt-2">
              {activeTab === 'profile' ? <ProfileSection /> : null}
              {activeTab === 'security' ? <SecuritySection /> : null}
            </div>
          </main>
        </SidebarProvider>
      </DialogContent>
    </Dialog>
  )
}
