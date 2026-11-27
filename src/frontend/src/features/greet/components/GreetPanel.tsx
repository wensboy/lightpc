import type { ChangeEvent } from 'react'
import { useGreet } from '@/features/greet/hooks/useGreet'

export type GreetPanelProps = {
  // onGreet receives the resolved name so the shell can animate its heading.
  onGreet: (name: string) => void
}

// GreetPanel owns the greeting form and its toast. It reports the submitted name upwards and never
// talks to the bindings itself.
export function GreetPanel({ onGreet }: GreetPanelProps) {
  const { name, result, visible, setName, submit } = useGreet()

  const handleChange = (event: ChangeEvent<HTMLInputElement>) => {
    setName(event.target.value)
  }

  const handleClick = () => {
    const value = name || 'anonymous'
    onGreet(value)
    void submit(value)
  }

  return (
    <>
      <div className="greet">
        <div className="input-box">
          <svg className="input-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
          <input aria-label="input" className="input" value={name} onChange={handleChange} type="text" placeholder="Your name" autoComplete="off"/>
          <button aria-label="greet-btn" className="btn" onClick={handleClick}>Greet
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><line x1="5" y1="12" x2="19" y2="12"/><polyline points="12 5 19 12 12 19"/></svg>
          </button>
        </div>
      </div>

      <div className={visible ? 'toast is-visible' : 'toast'} role="status" aria-live="polite">
        <span className="toast-label">From Go</span>
        <span aria-label="result" className="toast-msg">{result}</span>
      </div>
    </>
  )
}
