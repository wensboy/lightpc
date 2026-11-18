import { Link } from '@tanstack/react-router'
import { GalleryVerticalEndIcon } from 'lucide-react'
import type { ReactNode } from 'react'
import { useText } from '@/locales'

// AuthLayout stacks two layers: the artwork covers the whole window, and the form floats above it on
// the left half inside a translucent card so the image stays visible behind the content.
export function AuthLayout({ children }: { children: ReactNode }) {
  const text = useText()

  return (
    <div className="relative grid min-h-svh">
      <img
        src="/login-bg-light.jpg"
        alt=""
        className="absolute inset-0 size-full object-cover dark:hidden"
      />
      <img
        src="/login-bg-dark.jpg"
        alt=""
        className="absolute inset-0 hidden size-full object-cover dark:block"
      />
      <div className="relative z-10 flex flex-col gap-4 p-6 md:p-10">
        <div className="flex justify-center gap-2 md:justify-start">
          <Link to="/" className="flex items-center gap-2 font-medium">
            <div className="flex size-6 items-center justify-center rounded-md bg-primary text-primary-foreground">
              <GalleryVerticalEndIcon className="size-4" />
            </div>
            {text.common.app.name}
          </Link>
        </div>
        <div className="flex flex-1 items-center justify-center lg:w-1/2">
          {/* A frosted layer: the artwork shows through the translucent surface, while the ring and the
              shadow keep the card outline visible against it. */}
          <div className="w-full max-w-xs rounded-xl bg-background/60 p-6 shadow-lg ring-1 ring-foreground/15 backdrop-blur-md">
            {children}
          </div>
        </div>
      </div>
    </div>
  )
}
