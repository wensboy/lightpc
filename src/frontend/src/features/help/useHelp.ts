import { useCallback, useEffect, useState } from 'react'
import { helpService, type DocsVersion } from '@/services/help'

export type DocsState = {
  /** versions lists every documentation set, ordered by the backend. */
  versions: DocsVersion[]
  /** version is the set the tree currently shows. */
  version: string
  /** content is the markdown of the selected document. */
  content: string
  /** error carries a load failure, empty when everything is fine. */
  error: string
  /** selectVersion switches the set and clears the document. */
  selectVersion: (next: string) => void
  /** selectDocument loads one document of the active set. */
  selectDocument: (path: string) => void
}

// useHelp owns the documentation index, the active version and the loaded document. The tree, the picker
// and the reader all read from here, so the three stay in step without extra state.
export function useHelp(active: boolean): DocsState {
  const [versions, setVersions] = useState<DocsVersion[]>([])
  const [version, setVersion] = useState('')
  const [content, setContent] = useState('')
  const [error, setError] = useState('')

  useEffect(() => {
    if (!active || versions.length > 0) {
      return
    }
    let cancelled = false
    helpService
      .payload()
      .then((loaded) => {
        if (cancelled) {
          return
        }
        setVersions(loaded)
        setVersion((current) => current || (loaded[0]?.version ?? ''))
      })
      .catch((cause: unknown) => {
        if (cancelled) {
          return
        }
        setError(cause instanceof Error ? cause.message : 'documentation could not be listed')
      })
    return () => {
      cancelled = true
    }
  }, [active, versions.length])

  const selectVersion = useCallback((next: string) => {
    setVersion(next)
    setContent('')
  }, [])

  const selectDocument = useCallback(
    (path: string) => {
      if (version === '') {
        return
      }
      setError('')
      helpService
        .read(version, path)
        .then(setContent)
        .catch((cause: unknown) => {
          setError(cause instanceof Error ? cause.message : 'document could not be read')
        })
    },
    [version],
  )

  return { versions, version, content, error, selectVersion, selectDocument }
}
