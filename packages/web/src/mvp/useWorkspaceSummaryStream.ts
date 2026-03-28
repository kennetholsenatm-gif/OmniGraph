import { useEffect, useRef, useState } from 'react'

import { omnigraphApiBase, type WorkspaceSummary } from './omnigraphApi'

/**
 * Subscribes to GET /api/v1/workspace/stream (SSE). Summary updates arrive only from
 * workspace_summary events — never optimistically.
 */
export function useWorkspaceSummaryStream(
  path: string,
  onSummary: (s: WorkspaceSummary) => void,
): { connected: boolean; error: string | null } {
  const [connected, setConnected] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const onSummaryRef = useRef(onSummary)
  onSummaryRef.current = onSummary

  useEffect(() => {
    const base = omnigraphApiBase()
    const q = encodeURIComponent(path.trim() || '.')
    const url = `${base}/api/v1/workspace/stream?path=${q}`
    const es = new EventSource(url)

    es.onopen = () => {
      setConnected(true)
      setError(null)
    }

    es.addEventListener('workspace_summary', ((ev: MessageEvent) => {
      try {
        const data = JSON.parse(ev.data as string) as WorkspaceSummary
        onSummaryRef.current(data)
        setError(null)
      } catch {
        setError('Invalid workspace_summary payload')
      }
    }) as EventListener)

    es.addEventListener('workspace_error', ((ev: MessageEvent) => {
      try {
        const raw = JSON.parse(ev.data as string)
        setError(typeof raw === 'string' ? raw : 'Workspace stream error')
      } catch {
        setError('Workspace stream error')
      }
    }) as EventListener)

    es.onerror = () => {
      setConnected(false)
      setError((prev) => prev ?? 'SSE disconnected (start omnigraph serve with --web-dist or set VITE_OMNIGRAPH_API)')
    }

    return () => {
      es.close()
      setConnected(false)
    }
  }, [path])

  return { connected, error }
}
