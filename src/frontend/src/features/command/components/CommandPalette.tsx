import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command'
import { useText } from '@/locales'
import { useDialogStore } from '@/stores/dialog'

// CommandPalette is the global command surface opened from the sidebar. Selecting a document records it
// in the shared store, so the help dialog opens on that page next time.
export function CommandPalette() {
  const text = useText()
  const commandOpen = useDialogStore((state) => state.commandOpen)
  const closeCommand = useDialogStore((state) => state.closeCommand)
  const selectHelpDocument = useDialogStore((state) => state.selectHelpDocument)
  const openHelp = useDialogStore((state) => state.openHelp)

  const sections = Object.values(text.help.sections)

  return (
    <CommandDialog
      open={commandOpen}
      onOpenChange={(open) => {
        if (!open) {
          closeCommand()
        }
      }}
      title={text.help.command.title}
      description={text.help.command.description}
    >
      <CommandInput placeholder={text.help.command.placeholder} />
      <CommandList>
        <CommandEmpty>{text.help.command.empty}</CommandEmpty>
        {sections.map((section) => (
          <CommandGroup key={section.title} heading={section.title}>
            {Object.values(section.items).map((item) => (
              <CommandItem
                key={item}
                value={item}
                onSelect={() => {
                  selectHelpDocument(item)
                  closeCommand()
                  openHelp()
                }}
              >
                {item}
              </CommandItem>
            ))}
          </CommandGroup>
        ))}
      </CommandList>
    </CommandDialog>
  )
}
