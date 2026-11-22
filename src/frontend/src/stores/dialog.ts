import { create } from 'zustand'

// DialogStore keeps global overlay visibility plus the document selected in the help browser. The
// selection lives here so the sidebar command palette and the help dialog stay in sync.
type DialogStore = {
  settingsOpen: boolean
  helpOpen: boolean
  commandOpen: boolean
  consoleOpen: boolean
  accountOpen: boolean
  helpDocument: string
  openSettings: () => void
  closeSettings: () => void
  openHelp: () => void
  closeHelp: () => void
  openCommand: () => void
  closeCommand: () => void
  openConsole: () => void
  closeConsole: () => void
  openAccount: () => void
  closeAccount: () => void
  selectHelpDocument: (document: string) => void
}

export const useDialogStore = create<DialogStore>((set) => ({
  settingsOpen: false,
  helpOpen: false,
  commandOpen: false,
  consoleOpen: false,
  accountOpen: false,
  helpDocument: '',
  openSettings: () => set({ settingsOpen: true }),
  closeSettings: () => set({ settingsOpen: false }),
  openHelp: () => set({ helpOpen: true }),
  closeHelp: () => set({ helpOpen: false }),
  openCommand: () => set({ commandOpen: true }),
  closeCommand: () => set({ commandOpen: false }),
  openConsole: () => set({ consoleOpen: true }),
  closeConsole: () => set({ consoleOpen: false }),
  openAccount: () => set({ accountOpen: true }),
  closeAccount: () => set({ accountOpen: false }),
  selectHelpDocument: (document) => set({ helpDocument: document }),
}))
