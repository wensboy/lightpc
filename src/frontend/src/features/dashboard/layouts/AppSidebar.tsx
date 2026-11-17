import {
  ChartBarIcon,
  CircleHelpIcon,
  FolderIcon,
  GalleryVerticalEndIcon,
  LayoutDashboardIcon,
  ListIcon,
  SearchIcon,
  Settings2Icon,
  UsersIcon,
} from 'lucide-react'
import { NavMain, type NavMainItem } from '@/components/nav-main'
import { NavSecondary } from '@/components/nav-secondary'
import { NavUser } from '@/components/nav-user'
import { TeamSwitcher } from '@/components/team-switcher'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
} from '@/components/ui/sidebar'

const teams = [
  {
    name: 'lightpc',
    logo: <GalleryVerticalEndIcon className="size-4" />,
    plan: 'Local workspace',
  },
  {
    name: 'lightpc dev',
    logo: <GalleryVerticalEndIcon className="size-4" />,
    plan: 'Development',
  },
]

const user = {
  name: 'lightpc user',
  email: 'user@lightpc.local',
  avatar: '/wails.png',
}

const navMain: NavMainItem[] = [
  { title: 'Dashboard', to: '/', icon: <LayoutDashboardIcon /> },
  { title: 'Lifecycle', to: '/lifecycle', icon: <ListIcon /> },
  { title: 'Analytics', to: '/analytics', icon: <ChartBarIcon /> },
  { title: 'Projects', to: '/projects', icon: <FolderIcon /> },
  { title: 'Team', to: '/team', icon: <UsersIcon /> },
]

const navSecondary = [
  { title: 'Settings', url: '#', icon: <Settings2Icon /> },
  { title: 'Get Help', url: '#', icon: <CircleHelpIcon /> },
  { title: 'Search', url: '#', icon: <SearchIcon /> },
]

// AppSidebar assembles the shell: team switcher on top, the primary and secondary navigation in the
// middle, and the user menu at the bottom.
export function AppSidebar() {
  return (
    <Sidebar collapsible="icon">
      <SidebarHeader>
        <TeamSwitcher teams={teams} />
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={navMain} />
        <NavSecondary items={navSecondary} className="mt-auto" />
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={user} />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
