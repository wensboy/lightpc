// The shadcn dark variant is class driven, so the chosen mode has to be mirrored onto the root element.
// Keeping it here means every "dark:" utility, images included, follows one source of truth.

export type ColorMode = 'light' | 'dark' | 'system'

const darkQuery = '(prefers-color-scheme: dark)'

let mode: ColorMode = 'system'
let watching = false

// sync mirrors the current mode onto the root element. System resolves through the media query.
function sync(): void {
  const dark = mode === 'dark' || (mode === 'system' && window.matchMedia(darkQuery).matches)
  document.documentElement.classList.toggle('dark', dark)
}

// setColorMode applies an explicit choice. A system choice keeps following the operating system, so the
// listener is installed once and stays registered for the lifetime of the page.
export function setColorMode(next: ColorMode): void {
  mode = next
  if (!watching) {
    watching = true
    window.matchMedia(darkQuery).addEventListener('change', sync)
  }
  sync()
}

// applySystemColorMode installs the system default before any setting has been loaded.
export function applySystemColorMode(): void {
  setColorMode('system')
}
