import { useCallback, useEffect, useEffectEvent, useState } from 'react'

interface Result<T> {
  key: string // which request produced this result
  data?: T
  error?: unknown
}

/**
 * Runs `load` on mount and whenever `key` changes, cancelling the previous request.
 * While a new request runs, the previous data stays visible and `loading` is true.
 * `reload` repeats the request; `setData` replaces the data after a local change.
 */
export function useLoad<T>(load: (signal: AbortSignal) => Promise<T>, key: string | number = '') {
  const [version, setVersion] = useState(0)
  const requestKey = `${key}#${version}`
  const [result, setResult] = useState<Result<T>>({ key: '' })

  // Always calls the latest `load` without making the effect re-run when it changes.
  const run = useEffectEvent(load)

  useEffect(() => {
    const controller = new AbortController()
    run(controller.signal).then(
      (data) => setResult({ key: requestKey, data }),
      (error: unknown) => {
        if (controller.signal.aborted) return
        setResult((prev) => ({ key: requestKey, data: prev.data, error }))
      },
    )
    return () => controller.abort()
  }, [requestKey])

  const reload = useCallback(() => setVersion((v) => v + 1), [])
  const setData = useCallback((data: T) => setResult({ key: requestKey, data }), [requestKey])

  const settled = result.key === requestKey
  return {
    data: result.data,
    error: settled ? result.error : undefined,
    loading: !settled,
    reload,
    setData,
  }
}
