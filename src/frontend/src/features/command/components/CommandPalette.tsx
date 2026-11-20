import { CircleHelpIcon, BellIcon, SettingsIcon, UserRoundIcon } from 'lucide-react'
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command'
import { useLogout } from '@/features/auth/useLogout'
import { useNotificationStore } from '@/features/notifications/store/useNotificationStore'
import { useText } from '@/locales'
import { useDialogStore } from '@/stores/dialog'

// CommandPalette lists the actions that are not reachable from a page, which is what the command entry
// means: it changes the application state instead of navigating. Titles come from the dashboard dictionary
// because they name the same entries the sidebar shows.
export function CommandPalette() {
  const text = useText()
  const commandOpen = useDialogStore((state) => state.commandOpen)
  const closeCommand = useDialogStore((state) => state.closeCommand)
  const openSettings = useDialogStore((state) => state.openSettings)
  const openHelp = useDialogStore((state) => state.openHelp)
  const openAccount = useDialogStore((state) => state.openAccount)
  const openNotifications = useNotificationStore((state) => state.openSheet)
  const logout = useLogout()

  const close = () => closeCommand()

  return (
    <CommandDialog
      open={commandOpen}
      onOpenChange={(next) => {
        if (!next) {
          close()
        }
      }}
      title={text.help.command.title}
      description={text.help.command.description}
    >
      <CommandInput placeholder={text.help.command.placeholder} />
      <CommandList>
        <CommandEmpty>{text.help.command.empty}</CommandEmpty>
        <CommandGroup heading={text.dashboard.nav.command}>
          <CommandItem
            onSelect={() => {
              close()
              openSettings()
            }}
          >
            <SettingsIcon />
            {text.dashboard.nav.settings}
          </CommandItem>
          <CommandItem
            onSelect={() => {
              close()
              openHelp()
            }}
          >
            <CircleHelpIcon />
            {text.dashboard.nav.help}
          </CommandItem>
          <CommandItem
            onSelect={() => {
              close()
              openNotifications()
            }}
          >
            <BellIcon />
            {text.dashboard.user.notifications}
          </CommandItem>
          <CommandItem
            onSelect={() => {
              close()
              openAccount()
            }}
          >
            <UserRoundIcon />
            {text.dashboard.user.account}
          </CommandItem>
          <CommandItem
            onSelect={() => {
              close()
              void logout()
            }}
          >
            {text.dashboard.user.logout}
          </CommandItem>
        </CommandGroup>
      </CommandList>
    </CommandDialog>
  )
}
