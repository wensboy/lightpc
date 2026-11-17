import { useEffect } from 'react'
import { Events } from '@wailsio/runtime'

export type WailsEventHandler = (payload: unknown) => void

// useWailsEvent subscribes to a backend event and releases the subscription on unmount or when the
// handler identity changes. Pass a memoised handler so the subscription is not rebuilt on every render.
export function useWailsEvent(event: string, handler: WailsEventHandler): void {
  useEffect(() => {
    const cancel: unknown = Events.On(event, handler)
    return () => {
      if (typeof cancel === 'function') {
        ;(cancel as () => void)()
      }
    }
  }, [event, handler])
}
