import { useCallback, useEffect, useMemo, useState } from 'react'
import { helpService, type DocsVersionDto } from '@/services/help'

export type DocsState = {
  /** versions lists every documentation set, ordered by the backend. */
  versions: DocsVersionDto[]
  /** version is the set the tree currently shows. */
  version: string
  /** selected is the chosen document, empty while the reader shows the index page. */
  selected: string
  /** content is the markdown of the selection, or of the index page when nothing is selected. */
  content: string
  /** error carries a load failure, empty when everything is fine. */
  error: string
  /** selectVersion switches the set and returns to its index page. */
  selectVersion: (next: string) => void
  /** selectDocument loads one document of the active set. */
  selectDocument: (path: string) => void
  /** clear drops the selection, which brings the index page back. */
  clear: () => void
}

// useHelp owns the documentation index, the active version and the chosen document. The tree, the picker
// and the reader all read from here. Selection and index are one piece of state: an empty selection means
// the index page, so the reader never has to guess what to show.
export function useHelp(active: boolean): DocsState {
  const [versions, setVersions] = useState<DocsVersionDto[]>([])
  const [version, setVersion] = useState('')
  const [selected, setSelected] = useState('')
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

  // The index path is a plain string, so the effect below reruns on a version switch rather than on every
  // render: the found version object is rebuilt each time.
  const index = useMemo(
    () => versions.find((entry) => entry.version === version)?.index ?? '',
    [versions, version],
  )

  useEffect(() => {
    const target = selected !== '' ? selected : index
    if (version === '' || target === '') {
      return
    }
    let cancelled = false
    setError('')
    helpService
      .read(version, target)
      .then((loaded) => {
        if (!cancelled) {
          setContent(loaded)
        }
      })
      .catch((cause: unknown) => {
        if (!cancelled) {
          setError(cause instanceof Error ? cause.message : 'document could not be read')
        }
      })
    return () => {
      cancelled = true
    }
  }, [version, index, selected])

  const selectVersion = useCallback((next: string) => {
    setVersion(next)
    setSelected('')
    setContent('')
  }, [])

  const selectDocument = useCallback((path: string) => {
    setSelected(path)
  }, [])

  const clear = useCallback(() => {
    setSelected('')
  }, [])

  return { versions, version, selected, content, error, selectVersion, selectDocument, clear }
}
