import { useEffect, useState } from 'react'

import { GraphCanvas } from '../graph/GraphCanvas'
import { defaultWorkspaceSnapshot, loadWorkspace } from './workspaceStorage'
import { createGraphPopoutChannel, isGraphPopoutMessage } from './graphPopoutChannel'

/**
 * Minimal topology-only window: open from Topology via "Graph in new window".
 * Stays in sync with the main workspace over BroadcastChannel (same origin).
 */
export default function GraphPopoutApp() {
  const initial = loadWorkspace() ?? defaultWorkspaceSnapshot()
  const [graphText, setGraphText] = useState(initial.graphText)

  useEffect(() => {
    const ch = createGraphPopoutChannel()
    if (!ch) {
      return
    }
    const onMessage = (ev: MessageEvent) => {
      if (isGraphPopoutMessage(ev.data)) {
        setGraphText(ev.data.text)
      }
    }
    ch.addEventListener('message', onMessage)
    return () => {
      ch.removeEventListener('message', onMessage)
      ch.close()
    }
  }, [])

  return (
    <div className="flex h-dvh min-h-dvh flex-col overflow-hidden bg-gray-950 font-sans text-gray-100">
      <header className="shrink-0 border-b border-gray-800 bg-gray-900/80 px-4 py-2.5 text-sm text-gray-400 backdrop-blur-sm">
        <span className="font-medium text-gray-300">Topology</span>
        <span className="text-gray-600"> · </span>
        <span>Synced from the main OmniGraph tab. Move this window to a second display if you like.</span>
      </header>
      <div className="flex min-h-0 flex-1 flex-col p-2">
        <GraphCanvas graphText={graphText} className="min-h-0 flex-1" />
      </div>
    </div>
  )
}
