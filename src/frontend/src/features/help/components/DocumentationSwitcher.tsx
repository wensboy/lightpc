import { CheckIcon, ChevronsUpDownIcon, GalleryVerticalEndIcon } from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from '@/components/ui/sidebar'
import { useText } from '@/locales'
import type { DocsVersionDto } from '@/services/help'

type DocumentationSwitcherProps = {
  versions: readonly DocsVersionDto[]
  active: string
  onChange: (version: string) => void
}

// DocumentationSwitcher mirrors the sidebar-02 header: the brand block is a dropdown trigger and the list
// offers every documentation version the backend reported, with the active one checked.
export function DocumentationSwitcher({ versions, active, onChange }: DocumentationSwitcherProps) {
  const text = useText()

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger
            render={
              <SidebarMenuButton
                size="lg"
                className="data-open:bg-sidebar-accent data-open:text-sidebar-accent-foreground"
              />
            }
          >
            <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground">
              <GalleryVerticalEndIcon className="size-4" />
            </div>
            <div className="flex flex-col gap-0.5 leading-none">
              <span className="font-medium">{text.help.switcher.title}</span>
              <span className="text-xs text-muted-foreground">{active}</span>
            </div>
            <ChevronsUpDownIcon className="ml-auto size-4" />
          </DropdownMenuTrigger>
          <DropdownMenuContent align="start" className="w-56">
            {versions.map((version) => (
              <DropdownMenuItem key={version.version} onClick={() => onChange(version.version)}>
                <span className="truncate">{version.version}</span>
                {version.version === active ? <CheckIcon className="ml-auto size-4" /> : null}
              </DropdownMenuItem>
            ))}
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  )
}
