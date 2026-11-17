import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

// cn merges conditional class names and resolves Tailwind conflicts, so a later utility always wins
// over the component default it overrides.
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
