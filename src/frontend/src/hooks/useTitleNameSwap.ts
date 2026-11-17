import { useCallback, useRef } from 'react'

// useTitleNameSwap crossfades the heading word: the outgoing word fades out while the incoming one
// fades in over the same spot with the class names defined in style.css.
//
// The template shipped this as imperative DOM code inside the component; it is isolated here so the
// component stays declarative. A state driven rewrite lands together with the design system.
export function useTitleNameSwap() {
  const containerRef = useRef<HTMLSpanElement | null>(null)

  const swap = useCallback((text: string) => {
    const container = containerRef.current
    if (!container) {
      return
    }

    const current = container.querySelector('.title-name-text:not(.is-outgoing)')
    if (!current || current.textContent === text) {
      return
    }

    const incoming = document.createElement('span')
    incoming.className = 'title-name-text is-entering'
    incoming.textContent = text
    current.classList.add('is-outgoing')
    container.appendChild(incoming)

    // Force a reflow so the transitions run from their starting state.
    void incoming.offsetWidth
    incoming.classList.remove('is-entering')
    current.classList.add('is-leaving')
    current.addEventListener('transitionend', () => current.remove(), { once: true })
  }, [])

  return { containerRef, swap }
}
