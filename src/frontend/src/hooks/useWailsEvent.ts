import { useEffect } from 'react'
import { Events } from '@wailsio/runtime'

/** WailsEventHandler receives the payload the backend emitted, not the event wrapper. */
export type WailsEventHandler = (payload: unknown) => void

// useWailsEvent subscribes to a backend event and releases the subscription on unmount or when the
// handler identity changes. Pass a memoised handler so the subscription is not rebuilt on every render.
//
// The runtime hands the callback a WailsEvent, whose payload sits in `data`. Unwrapping it here means a handler reads
// what the backend emitted instead of the envelope around it - the same shape the publisher sent, which is what every
// handler in this app expects.
export function useWailsEvent(event: string, handler: WailsEventHandler): void {
  useEffect(() => {
    const cancel: unknown = Events.On(event, (delivered: { data?: unknown }) => {
      handler(delivered?.data)
    })
    return () => {
      if (typeof cancel === 'function') {
        ;(cancel as () => void)()
      }
    }
  }, [event, handler])
}
