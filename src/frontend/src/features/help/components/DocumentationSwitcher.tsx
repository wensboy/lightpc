import { CheckIcon, ChevronsUpDownIcon, GalleryVerticalEndIcon } from 'lucide-react'
import { useState } from 'react'
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

type DocumentationSwitcherProps = {
  versions: readonly string[]
}

// DocumentationSwitcher mirrors the sidebar-02 header: the brand block is a dropdown trigger and the
// list offers every documentation version with the active one checked.
export function DocumentationSwitcher({ versions }: DocumentationSwitcherProps) {
  const text = useText()
  const [activeVersion, setActiveVersion] = useState(versions[0])

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
              <span className="text-xs">{activeVersion}</span>
            </div>
            <ChevronsUpDownIcon className="ml-auto size-4" />
          </DropdownMenuTrigger>
          <DropdownMenuContent align="start">
            {versions.map((version) => (
              <DropdownMenuItem key={version} onClick={() => setActiveVersion(version)}>
                {version}
                {version === activeVersion ? <CheckIcon className="ml-auto size-4" /> : null}
              </DropdownMenuItem>
            ))}
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  )
}
