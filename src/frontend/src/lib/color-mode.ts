// The shadcn dark variant is class driven, so the system preference has to be mirrored onto the root
// element. Applying it here keeps every "dark:" utility (images included) working without a provider.
const darkQuery = '(prefers-color-scheme: dark)'

export function applySystemColorMode(): () => void {
  const media = window.matchMedia(darkQuery)

  const sync = () => {
    document.documentElement.classList.toggle('dark', media.matches)
  }

  sync()
  media.addEventListener('change', sync)
  return () => media.removeEventListener('change', sync)
}
