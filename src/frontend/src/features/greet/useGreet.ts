import { useCallback, useRef, useState } from 'react'
import { greet } from '../../services/greet'

// toastDuration is how long the greeting stays visible, matching the stylesheet transition.
const toastDuration = 4000

export type GreetState = {
  name: string
  result: string
  visible: boolean
  setName: (name: string) => void
  submit: (name: string) => Promise<void>
}

// useGreet owns the greeting feature state: the input value, the last answer and the toast visibility.
export function useGreet(): GreetState {
  const [name, setName] = useState('')
  const [result, setResult] = useState('')
  const [visible, setVisible] = useState(false)
  const timer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined)

  const submit = useCallback(async (value: string) => {
    try {
      setResult(await greet(value))
      setVisible(true)
      clearTimeout(timer.current)
      timer.current = setTimeout(() => setVisible(false), toastDuration)
    } catch (error) {
      console.error(error)
    }
  }, [])

  return { name, result, visible, setName, submit }
}
